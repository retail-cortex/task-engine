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

package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rmcguinness/gemini_task_engine/pkg/api"
	"github.com/rmcguinness/gemini_task_engine/pkg/model"
	"github.com/rmcguinness/gemini_task_engine/pkg/service"
	"github.com/stretchr/testify/assert"
)

// Mock UserRepository implementation for API testing.
type mockUserRepository struct {
	users map[string]*model.User
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		users: make(map[string]*model.User),
	}
}

func (m *mockUserRepository) Create(ctx context.Context, u *model.User) error {
	if _, ok := m.users[u.ID]; ok {
		return fmt.Errorf("user already exists")
	}
	m.users[u.ID] = u
	return nil
}

func (m *mockUserRepository) FindByID(ctx context.Context, id string) (*model.User, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, fmt.Errorf("user not found")
	}
	return u, nil
}

func (m *mockUserRepository) FindByOAuth(ctx context.Context, provider, oauthID string) (*model.User, error) {
	return nil, nil
}

func (m *mockUserRepository) Update(ctx context.Context, u *model.User) error {
	m.users[u.ID] = u
	return nil
}

func (m *mockUserRepository) AddRole(ctx context.Context, userID, roleID string) error {
	return nil
}

func (m *mockUserRepository) List(ctx context.Context) ([]*model.User, error) {
	return nil, nil
}

func (m *mockUserRepository) ListRange(ctx context.Context, offset, limit int) ([]*model.User, error) {
	return nil, nil
}

func (m *mockUserRepository) Delete(ctx context.Context, id string) error {
	delete(m.users, id)
	return nil
}

func (m *mockUserRepository) ListActiveOnShiftUsers(ctx context.Context, siteID string) ([]*model.User, error) {
	return nil, nil
}

func (m *mockUserRepository) CreateRole(ctx context.Context, r *model.Role) error {
	return nil
}

func (m *mockUserRepository) FindRoleByID(ctx context.Context, id string) (*model.Role, error) {
	return nil, nil
}

func (m *mockUserRepository) UpdateRole(ctx context.Context, r *model.Role) error {
	return nil
}

func (m *mockUserRepository) DeleteRole(ctx context.Context, id string) error {
	return nil
}

func (m *mockUserRepository) ListRoles(ctx context.Context) ([]*model.Role, error) {
	return nil, nil
}

func (m *mockUserRepository) ListRolesRange(ctx context.Context, offset, limit int) ([]*model.Role, error) {
	return nil, nil
}

// Mock TranslationService implementation for API testing.
type mockTranslationService struct {
	TranslateTalkFunc   func(ctx context.Context, associateID, targetLang string, audioContent []byte) ([]byte, string, error)
	TranslateListenFunc func(ctx context.Context, associateID, customerGender, targetLang string, audioContent []byte) ([]byte, string, error)
	ListHDVoicesFunc    func(ctx context.Context) ([]*service.HDVoice, error)
}

func (m *mockTranslationService) TranslateTalk(ctx context.Context, associateID, targetLang string, audioContent []byte) ([]byte, string, error) {
	if m.TranslateTalkFunc != nil {
		return m.TranslateTalkFunc(ctx, associateID, targetLang, audioContent)
	}
	return []byte("mock-audio"), "Hello World", nil
}

func (m *mockTranslationService) TranslateListen(ctx context.Context, associateID, customerGender, targetLang string, audioContent []byte) ([]byte, string, error) {
	if m.TranslateListenFunc != nil {
		return m.TranslateListenFunc(ctx, associateID, customerGender, targetLang, audioContent)
	}
	return []byte("mock-audio-listen"), "Welcome to the store", nil
}

func (m *mockTranslationService) ListHDVoices(ctx context.Context) ([]*service.HDVoice, error) {
	if m.ListHDVoicesFunc != nil {
		return m.ListHDVoicesFunc(ctx)
	}
	return []*service.HDVoice{
		{Name: "en-US-Journey-F", LanguageCode: "en-US", Gender: "FEMALE", QualityClass: "Journey"},
		{Name: "es-ES-Neural2-B", LanguageCode: "es-ES", Gender: "MALE", QualityClass: "Neural2"},
	}, nil
}

func (m *mockTranslationService) GenerateVoiceCloningKey(ctx context.Context, consentAudio []byte, languageCode string) (string, error) {
	return "projects/mock-project/locations/us-central1/voiceCloningKeys/mock-key-12345", nil
}

func (m *mockTranslationService) SynthesizeText(ctx context.Context, text string, voiceName, languageCode string) ([]byte, error) {
	return []byte("mock-preview-audio"), nil
}

func TestVoiceTranslationHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userRepo := newMockUserRepository()
	transSvc := &mockTranslationService{}
	handler := api.NewVoiceTranslationHandler(userRepo, transSvc)

	// Pre-seed a user for testing
	userRepo.users["user-123"] = &model.User{
		ID:                    "user-123",
		Email:                 "test@example.com",
		Name:                  "Test User",
		VoiceGenderPreference: "FEMALE",
		VoiceNamePreference:   "en-US-Journey-F",
		ClonedVoiceKey:        "",
	}

	t.Run("GET /profile/:id returns profile settings", func(t *testing.T) {
		router := gin.New()
		router.GET("/api/v1/profile/:id", handler.GetProfile)

		req, _ := http.NewRequest(http.MethodGet, "/api/v1/profile/user-123", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp api.ProfileDTO
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "user-123", resp.ID)
		assert.Equal(t, "FEMALE", resp.VoiceGenderPreference)
		assert.Equal(t, "en-US-Journey-F", resp.VoiceNamePreference)
	})

	t.Run("POST /profile/:id creates new profile", func(t *testing.T) {
		router := gin.New()
		router.POST("/api/v1/profile/:id", handler.SaveProfile)

		newProfile := api.ProfileDTO{
			Email:                 "new@example.com",
			Name:                  "New User",
			VoiceGenderPreference: "MALE",
			VoiceNamePreference:   "en-US-Journey-D",
		}
		body, _ := json.Marshal(newProfile)

		req, _ := http.NewRequest(http.MethodPost, "/api/v1/profile/user-456", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var resp api.ProfileDTO
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "user-456", resp.ID)
		assert.Equal(t, "MALE", resp.VoiceGenderPreference)
		assert.Equal(t, "en-US-Journey-D", resp.VoiceNamePreference)
	})

	t.Run("PUT /profile/:id updates existing profile", func(t *testing.T) {
		router := gin.New()
		router.PUT("/api/v1/profile/:id", handler.UpdateProfile)

		updatedProfile := api.ProfileDTO{
			Name:                  "Updated User",
			VoiceGenderPreference: "MALE",
			VoiceNamePreference:   "en-US-Journey-D",
			ClonedVoiceKey:        "custom-key",
		}
		body, _ := json.Marshal(updatedProfile)

		req, _ := http.NewRequest(http.MethodPut, "/api/v1/profile/user-123", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp api.ProfileDTO
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "Updated User", resp.Name)
		assert.Equal(t, "MALE", resp.VoiceGenderPreference)
		assert.Equal(t, "custom-key", resp.ClonedVoiceKey)
	})

	t.Run("POST /profile/:id/voice/clone saves consent bytes and returns clone key", func(t *testing.T) {
		router := gin.New()
		router.POST("/api/v1/profile/:id/voice/clone", handler.CloneVoice)

		consentAudio := []byte("fake-consent-recording-pcm-bytes")
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/profile/user-123/voice/clone", bytes.NewBuffer(consentAudio))
		req.Header.Set("Content-Type", "audio/wav")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "success", resp["status"])
		assert.Contains(t, resp["cloned_voice_key"].(string), "mock-key-12345")

		// Verify database user state has consent recording bytes populated!
		user, err := userRepo.FindByID(context.Background(), "user-123")
		assert.NoError(t, err)
		assert.Equal(t, consentAudio, user.ConsentRecording)
		assert.Equal(t, resp["cloned_voice_key"].(string), user.ClonedVoiceKey)
	})

	t.Run("POST /translate/talk transcribes, translates and returns audio stream", func(t *testing.T) {
		router := gin.New()
		router.POST("/api/v1/translate/talk", handler.TranslateTalk)

		transSvc.TranslateTalkFunc = func(ctx context.Context, associateID, targetLang string, audioContent []byte) ([]byte, string, error) {
			assert.Equal(t, "user-123", associateID)
			assert.Equal(t, "es-ES", targetLang)
			assert.Equal(t, []byte("fake-input-audio"), audioContent)
			return []byte("fake-translated-audio-mp3"), "Hola Mundo", nil
		}

		req, _ := http.NewRequest(http.MethodPost, "/api/v1/translate/talk?associate_id=user-123&target_language=es-ES", bytes.NewBuffer([]byte("fake-input-audio")))
		req.Header.Set("Content-Type", "application/octet-stream")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "audio/mpeg", w.Header().Get("Content-Type"))
		decodedText, err := url.QueryUnescape(w.Header().Get("X-Translated-Text"))
		assert.NoError(t, err)
		assert.Equal(t, "Hola Mundo", decodedText)
		assert.Equal(t, []byte("fake-translated-audio-mp3"), w.Body.Bytes())
	})

	t.Run("POST /translate/listen translates customer voice to associate language", func(t *testing.T) {
		router := gin.New()
		router.POST("/api/v1/translate/listen", handler.TranslateListen)

		transSvc.TranslateListenFunc = func(ctx context.Context, associateID, customerGender, targetLang string, audioContent []byte) ([]byte, string, error) {
			assert.Equal(t, "user-123", associateID)
			assert.Equal(t, "MALE", customerGender)
			assert.Equal(t, "es-ES", targetLang)
			assert.Equal(t, []byte("fake-customer-audio"), audioContent)
			return []byte("fake-translated-customer-audio-mp3"), "Welcome my friend", nil
		}

		req, _ := http.NewRequest(http.MethodPost, "/api/v1/translate/listen?associate_id=user-123&gender=MALE&target_language=es-ES", bytes.NewBuffer([]byte("fake-customer-audio")))
		req.Header.Set("Content-Type", "application/octet-stream")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "audio/mpeg", w.Header().Get("Content-Type"))
		decodedText, err := url.QueryUnescape(w.Header().Get("X-Translated-Text"))
		assert.NoError(t, err)
		assert.Equal(t, "Welcome my friend", decodedText)
		assert.Equal(t, []byte("fake-translated-customer-audio-mp3"), w.Body.Bytes())
	})

	t.Run("GET /translate/voices returns available premium voices", func(t *testing.T) {
		router := gin.New()
		router.GET("/api/v1/translate/voices", handler.ListVoices)

		req, _ := http.NewRequest(http.MethodGet, "/api/v1/translate/voices", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp []service.HDVoice
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Len(t, resp, 2)
		assert.Equal(t, "en-US-Journey-F", resp[0].Name)
		assert.Equal(t, "Journey", resp[0].QualityClass)
	})
}
