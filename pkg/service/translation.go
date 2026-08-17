// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	speech "cloud.google.com/go/speech/apiv1"
	speechpb "cloud.google.com/go/speech/apiv1/speechpb"
	"cloud.google.com/go/translate"
	tts "cloud.google.com/go/texttospeech/apiv1"
	ttspb "cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
	"github.com/rmcguinness/gemini_task_engine/pkg/model"
	"github.com/rmcguinness/gemini_task_engine/pkg/persistence"
	"golang.org/x/oauth2/google"
	"golang.org/x/text/language"
)

// HDVoice represents a premium Google Cloud voice.
type HDVoice struct {
	Name         string   `json:"name"`
	LanguageCode string   `json:"language_code"`
	Gender       string   `json:"gender"`
	QualityClass string   `json:"quality_class"` // Journey, Neural2, Wavenet
}

// TranslationService manages voice transcription, translation, and synthesis.
type TranslationService interface {
	TranslateTalk(ctx context.Context, associateID, targetLang string, audioContent []byte) ([]byte, string, error)
	TranslateListen(ctx context.Context, associateID, customerGender, targetLang string, audioContent []byte) ([]byte, string, error)
	ListHDVoices(ctx context.Context) ([]*HDVoice, error)
	GenerateVoiceCloningKey(ctx context.Context, consentAudio []byte, languageCode string) (string, error)
	SynthesizeText(ctx context.Context, text string, voiceName, languageCode string) ([]byte, error)
}

type translationService struct {
	userRepo    persistence.UserRepository
	speechClient *speech.Client
	ttsClient    *tts.Client
	transClient  *translate.Client
}

// NewTranslationService creates a new TranslationService instance.
func NewTranslationService(userRepo persistence.UserRepository) (TranslationService, error) {
	ctx := context.Background()

	speechClient, err := speech.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create Speech-to-Text client: %w", err)
	}

	ttsClient, err := tts.NewClient(ctx)
	if err != nil {
		speechClient.Close()
		return nil, fmt.Errorf("failed to create Text-to-Speech client: %w", err)
	}

	transClient, err := translate.NewClient(ctx)
	if err != nil {
		speechClient.Close()
		ttsClient.Close()
		return nil, fmt.Errorf("failed to create Translation client: %w", err)
	}

	return &translationService{
		userRepo:     userRepo,
		speechClient: speechClient,
		ttsClient:    ttsClient,
		transClient:  transClient,
	}, nil
}

func (s *translationService) TranslateTalk(ctx context.Context, associateID, targetLang string, audioContent []byte) ([]byte, string, error) {
	// 1. Load associate profile
	user, err := s.userRepo.FindByID(ctx, associateID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to load associate profile: %w", err)
	}

	// Resolve associate's preferred language (source language for talk)
	sourceLang := "en-US"
	if user.PreferredLanguage != nil {
		sourceLang = string(user.PreferredLanguage.Code)
	}

	// 2. Perform Speech-to-Text (STT) on associate's voice
	transcription, err := s.transcribe(ctx, audioContent, sourceLang)
	if err != nil {
		return nil, "", fmt.Errorf("speech-to-text failed: %w", err)
	}

	// 3. Translate transcription to target language
	translatedText, err := s.translateText(ctx, transcription, sourceLang, targetLang)
	if err != nil {
		return nil, "", fmt.Errorf("translation failed: %w", err)
	}

	// 4. Synthesize translated text (TTS) into target language
	var voiceParams *ttspb.VoiceSelectionParams
	useClonedVoice := user.ClonedVoiceKey != ""

	if useClonedVoice {
		// Use Chirp 3 Instant Custom Voice Clone!
		voiceParams = &ttspb.VoiceSelectionParams{
			LanguageCode: targetLang,
			VoiceClone: &ttspb.VoiceCloneParams{
				VoiceCloningKey: user.ClonedVoiceKey,
			},
		}
	} else {
		// Use user's selected voice if compatible, or default premium HD voice
		voiceParams = s.resolveVoiceParams(targetLang, user)
	}

	synthesizedAudio, err := s.synthesize(ctx, translatedText, voiceParams)
	if err != nil && useClonedVoice {
		// Self-healing fallback: if the custom cloned voice model failed,
		// gracefully fall back to the user's selected voice preference or default premium HD voice!
		fallbackParams := s.resolveVoiceParams(targetLang, user)
		log.Printf("[TTS] WARNING: Custom cloned voice model '%s' failed or was not found (%v). Gracefully falling back to selected/default premium voice '%s'.", user.ClonedVoiceKey, err, fallbackParams.Name)
		synthesizedAudio, err = s.synthesize(ctx, translatedText, fallbackParams)
	}

	if err != nil {
		return nil, "", fmt.Errorf("text-to-speech failed: %w", err)
	}

	return synthesizedAudio, translatedText, nil
}

func (s *translationService) TranslateListen(ctx context.Context, associateID, customerGender, targetLang string, audioContent []byte) ([]byte, string, error) {
	// 1. Load associate profile
	user, err := s.userRepo.FindByID(ctx, associateID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to load associate profile: %w", err)
	}

	// Resolve associate's preferred language (target language for listen)
	associateLang := "en-US"
	if user.PreferredLanguage != nil {
		associateLang = string(user.PreferredLanguage.Code)
	}

	// 2. Perform Speech-to-Text (STT) on customer's voice (source language is targetLang)
	transcription, err := s.transcribe(ctx, audioContent, targetLang)
	if err != nil {
		return nil, "", fmt.Errorf("customer speech-to-text failed: %w", err)
	}

	// 3. Translate customer transcription to associate's preferred language
	translatedText, err := s.translateText(ctx, transcription, targetLang, associateLang)
	if err != nil {
		return nil, "", fmt.Errorf("translation failed: %w", err)
	}

	// 4. Synthesize translated text in associate's language using customer's selected gender voice
	voiceParams := s.resolveHDVoice(associateLang, customerGender)
	synthesizedAudio, err := s.synthesize(ctx, translatedText, voiceParams)
	if err != nil {
		return nil, "", fmt.Errorf("text-to-speech failed: %w", err)
	}

	return synthesizedAudio, translatedText, nil
}

func (s *translationService) ListHDVoices(ctx context.Context) ([]*HDVoice, error) {
	if s.ttsClient == nil {
		return nil, fmt.Errorf("ttsClient is not initialized")
	}
	req := &ttspb.ListVoicesRequest{}
	resp, err := s.ttsClient.ListVoices(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list voices: %w", err)
	}

	return mapHDVoices(resp.Voices), nil
}

func mapHDVoices(voices []*ttspb.Voice) []*HDVoice {
	var list []*HDVoice
	for _, v := range voices {
		name := v.Name
		quality := ""
		if strings.Contains(name, "-Chirp3-HD-") {
			quality = "Chirp 3 HD"
		} else if strings.Contains(name, "-Chirp-HD-") {
			quality = "Chirp HD"
		}

		if quality != "" {
			gender := "NEUTRAL"
			if v.SsmlGender == ttspb.SsmlVoiceGender_MALE {
				gender = "MALE"
			} else if v.SsmlGender == ttspb.SsmlVoiceGender_FEMALE {
				gender = "FEMALE"
			}

			// Capture the primary language code
			langCode := "en-US"
			if len(v.LanguageCodes) > 0 {
				langCode = v.LanguageCodes[0]
			}

			list = append(list, &HDVoice{
				Name:         name,
				LanguageCode: langCode,
				Gender:       gender,
				QualityClass: quality,
			})
		}
	}
	return list
}

// Private helper for speech-to-text transcription.
func (s *translationService) transcribe(ctx context.Context, audioBytes []byte, languageCode string) (string, error) {
	log.Printf("[STT] Transcribing audio content: received %d bytes (target language: %s)", len(audioBytes), languageCode)

	// Diagnostic check for absolute silence (all zeros) and amplitude analysis
	isSilent := true
	maxVal := int16(0)
	sumAbs := int64(0)
	sampleCount := 0

	for i := 0; i < len(audioBytes)-1; i += 2 {
		// Parse little-endian 16-bit signed PCM sample
		val := int16(audioBytes[i]) | (int16(audioBytes[i+1]) << 8)
		absVal := val
		if absVal < 0 {
			absVal = -absVal
		}
		if absVal > maxVal {
			maxVal = absVal
		}
		sumAbs += int64(absVal)
		sampleCount++

		if val != 0 {
			isSilent = false
		}
	}

	avgVal := float64(0)
	if sampleCount > 0 {
		avgVal = float64(sumAbs) / float64(sampleCount)
	}

	if isSilent {
		log.Printf("[STT] WARNING: The received audio payload is COMPLETELY SILENT (contains only zeros). Google Speech-to-Text cannot transcribe silence. Please check your emulator microphone/input configuration.")
	} else {
		log.Printf("[STT] Diagnostic: Audio contains non-zero bytes. Amplitude analysis: Max = %d, Average = %.2f (scale: 0 to 32767).", maxVal, avgVal)
		if maxVal < 300 {
			log.Printf("[STT] WARNING: The audio signal is EXTREMELY WEAK (Max Amplitude %d is basically background static/hiss). Real speech typically has peak amplitudes exceeding 5,000. Please ensure your microphone is not muted and you are speaking directly into it.", maxVal)
		}
	}

	req := &speechpb.RecognizeRequest{
		Config: &speechpb.RecognitionConfig{
			Encoding:        speechpb.RecognitionConfig_LINEAR16, // Default LINEAR16 PCM
			SampleRateHertz: 16000,                               // 16kHz default
			LanguageCode:    languageCode,
			UseEnhanced:     true,                                // Enable advanced noise-resilient models
		},
		Audio: &speechpb.RecognitionAudio{
			AudioSource: &speechpb.RecognitionAudio_Content{Content: audioBytes},
		},
	}

	if s.speechClient == nil {
		return "", fmt.Errorf("speechClient is not initialized")
	}

	resp, err := s.speechClient.Recognize(ctx, req)
	if err != nil {
		return "", err
	}

	if len(resp.Results) == 0 || len(resp.Results[0].Alternatives) == 0 {
		return "", fmt.Errorf("no transcription alternatives found")
	}

	return resp.Results[0].Alternatives[0].Transcript, nil
}

// Private helper for text translation.
func (s *translationService) translateText(ctx context.Context, text, sourceLang, targetLang string) (string, error) {
	// Parse language tags
	targetTag, err := language.Parse(targetLang)
	if err != nil {
		return "", fmt.Errorf("invalid target language code: %w", err)
	}

	sourceTag, err := language.Parse(sourceLang)
	if err != nil {
		return "", fmt.Errorf("invalid source language code: %w", err)
	}

	if s.transClient == nil {
		return "", fmt.Errorf("transClient is not initialized")
	}

	resp, err := s.transClient.Translate(ctx, []string{text}, targetTag, &translate.Options{
		Source: sourceTag,
		Format: translate.Text,
	})
	if err != nil {
		return "", err
	}

	if len(resp) == 0 {
		return "", fmt.Errorf("empty translation response")
	}

	return resp[0].Text, nil
}

// Private helper for text-to-speech synthesis.
func (s *translationService) synthesize(ctx context.Context, text string, voice *ttspb.VoiceSelectionParams) ([]byte, error) {
	req := &ttspb.SynthesizeSpeechRequest{
		Input: &ttspb.SynthesisInput{
			InputSource: &ttspb.SynthesisInput_Text{Text: text},
		},
		Voice: voice,
		AudioConfig: &ttspb.AudioConfig{
			AudioEncoding: ttspb.AudioEncoding_MP3, // Stream back MP3 bytes
		},
	}

	if s.ttsClient == nil {
		return nil, fmt.Errorf("ttsClient is not initialized")
	}

	resp, err := s.ttsClient.SynthesizeSpeech(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.AudioContent, nil
}

// Private helper to resolve a high-fidelity voice based on language and gender preference.
func (s *translationService) resolveHDVoice(langCode, gender string) *ttspb.VoiceSelectionParams {
	g := ttspb.SsmlVoiceGender_FEMALE
	if strings.ToUpper(gender) == "MALE" {
		g = ttspb.SsmlVoiceGender_MALE
	} else if strings.ToUpper(gender) == "NEUTRAL" {
		g = ttspb.SsmlVoiceGender_NEUTRAL
	}

	// Provide premium defaults for common languages (Journey/Neural2/Wavenet)
	voiceName := ""
	langClean := strings.ToLower(langCode)
	resolvedLang := langCode
	
	if strings.HasPrefix(langClean, "en") {
		if g == ttspb.SsmlVoiceGender_MALE {
			voiceName = "en-US-Studio-Q" // Ultra-realistic Studio Male
		} else {
			voiceName = "en-US-Studio-O" // Ultra-realistic Studio Female
		}
		resolvedLang = "en-US"
	} else if strings.HasPrefix(langClean, "es") {
		if g == ttspb.SsmlVoiceGender_MALE {
			voiceName = "es-ES-Studio-F" // Verified Male in Google Cloud catalog
		} else {
			voiceName = "es-ES-Studio-C" // Verified Female in Google Cloud catalog
		}
		resolvedLang = "es-ES"
	} else if strings.HasPrefix(langClean, "fr") {
		if g == ttspb.SsmlVoiceGender_MALE {
			voiceName = "fr-FR-Studio-D" // Ultra-realistic Studio Male
		} else {
			voiceName = "fr-FR-Studio-A" // Ultra-realistic Studio Female
		}
		resolvedLang = "fr-FR"
	}

	params := &ttspb.VoiceSelectionParams{
		LanguageCode: resolvedLang,
	}
	if voiceName != "" {
		params.Name = voiceName
		params.SsmlGender = ttspb.SsmlVoiceGender_SSML_VOICE_GENDER_UNSPECIFIED // Bypass strict catalog validation
	} else {
		params.SsmlGender = g
	}
	return params
}

// Private helper to resolve voice params: prefers the user's selected voice preference if compatible with target language
func (s *translationService) resolveVoiceParams(targetLang string, user *model.User) *ttspb.VoiceSelectionParams {
	if user.VoiceNamePreference != "" {
		parts := strings.Split(user.VoiceNamePreference, "-")
		if len(parts) >= 2 {
			voiceLang := parts[0] + "-" + parts[1] // e.g. "es-ES"
			
			// Check if the language prefixes match (e.g. "es" == "es")
			voicePrefix := strings.ToLower(parts[0])
			targetPrefix := strings.ToLower(strings.Split(targetLang, "-")[0])
			
			if voicePrefix == targetPrefix {
				log.Printf("[TTS] Using user selected voice preference: %s (lang: %s)", user.VoiceNamePreference, voiceLang)
				return &ttspb.VoiceSelectionParams{
					LanguageCode: voiceLang,
					Name:         user.VoiceNamePreference,
					SsmlGender:   ttspb.SsmlVoiceGender_SSML_VOICE_GENDER_UNSPECIFIED,
				}
			}
		}
	}

	// Default fallback to premium default HD voice
	return s.resolveHDVoice(targetLang, user.VoiceGenderPreference)
}

func resolveVoiceConsentScript(languageCode string) string {
	consentScript := "I am the owner of this voice and I consent to Google using this voice to create a synthetic voice model."
	cleanLang := strings.ToLower(languageCode)
	if strings.HasPrefix(cleanLang, "es") {
		consentScript = "Soy el propietario de esta voz y doy mi consentimiento para que Google la utilice para crear un modelo de voz sintética."
	} else if strings.HasPrefix(cleanLang, "fr") {
		consentScript = "Je suis le propriétaire de cette voix et j'autorise Google à utiliser cette voix pour créer un modèle de voix synthétique."
	} else if strings.HasPrefix(cleanLang, "de") {
		consentScript = "Ich bin der Eigentümer dieser Stimme und bin damit einverstanden, dass Google diese Stimme zur Erstellung eines synthetischen Stimmmodells verwendet."
	} else if strings.HasPrefix(cleanLang, "it") {
		consentScript = "Sono il proprietario di questa voce e acconsento che Google la utilizzi per creare un modelo di voce sintetica."
	} else if strings.HasPrefix(cleanLang, "ja") {
		consentScript = "私はこの音声の所有者であり、Googleがこの音声を使用して音声合成モデルを作成することを承認します。"
	} else if strings.HasPrefix(cleanLang, "ko") {
		consentScript = "나는 이 음성의 소유자이며 구글이 이 음성을 사용하여 음성 합성 모델을 생성할 것을 허용합니다。"
	}
	return consentScript
}

func (s *translationService) GenerateVoiceCloningKey(ctx context.Context, consentAudio []byte, languageCode string) (string, error) {
	log.Printf("[TTS] Requesting Chirp 3 Instant Custom Voice Cloning Key from Google Cloud...")

	// 1. Resolve GCP Project ID
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		projectID = os.Getenv("GCP_PROJECT")
	}
	if projectID == "" {
		projectID = "cs-poc-gvosjaln9q6gcudiayjqdzq" // fallback to active project
	}

	// 2. Create authorized HTTP client using Application Default Credentials (ADC)
	client, err := google.DefaultClient(ctx, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return "", fmt.Errorf("failed to create authorized client: %w", err)
	}

	// 3. Map consent script to the exact compliance wording based on language code
	consentScript := resolveVoiceConsentScript(languageCode)

	// 4. Encode audio content (PCM LINEAR16 16kHz Mono) to Base64
	base64Audio := base64.StdEncoding.EncodeToString(consentAudio)

	// 5. Build request payload
	reqBody := map[string]interface{}{
		"reference_audio": map[string]interface{}{
			"audio_config": map[string]interface{}{
				"audio_encoding": "LINEAR16",
			},
			"content": base64Audio,
		},
		"voice_talent_consent": map[string]interface{}{
			"audio_config": map[string]interface{}{
				"audio_encoding": "LINEAR16",
			},
			"content": base64Audio,
		},
		"consent_script": consentScript,
		"language_code":  languageCode,
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := "https://texttospeech.googleapis.com/v1beta1/voices:generateVoiceCloningKey"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create http request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("x-goog-user-project", projectID)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Google Cloud TTS API error (HTTP %d): %s", resp.StatusCode, string(respBytes))
	}

	var respJSON struct {
		VoiceCloningKey string `json:"voiceCloningKey"`
	}
	if err := json.Unmarshal(respBytes, &respJSON); err != nil {
		return "", fmt.Errorf("failed to parse response JSON: %w", err)
	}

	log.Printf("[TTS] Chirp 3 Instant Custom Voice Cloning Key generated successfully!")
	return respJSON.VoiceCloningKey, nil
}

func (s *translationService) SynthesizeText(ctx context.Context, text string, voiceName, languageCode string) ([]byte, error) {
	log.Printf("[TTS] Synthesizing preview text in voice %s (lang: %s)", voiceName, languageCode)

	// Force fully-qualified regional language code matching the voice model
	resolvedLang := languageCode
	langClean := strings.ToLower(languageCode)
	if strings.HasPrefix(langClean, "es") {
		resolvedLang = "es-ES"
	} else if strings.HasPrefix(langClean, "fr") {
		resolvedLang = "fr-FR"
	} else if strings.HasPrefix(langClean, "en") {
		resolvedLang = "en-US"
	} else if strings.HasPrefix(langClean, "de") {
		resolvedLang = "de-DE"
	} else if strings.HasPrefix(langClean, "it") {
		resolvedLang = "it-IT"
	}

	voiceParams := &ttspb.VoiceSelectionParams{
		LanguageCode: resolvedLang,
		Name:         voiceName,
		SsmlGender:   ttspb.SsmlVoiceGender_SSML_VOICE_GENDER_UNSPECIFIED,
	}

	return s.synthesize(ctx, text, voiceParams)
}
