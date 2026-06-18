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

package api

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rmcguinness/gemini_task_engine/pkg/model"
	"github.com/rmcguinness/gemini_task_engine/pkg/persistence"
	"github.com/rmcguinness/gemini_task_engine/pkg/service"
)

// ProfileDTO serializes the associate's language and voice configuration preferences.
type ProfileDTO struct {
	ID                    string  `json:"id"`
	Email                 string  `json:"email"`
	Name                  string  `json:"name"`
	PreferredLanguageID   *string `json:"preferred_language_id"`
	VoiceGenderPreference string  `json:"voice_gender_preference"`
	VoiceNamePreference   string  `json:"voice_name_preference"`
	ClonedVoiceKey        string  `json:"cloned_voice_key"`
}

// VoiceTranslationHandler handles HTTP routing and DTO conversions for translation and profiles.
type VoiceTranslationHandler struct {
	userRepo     persistence.UserRepository
	transService service.TranslationService
}

// NewVoiceTranslationHandler creates a new VoiceTranslationHandler instance.
func NewVoiceTranslationHandler(userRepo persistence.UserRepository, transService service.TranslationService) *VoiceTranslationHandler {
	return &VoiceTranslationHandler{
		userRepo:     userRepo,
		transService: transService,
	}
}

// GetProfile handles GET /api/v1/profile/{id}
func (h *VoiceTranslationHandler) GetProfile(c *gin.Context) {
	id := c.Param("id")
	user, err := h.userRepo.FindByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Associate profile not found"})
		return
	}
	c.JSON(http.StatusOK, toProfileDTO(user))
}

// SaveProfile handles POST /api/v1/profile/{id} (Create / Init Profile)
func (h *VoiceTranslationHandler) SaveProfile(c *gin.Context) {
	id := c.Param("id")
	var req ProfileDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	// Check if user already exists
	existing, err := h.userRepo.FindByID(ctx, id)
	if err == nil && existing != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Profile already exists, use PUT to update"})
		return
	}

	newUser := &model.User{
		ID:                    id,
		Email:                 req.Email,
		Name:                  req.Name,
		PreferredLanguageID:   req.PreferredLanguageID,
		VoiceGenderPreference: req.VoiceGenderPreference,
		VoiceNamePreference:   req.VoiceNamePreference,
		ClonedVoiceKey:        req.ClonedVoiceKey,
	}

	if newUser.VoiceGenderPreference == "" {
		newUser.VoiceGenderPreference = "FEMALE"
	}
	if newUser.VoiceNamePreference == "" {
		newUser.VoiceNamePreference = "en-US-Journey-F"
	}

	err = h.userRepo.Create(ctx, newUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create profile: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, toProfileDTO(newUser))
}

// UpdateProfile handles PUT /api/v1/profile/{id}
func (h *VoiceTranslationHandler) UpdateProfile(c *gin.Context) {
	id := c.Param("id")
	var req ProfileDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	user, err := h.userRepo.FindByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Associate profile not found"})
		return
	}

	// Update mutable profile fields
	user.Name = req.Name
	user.PreferredLanguageID = req.PreferredLanguageID
	user.VoiceGenderPreference = req.VoiceGenderPreference
	user.VoiceNamePreference = req.VoiceNamePreference
	user.ClonedVoiceKey = req.ClonedVoiceKey

	if user.VoiceGenderPreference == "" {
		user.VoiceGenderPreference = "FEMALE"
	}
	if user.VoiceNamePreference == "" {
		user.VoiceNamePreference = "en-US-Journey-F"
	}

	err = h.userRepo.Update(ctx, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, toProfileDTO(user))
}

// CloneVoice handles POST /api/v1/profile/{id}/voice/clone
func (h *VoiceTranslationHandler) CloneVoice(c *gin.Context) {
	id := c.Param("id")

	// Read audio content (handles either form-data 'audio' file or raw body stream)
	var audioBytes []byte
	file, err := c.FormFile("audio")
	if err == nil {
		f, err := file.Open()
		if err == nil {
			defer f.Close()
			buf := make([]byte, file.Size)
			_, err = f.Read(buf)
			if err == nil {
				audioBytes = buf
			}
		}
	}
	if len(audioBytes) == 0 {
		buf, err := c.GetRawData()
		if err == nil {
			audioBytes = buf
		}
	}

	if len(audioBytes) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Audio recording of consent/authorization phrase is required"})
		return
	}

	ctx := c.Request.Context()
	user, err := h.userRepo.FindByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Associate profile not found"})
		return
	}

	// Save raw consent audio bytes directly to GORM Bytea field
	user.ConsentRecording = audioBytes

	// Call the live Google Cloud Chirp 3 Instant Custom Voice API to generate a cloning key!
	cloningKey, err := h.transService.GenerateVoiceCloningKey(ctx, audioBytes, "en-US")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Google Cloud voice cloning failed: " + err.Error()})
		return
	}
	user.ClonedVoiceKey = cloningKey

	err = h.userRepo.Update(ctx, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile with voice clone: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":           "success",
		"cloned_voice_key": user.ClonedVoiceKey,
	})
}

// TranslateTalk handles POST /api/v1/translate/talk
func (h *VoiceTranslationHandler) TranslateTalk(c *gin.Context) {
	associateID := c.PostForm("associate_id")
	if associateID == "" {
		associateID = c.Query("associate_id")
	}

	targetLang := c.PostForm("target_language")
	if targetLang == "" {
		targetLang = c.Query("target_language")
	}

	if associateID == "" || targetLang == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "associate_id and target_language parameters are required"})
		return
	}

	// Read audio content
	var audioBytes []byte
	file, err := c.FormFile("audio")
	if err == nil {
		f, err := file.Open()
		if err == nil {
			defer f.Close()
			buf := make([]byte, file.Size)
			_, err = f.Read(buf)
			if err == nil {
				audioBytes = buf
			}
		}
	}
	if len(audioBytes) == 0 {
		buf, err := c.GetRawData()
		if err == nil {
			audioBytes = buf
		}
	}

	if len(audioBytes) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Audio content is required for transcription"})
		return
	}

	audioOut, translatedText, err := h.transService.TranslateTalk(c.Request.Context(), associateID, targetLang, audioBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Set translated transcription in header so client UI can display it
	c.Header("X-Translated-Text", url.QueryEscape(translatedText))

	// Stream synthesized MP3 bytes back to client
	c.Data(http.StatusOK, "audio/mpeg", audioOut)
}

// TranslateListen handles POST /api/v1/translate/listen
func (h *VoiceTranslationHandler) TranslateListen(c *gin.Context) {
	associateID := c.PostForm("associate_id")
	if associateID == "" {
		associateID = c.Query("associate_id")
	}

	customerGender := c.PostForm("gender")
	if customerGender == "" {
		customerGender = c.Query("gender")
	}

	targetLang := c.PostForm("target_language") // The language the customer is speaking
	if targetLang == "" {
		targetLang = c.Query("target_language")
	}

	if associateID == "" || customerGender == "" || targetLang == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "associate_id, gender, and target_language parameters are required"})
		return
	}

	// Read audio content
	var audioBytes []byte
	file, err := c.FormFile("audio")
	if err == nil {
		f, err := file.Open()
		if err == nil {
			defer f.Close()
			buf := make([]byte, file.Size)
			_, err = f.Read(buf)
			if err == nil {
				audioBytes = buf
			}
		}
	}
	if len(audioBytes) == 0 {
		buf, err := c.GetRawData()
		if err == nil {
			audioBytes = buf
		}
	}

	if len(audioBytes) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Audio content is required for transcription"})
		return
	}

	audioOut, translatedText, err := h.transService.TranslateListen(c.Request.Context(), associateID, customerGender, targetLang, audioBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Set translated transcription in header
	c.Header("X-Translated-Text", url.QueryEscape(translatedText))

	// Stream synthesized MP3 bytes back
	c.Data(http.StatusOK, "audio/mpeg", audioOut)
}

// ListVoices handles GET /api/v1/translate/voices
func (h *VoiceTranslationHandler) ListVoices(c *gin.Context) {
	voices, err := h.transService.ListHDVoices(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, voices)
}

// Helper mapper DTO converter.
func toProfileDTO(u *model.User) ProfileDTO {
	return ProfileDTO{
		ID:                    u.ID,
		Email:                 u.Email,
		Name:                  u.Name,
		PreferredLanguageID:   u.PreferredLanguageID,
		VoiceGenderPreference: u.VoiceGenderPreference,
		VoiceNamePreference:   u.VoiceNamePreference,
		ClonedVoiceKey:        u.ClonedVoiceKey,
	}
}

// PreviewVoice handles POST /api/v1/translate/preview
func (h *VoiceTranslationHandler) PreviewVoice(c *gin.Context) {
	var req struct {
		VoiceName    string `json:"voice_name"`
		LanguageCode string `json:"language_code"`
		Text         string `json:"text"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.VoiceName == "" || req.LanguageCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "voice_name and language_code are required"})
		return
	}

	// Default text if empty
	text := req.Text
	if text == "" {
		text = "Hello! This is a preview of my voice."
		cleanLang := strings.ToLower(req.LanguageCode)
		if strings.HasPrefix(cleanLang, "es") {
			text = "¡Hola! Esta es una vista previa de mi voz."
		} else if strings.HasPrefix(cleanLang, "fr") {
			text = "Bonjour ! Ceci est un aperçu de ma voix."
		} else if strings.HasPrefix(cleanLang, "de") {
			text = "Hallo! Dies ist eine Vorschau meiner Stimme."
		} else if strings.HasPrefix(cleanLang, "it") {
			text = "Ciao! Questa è un'anteprima della mia voce."
		} else if strings.HasPrefix(cleanLang, "ja") {
			text = "こんにちは！これは私の声のプレビューです。"
		}
	}

	audioBytes, err := h.transService.SynthesizeText(c.Request.Context(), text, req.VoiceName, req.LanguageCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "audio/mpeg", audioBytes)
}
