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
	"context"
	"os"
	"testing"

	"github.com/rmcguinness/gemini_task_engine/pkg/model"
	"github.com/rmcguinness/gemini_task_engine/pkg/persistence"
	"github.com/stretchr/testify/assert"
	ttspb "cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
)

func TestTranslationService_ResolveHDVoice(t *testing.T) {
	s := &translationService{}

	tests := []struct {
		name         string
		langCode     string
		gender       string
		expectedName string
		expectedLang string
	}{
		{"English US Male", "en-US", "MALE", "en-US-Studio-Q", "en-US"},
		{"English US Female", "en-US", "FEMALE", "en-US-Studio-O", "en-US"},
		{"Spanish Studio Male", "es-ES", "MALE", "es-ES-Studio-F", "es-ES"},
		{"Spanish Studio Female", "es-ES", "FEMALE", "es-ES-Studio-C", "es-ES"},
		{"French Studio Male", "fr-FR", "MALE", "fr-FR-Studio-D", "fr-FR"},
		{"French Studio Female", "fr-FR", "FEMALE", "fr-FR-Studio-A", "fr-FR"},
		{"Other lang neutral", "de-DE", "NEUTRAL", "", "de-DE"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			params := s.resolveHDVoice(tc.langCode, tc.gender)
			assert.Equal(t, tc.expectedLang, params.LanguageCode)
			if tc.expectedName != "" {
				assert.Equal(t, tc.expectedName, params.Name)
				assert.Equal(t, ttspb.SsmlVoiceGender_SSML_VOICE_GENDER_UNSPECIFIED, params.SsmlGender)
			}
		})
	}
}

func TestTranslationService_ResolveVoiceParams(t *testing.T) {
	s := &translationService{}

	t.Run("user custom voice preference matches target language", func(t *testing.T) {
		u := &model.User{
			VoiceNamePreference: "es-ES-Studio-F",
		}
		params := s.resolveVoiceParams("es-MX", u)
		assert.Equal(t, "es-ES-Studio-F", params.Name)
		assert.Equal(t, "es-ES", params.LanguageCode)
	})

	t.Run("user custom voice preference mismatch language fallback", func(t *testing.T) {
		u := &model.User{
			VoiceNamePreference:   "en-US-Studio-O",
			VoiceGenderPreference: "FEMALE",
		}
		params := s.resolveVoiceParams("fr-FR", u)
		assert.Equal(t, "fr-FR-Studio-A", params.Name)
		assert.Equal(t, "fr-FR", params.LanguageCode)
	})
}

func TestTranslationService_GenerateVoiceCloningKey_ConsentScript(t *testing.T) {
	s := &translationService{}
	os.Setenv("GOOGLE_CLOUD_PROJECT", "test-proj")

	// Verify different consent script mappings by triggering GenerateVoiceCloningKey
	_, err := s.GenerateVoiceCloningKey(context.Background(), []byte("test"), "es-ES")
	assert.Error(t, err)

	_, err = s.GenerateVoiceCloningKey(context.Background(), []byte("test"), "fr-FR")
	assert.Error(t, err)

	_, err = s.GenerateVoiceCloningKey(context.Background(), []byte("test"), "de-DE")
	assert.Error(t, err)

	_, err = s.GenerateVoiceCloningKey(context.Background(), []byte("test"), "it-IT")
	assert.Error(t, err)

	_, err = s.GenerateVoiceCloningKey(context.Background(), []byte("test"), "ja-JP")
	assert.Error(t, err)

	_, err = s.GenerateVoiceCloningKey(context.Background(), []byte("test"), "ko-KR")
	assert.Error(t, err)

	assert.Contains(t, resolveVoiceConsentScript("en-US"), "I am the owner")
	assert.Contains(t, resolveVoiceConsentScript("es-ES"), "Soy el propietario")
	assert.Contains(t, resolveVoiceConsentScript("fr-FR"), "Je suis le propriétaire")
	assert.Contains(t, resolveVoiceConsentScript("de-DE"), "Ich bin der Eigentümer")
	assert.Contains(t, resolveVoiceConsentScript("it-IT"), "Sono il proprietario")
	assert.Contains(t, resolveVoiceConsentScript("ja-JP"), "私はこの音声の所有者")
	assert.Contains(t, resolveVoiceConsentScript("ko-KR"), "나는 이 음성의 소유자")

	hdVoices := mapHDVoices([]*ttspb.Voice{
		{
			Name:          "en-US-Chirp3-HD-A",
			LanguageCodes: []string{"en-US"},
			SsmlGender:    ttspb.SsmlVoiceGender_MALE,
		},
		{
			Name:          "es-ES-Chirp-HD-B",
			LanguageCodes: []string{"es-ES"},
			SsmlGender:    ttspb.SsmlVoiceGender_FEMALE,
		},
		{
			Name:          "fr-FR-Standard-A",
			LanguageCodes: []string{"fr-FR"},
			SsmlGender:    ttspb.SsmlVoiceGender_NEUTRAL,
		},
	})
	assert.Len(t, hdVoices, 2)
	assert.Equal(t, "Chirp 3 HD", hdVoices[0].QualityClass)
	assert.Equal(t, "Chirp HD", hdVoices[1].QualityClass)
}

type mockUserRepoForTranslation struct {
	persistence.UserRepository
	user *model.User
	err  error
}

func (m *mockUserRepoForTranslation) FindByID(ctx context.Context, id string) (*model.User, error) {
	return m.user, m.err
}

func TestTranslationService_TalkAndListen_ValidationAndAudioCheck(t *testing.T) {
	ctx := context.Background()

	t.Run("user not found returns error", func(t *testing.T) {
		repo := &mockUserRepoForTranslation{err: assert.AnError}
		svc := &translationService{userRepo: repo}

		_, _, err := svc.TranslateTalk(ctx, "u1", "es-ES", []byte{0, 0})
		assert.Error(t, err)

		_, _, err = svc.TranslateListen(ctx, "u1", "MALE", "es-ES", []byte{0, 0})
		assert.Error(t, err)
	})

	t.Run("silent PCM audio diagnostic and nil speech client returns error", func(t *testing.T) {
		lang := model.Language{Code: "en-US"}
		user := &model.User{
			ID:                "u1",
			PreferredLanguage: &lang,
			ClonedVoiceKey:    "custom-key",
		}
		repo := &mockUserRepoForTranslation{user: user}
		svc := &translationService{userRepo: repo}

		// Test silent audio bytes
		silentAudio := []byte{0, 0, 0, 0}
		_, _, err := svc.TranslateTalk(ctx, "u1", "es-ES", silentAudio)
		assert.Error(t, err)

		// Test weak amplitude non-silent audio bytes
		weakAudio := []byte{10, 0, 20, 0}
		_, _, err = svc.TranslateListen(ctx, "u1", "FEMALE", "es-ES", weakAudio)
		assert.Error(t, err)
	})

	t.Run("ListHDVoices and SynthesizeText return error on nil ttsClient", func(t *testing.T) {
		svc := &translationService{}
		_, err := svc.ListHDVoices(ctx)
		assert.Error(t, err)

		_, err = svc.SynthesizeText(ctx, "test", "en-US-Studio-O", "en")
		assert.Error(t, err)
		_, err = svc.SynthesizeText(ctx, "test", "es-ES-Studio-F", "es")
		assert.Error(t, err)
		_, err = svc.SynthesizeText(ctx, "test", "fr-FR-Studio-A", "fr")
		assert.Error(t, err)
		_, err = svc.SynthesizeText(ctx, "test", "de-DE-Studio-B", "de")
		assert.Error(t, err)
		_, err = svc.SynthesizeText(ctx, "test", "it-IT-Studio-C", "it")
		assert.Error(t, err)
		_, err = svc.SynthesizeText(ctx, "test", "zh-CN-Studio-D", "zh")
		assert.Error(t, err)
	})

	t.Run("NewTranslationService returns service", func(t *testing.T) {
		svc, _ := NewTranslationService(nil)
		assert.NotNil(t, svc)
	})

	t.Run("TranslateTalk and TranslateListen with found user enter transcribe and translateText", func(t *testing.T) {
		lang := model.Language{Code: "en-US"}
		user := &model.User{
			ID:                "u1",
			PreferredLanguage: &lang,
		}
		repo := &mockUserRepoForTranslation{user: user}
		svc := &translationService{userRepo: repo}

		// Audio amplitude > 300 triggers non-silent branch
		normalAudio := make([]byte, 100)
		for i := 0; i < len(normalAudio); i++ {
			normalAudio[i] = 127
		}
		_, _, err := svc.TranslateTalk(ctx, "u1", "es-ES", normalAudio)
		assert.Error(t, err) // Errors at s.speechClient == nil

		_, _, err = svc.TranslateListen(ctx, "u1", "MALE", "es-ES", normalAudio)
		assert.Error(t, err)

		_, err = svc.translateText(ctx, "hello", "invalid!!", "es-ES")
		assert.Error(t, err)
		_, err = svc.translateText(ctx, "hello", "en-US", "invalid!!")
		assert.Error(t, err)
		_, err = svc.translateText(ctx, "hello", "en-US", "es-ES")
		assert.Error(t, err)

		_, err = svc.synthesize(ctx, "hello", nil)
		assert.Error(t, err)
	})
}
