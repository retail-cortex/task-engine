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
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rmcguinness/gemini_task_engine/pkg/model"
	"github.com/rmcguinness/gemini_task_engine/pkg/service"
	"github.com/stretchr/testify/assert"
)

type mockTransForAPI struct {
	service.TranslationService
}

func (m *mockTransForAPI) SynthesizeText(ctx context.Context, text, voiceName, langCode string) ([]byte, error) {
	return []byte("mp3data"), nil
}
func (m *mockTransForAPI) ListHDVoices(ctx context.Context) ([]*service.HDVoice, error) {
	if ctx.Value("err") != nil {
		return nil, errors.New("err")
	}
	return []*service.HDVoice{{Name: "en-US-Journey-F"}}, nil
}
func (m *mockTransForAPI) TranslateTalk(ctx context.Context, assocID, targetLang string, srcAudio []byte) ([]byte, string, error) {
	if assocID == "err" {
		return nil, "", errors.New("err")
	}
	return []byte("mp3"), "hello", nil
}
func (m *mockTransForAPI) TranslateListen(ctx context.Context, assocID, custGender, targetLang string, targetAudio []byte) ([]byte, string, error) {
	if assocID == "err" {
		return nil, "", errors.New("err")
	}
	return []byte("mp3"), "hello", nil
}
func (m *mockTransForAPI) GenerateVoiceCloningKey(ctx context.Context, consentAudio []byte, languageCode string) (string, error) {
	if string(consentAudio) == "err" {
		return "", errors.New("err")
	}
	return "key-123", nil
}

type mockUserRepoForAPI struct{}

func (m *mockUserRepoForAPI) Create(ctx context.Context, u *model.User) error { return nil }
func (m *mockUserRepoForAPI) FindByID(ctx context.Context, id string) (*model.User, error) {
	if id == "new-u" || id == "err" {
		return nil, errors.New("not found")
	}
	return &model.User{ID: id, Name: "Test User"}, nil
}
func (m *mockUserRepoForAPI) FindByOAuth(ctx context.Context, p, o string) (*model.User, error) {
	return nil, nil
}
func (m *mockUserRepoForAPI) Update(ctx context.Context, u *model.User) error        { return nil }
func (m *mockUserRepoForAPI) AddRole(ctx context.Context, userID, roleID string) error { return nil }
func (m *mockUserRepoForAPI) List(ctx context.Context) ([]*model.User, error) {
	return []*model.User{{ID: "u1"}}, nil
}
func (m *mockUserRepoForAPI) ListRange(ctx context.Context, o, l int) ([]*model.User, error) {
	return []*model.User{{ID: "u1"}}, nil
}
func (m *mockUserRepoForAPI) Delete(ctx context.Context, id string) error { return nil }
func (m *mockUserRepoForAPI) ListActiveOnShiftUsers(ctx context.Context, siteID string) ([]*model.User, error) {
	return []*model.User{{ID: "u1"}}, nil
}
func (m *mockUserRepoForAPI) CreateRole(ctx context.Context, r *model.Role) error             { return nil }
func (m *mockUserRepoForAPI) FindRoleByID(ctx context.Context, id string) (*model.Role, error) {
	return &model.Role{ID: id}, nil
}
func (m *mockUserRepoForAPI) UpdateRole(ctx context.Context, r *model.Role) error { return nil }
func (m *mockUserRepoForAPI) DeleteRole(ctx context.Context, id string) error     { return nil }
func (m *mockUserRepoForAPI) ListRoles(ctx context.Context) ([]*model.Role, error) {
	return []*model.Role{{ID: "r1"}}, nil
}
func (m *mockUserRepoForAPI) ListRolesRange(ctx context.Context, o, l int) ([]*model.Role, error) {
	return []*model.Role{{ID: "r1"}}, nil
}

type mockChatShiftForAPI struct {
	service.ShiftService
}

func (m *mockChatShiftForAPI) InitializeShift(ctx context.Context, u, s string) (*model.ShiftAgentSession, error) {
	if s == "err" {
		return nil, errors.New("err")
	}
	return &model.ShiftAgentSession{ID: s}, nil
}
func (m *mockChatShiftForAPI) UpdateSession(ctx context.Context, s *model.ShiftAgentSession) error {
	return nil
}
func (m *mockChatShiftForAPI) ListActiveUsers(ctx context.Context) ([]*model.User, error) {
	return []*model.User{{
		ID:    "u2",
		Name:  "Test Coworker",
		Email: "u2@google.com",
		Sites: []model.Site{{ID: "44444444-4444-4444-4444-444444440000"}},
	}}, nil
}
func (m *mockChatShiftForAPI) ListActiveOnShiftUsers(ctx context.Context, siteID string) ([]*model.User, error) {
	return []*model.User{{ID: "u1", Name: "Test Coworker", Email: "c@google.com"}}, nil
}
func (m *mockChatShiftForAPI) GetUserProfile(ctx context.Context, id string) (*model.User, error) {
	return &model.User{ID: id, Name: "Test Coworker"}, nil
}

func TestExtraHandlers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 1. TTSHandler offline safety
	ttsHandler := NewTTSHandler()
	assert.NotNil(t, ttsHandler)
	ttsHandler.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/tts", bytes.NewBuffer([]byte(`{"text":"hello"}`)))
	ttsHandler.Synthesize(c)
	assert.Contains(t, []int{http.StatusInternalServerError, http.StatusServiceUnavailable}, w.Code)

	// 2. TraceProxyHandler
	tpHandler, _ := NewTraceProxyHandler(context.Background())
	assert.NotNil(t, tpHandler)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/traces", bytes.NewBuffer([]byte(`{"resourceSpans":[{"resource":{"attributes":[]}}]}`)))
	tpHandler.ProxyTraces(c)

	// 3. VoiceTranslationHandler
	vtHandler := NewVoiceTranslationHandler(&mockUserRepoForAPI{}, &mockTransForAPI{})
	assert.NotNil(t, vtHandler)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/preview", bytes.NewBuffer([]byte(`{"voice_name":"en-US-Journey-F","language_code":"es"}`)))
	vtHandler.PreviewVoice(c)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: "new-u"}}
	c.Request, _ = http.NewRequest("POST", "/profile/new-u", bytes.NewBuffer([]byte(`{"name":"New Profile","email":"new@google.com"}`)))
	vtHandler.SaveProfile(c)
	assert.Equal(t, http.StatusCreated, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: "u1"}}
	c.Request, _ = http.NewRequest("PUT", "/profile/u1", bytes.NewBuffer([]byte(`{"name":"Updated Profile"}`)))
	vtHandler.UpdateProfile(c)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: "u1"}}
	c.Request, _ = http.NewRequest("POST", "/profile/u1/voice/clone", bytes.NewBuffer([]byte(`consent audio bytes`)))
	vtHandler.CloneVoice(c)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/translate/talk?associate_id=u1&target_language=es", bytes.NewBuffer([]byte(`speech bytes`)))
	vtHandler.TranslateTalk(c)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/translate/listen?associate_id=u1&gender=FEMALE&target_language=en", bytes.NewBuffer([]byte(`speech bytes`)))
	vtHandler.TranslateListen(c)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: "u1"}}
	c.Request, _ = http.NewRequest("GET", "/profile/u1", nil)
	vtHandler.GetProfile(c)
	assert.Equal(t, http.StatusOK, w.Code)

	// ListVoices
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/voices", nil)
	vtHandler.ListVoices(c)
	assert.Equal(t, http.StatusOK, w.Code)

	// SaveProfile conflict
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: "u1"}}
	c.Request, _ = http.NewRequest("POST", "/profile/u1", bytes.NewBuffer([]byte(`{"name":"u1"}`)))
	vtHandler.SaveProfile(c)
	assert.Equal(t, http.StatusConflict, w.Code)

	// CloneVoice empty body 400
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: "u1"}}
	c.Request, _ = http.NewRequest("POST", "/profile/u1/voice/clone", nil)
	vtHandler.CloneVoice(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// CloneVoice 404 user not found
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: "err"}}
	c.Request, _ = http.NewRequest("POST", "/profile/err/voice/clone", bytes.NewBuffer([]byte(`audio`)))
	vtHandler.CloneVoice(c)
	assert.Equal(t, http.StatusNotFound, w.Code)

	// CloneVoice 500 cloning failed
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: "u1"}}
	c.Request, _ = http.NewRequest("POST", "/profile/u1/voice/clone", bytes.NewBuffer([]byte(`err`)))
	vtHandler.CloneVoice(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Translate missing params 400
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/translate/talk", nil)
	vtHandler.TranslateTalk(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Translate service error 500
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/translate/talk?associate_id=err&target_language=es", bytes.NewBuffer([]byte(`speech`)))
	vtHandler.TranslateTalk(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/translate/listen?associate_id=err&gender=FEMALE&target_language=en", bytes.NewBuffer([]byte(`speech`)))
	vtHandler.TranslateListen(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// ListVoices error 500
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	reqErr, _ := http.NewRequestWithContext(context.WithValue(context.Background(), "err", true), "GET", "/voices", nil)
	c.Request = reqErr
	vtHandler.ListVoices(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// PreviewVoice languages
	langs := []string{"es", "fr", "de", "it", "ja", "en"}
	for _, l := range langs {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("POST", "/preview", bytes.NewBuffer([]byte(`{"voice_name":"v1","language_code":"`+l+`"}`)))
		vtHandler.PreviewVoice(c)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// PreviewVoice bad request
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/preview", bytes.NewBuffer([]byte(`{bad`)))
	vtHandler.PreviewVoice(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/preview", bytes.NewBuffer([]byte(`{}`)))
	vtHandler.PreviewVoice(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// GenerateDeterministicWeather branches
	GenerateDeterministicWeather("MIA")
	GenerateDeterministicWeather("SEA")
	GenerateDeterministicWeather("ORD")
	GenerateDeterministicWeather("JFK")
	GenerateDeterministicWeather("LAX")
	GenerateDeterministicWeather("OTHER")

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/voices", nil)
	vtHandler.ListVoices(c)
	assert.Equal(t, http.StatusOK, w.Code)

	// 4. ChatHandler
	chatHandler := NewChatHandler(
		&mockOpTaskForAPI{},
		&mockChatShiftForAPI{},
		&mockMCPRAGForAPI{},
		&mockOpAutomationForAPI{},
	)
	assert.NotNil(t, chatHandler)

	intents := []string{
		"drop till register 4",
		"trade shift swap",
		"propose trade for task exec-1234",
		"weather metar observation",
		"sop guidelines",
		"map layout location",
		"what action can i do",
		"hello how are you",
		"airport codes available airports in dallas",
		"select store change store switch store",
		"filter by role select role manager",
		"filter by assignee filter by colleague ryan",
		"show store map vault safe location",
		"show store map register till checkout",
		"show store map greens produce wall wet",
		"show store map showcase atrium display",
		"show store map dock loading receiving cage",
		"create event sensor alert form stockout",
		"translate voice talk listen spanish",
	}

	for _, msg := range intents {
		w = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(w)
		c.Set("userID", "u1")
		c.Params = []gin.Param{{Key: "siteId", Value: "s1"}, {Key: "shiftId", Value: "sh1"}}
		c.Request, _ = http.NewRequest("POST", "/chat", bytes.NewBuffer([]byte(`{"message":"`+msg+`"}`)))
		chatHandler.SendMessage(c)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// SendMessage 401
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/chat", bytes.NewBuffer([]byte(`{"message":"hi"}`)))
	chatHandler.SendMessage(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// SendMessage 500 shift err
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Set("userID", "u1")
	c.Params = []gin.Param{{Key: "shiftId", Value: "err"}}
	c.Request, _ = http.NewRequest("POST", "/chat", bytes.NewBuffer([]byte(`{"message":"hi"}`)))
	chatHandler.SendMessage(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// 6. Middleware tests
	aud, errAud := extractAudienceFromJWT("eyJhbGciOiJSUzI1NiJ9.eyJhdWQiOiJteS1hdWQifQ.sig")
	assert.NoError(t, errAud)
	assert.Equal(t, "my-aud", aud)

	mw := UserContextMiddleware(&mockAdminServiceForAPI{}, Config{})
	r := gin.New()
	r.Use(mw)
	r.GET("/test-mw", func(c *gin.Context) {
		uid, _ := c.Get("userID")
		c.String(http.StatusOK, "uid:%v", uid)
	})

	w = httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test-mw", nil)
	req.Header.Set("Authorization", "Bearer u100")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "uid:u100", w.Body.String())

	// Forbidden bypass token
	w = httptest.NewRecorder()
	reqForbidden, _ := http.NewRequest("GET", "/test-mw", nil)
	reqForbidden.Header.Set("Authorization", "Bearer 00000000-0000-0000-0000-000000000000")
	r.ServeHTTP(w, reqForbidden)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// X-User-ID header (valid user)
	w = httptest.NewRecorder()
	reqXUID, _ := http.NewRequest("GET", "/test-mw", nil)
	reqXUID.Header.Set("X-User-ID", "A2A_USER_u300")
	r.ServeHTTP(w, reqXUID)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "uid:u300", w.Body.String())

	// X-User-ID bypass without Authorization header
	w = httptest.NewRecorder()
	reqXUIDForbidden, _ := http.NewRequest("GET", "/test-mw", nil)
	reqXUIDForbidden.Header.Set("X-User-ID", "00000000-0000-0000-0000-000000000000")
	r.ServeHTTP(w, reqXUIDForbidden)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// X-User-ID bypass WITH Authorization header
	w = httptest.NewRecorder()
	reqXUIDWithAuth, _ := http.NewRequest("GET", "/test-mw", nil)
	reqXUIDWithAuth.Header.Set("X-User-ID", "00000000-0000-0000-0000-000000000000")
	reqXUIDWithAuth.Header.Set("Authorization", "Bearer u400")
	r.ServeHTTP(w, reqXUIDWithAuth)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "uid:u400", w.Body.String())

	// Forbidden user-initiator token
	w = httptest.NewRecorder()
	reqForbiddenInit, _ := http.NewRequest("GET", "/test-mw", nil)
	reqForbiddenInit.Header.Set("Authorization", "Bearer user-initiator")
	r.ServeHTTP(w, reqForbiddenInit)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// IAP header fallback
	w = httptest.NewRecorder()
	reqIAP, _ := http.NewRequest("GET", "/test-mw", nil)
	reqIAP.Header.Set("X-Goog-IAP-JWT-Assertion", "invalid-token")
	r.ServeHTTP(w, reqIAP)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestTelemetryHandler(t *testing.T) {
	h := &TraceProxyHandler{
		httpClient: &http.Client{},
		projectID:  "test-proj",
	}
	assert.NotNil(t, h)

	// empty body
	out, err := h.injectProjectID([]byte{})
	assert.NoError(t, err)
	assert.Empty(t, out)

	// bad json
	_, err = h.injectProjectID([]byte(`{bad`))
	assert.Error(t, err)

	// non-slice resourceSpans
	out, err = h.injectProjectID([]byte(`{"resourceSpans":"not-a-slice"}`))
	assert.NoError(t, err)
	assert.NotEmpty(t, out)

	// OTLP JSON without gcp.project_id
	otlp := []byte(`{"resourceSpans":[{"resource":{"attributes":[{"key":"service.name","value":{"stringValue":"foo"}}]}}]}`)
	out, err = h.injectProjectID(otlp)
	assert.NoError(t, err)
	assert.Contains(t, string(out), `"gcp.project_id"`)
	assert.Contains(t, string(out), `"test-proj"`)

	// OTLP JSON WITH existing gcp.project_id
	otlpWith := []byte(`{"resourceSpans":[{"resource":{"attributes":[{"key":"gcp.project_id","value":{"stringValue":"old-proj"}}]}}]}`)
	out, err = h.injectProjectID(otlpWith)
	assert.NoError(t, err)
	assert.Contains(t, string(out), `"test-proj"`)
	assert.NotContains(t, string(out), `"old-proj"`)

	// Test ProxyTraces with mock transport
	transport := &mockTransport{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(`{"status":"ok"}`)),
				Header:     make(http.Header),
			}, nil
		},
	}
	h.httpClient = &http.Client{Transport: transport}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/traces", bytes.NewBuffer(otlp))
	h.ProxyTraces(c)
	assert.Equal(t, http.StatusOK, w.Code)

	// RoundTrip error
	transport.roundTripFunc = func(req *http.Request) (*http.Response, error) {
		return nil, errors.New("network err")
	}
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/traces", bytes.NewBuffer(otlp))
	h.ProxyTraces(c)
	assert.Equal(t, http.StatusBadGateway, w.Code)
}

type mockTransport struct {
	roundTripFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.roundTripFunc(req)
}

func TestServerAndTTS(t *testing.T) {
	// TTS nil client 503
	hTTS := &TTSHandler{client: nil}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/tts", bytes.NewBuffer([]byte(`{"text":"hello"}`)))
	hTTS.Synthesize(c)
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	hTTS.Close()

	// NewServer call
	srv, err := NewServer(
		Config{},
		&mockAdminServiceForAPI{},
		&mockTaskForAPI{},
		&mockShiftForAPI{},
		&mockRAGForAPI{},
		&mockOpAutomationForAPI{},
		&mockSchedulerForAPI{},
	)
	assert.NoError(t, err)
	assert.NotNil(t, srv)
}

func TestValidateIAPJWT(t *testing.T) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	assert.NoError(t, err)

	kid := "test-kid"
	iapKeysMutex.Lock()
	iapKeysCache = map[string]*ecdsa.PublicKey{kid: &priv.PublicKey}
	iapKeysCacheExpiry = time.Now().Add(1 * time.Hour)
	iapKeysMutex.Unlock()

	makeToken := func(iss, aud, sub, email string, exp time.Time, alg string) string {
		hdr := map[string]string{"alg": alg, "kid": kid}
		hdrBytes, _ := json.Marshal(hdr)
		hdrB64 := base64.RawURLEncoding.EncodeToString(hdrBytes)

		pld := map[string]interface{}{
			"iss":   iss,
			"aud":   aud,
			"sub":   sub,
			"email": email,
			"exp":   exp.Unix(),
		}
		pldBytes, _ := json.Marshal(pld)
		pldB64 := base64.RawURLEncoding.EncodeToString(pldBytes)

		signingInput := hdrB64 + "." + pldB64
		hasher := sha256.New()
		hasher.Write([]byte(signingInput))
		r, s, _ := ecdsa.Sign(rand.Reader, priv, hasher.Sum(nil))

		rBytes := r.Bytes()
		sBytes := s.Bytes()
		sigBytes := make([]byte, 64)
		copy(sigBytes[32-len(rBytes):32], rBytes)
		copy(sigBytes[64-len(sBytes):64], sBytes)
		sigB64 := base64.RawURLEncoding.EncodeToString(sigBytes)

		return signingInput + "." + sigB64
	}

	validTok := makeToken("https://cloud.google.com/iap", "test-aud", "sub1", "user@google.com", time.Now().Add(1*time.Hour), "ES256")
	claims, err := validateIAPJWT(context.Background(), validTok, "test-aud")
	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, "sub1", claims.Sub)

	// Expired
	expTok := makeToken("https://cloud.google.com/iap", "test-aud", "sub1", "user@google.com", time.Now().Add(-1*time.Hour), "ES256")
	_, err = validateIAPJWT(context.Background(), expTok, "test-aud")
	assert.Error(t, err)

	// Audience mismatch
	_, err = validateIAPJWT(context.Background(), validTok, "other-aud")
	assert.Error(t, err)

	// Wrong issuer
	issTok := makeToken("https://other.issuer", "test-aud", "sub1", "user@google.com", time.Now().Add(1*time.Hour), "ES256")
	_, err = validateIAPJWT(context.Background(), issTok, "test-aud")
	assert.Error(t, err)

	// Wrong alg
	algTok := makeToken("https://cloud.google.com/iap", "test-aud", "sub1", "user@google.com", time.Now().Add(1*time.Hour), "HS256")
	_, err = validateIAPJWT(context.Background(), algTok, "test-aud")
	assert.Error(t, err)

	// Malformed
	_, err = validateIAPJWT(context.Background(), "not-a-token", "test-aud")
	assert.Error(t, err)

	// Test UserContextMiddleware with IAP token
	mw := UserContextMiddleware(&mockAdminServiceForAPI{}, Config{IAP: IAPConfig{Audience: "test-aud"}})
	r := gin.New()
	r.Use(mw)
	r.GET("/iap-mw", func(c *gin.Context) {
		uid, _ := c.Get("userID")
		c.String(http.StatusOK, "uid:%v", uid)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/iap-mw", nil)
	req.Header.Set("X-Goog-IAP-JWT-Assertion", validTok)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Register/bind new user via IAP
	newTok := makeToken("https://cloud.google.com/iap", "test-aud", "sub-notfound", "reg@google.com", time.Now().Add(1*time.Hour), "ES256")
	w = httptest.NewRecorder()
	reqNew, _ := http.NewRequest("GET", "/iap-mw", nil)
	reqNew.Header.Set("X-Goog-IAP-JWT-Assertion", newTok)
	r.ServeHTTP(w, reqNew)
	assert.Equal(t, http.StatusOK, w.Code)

	// Error sub via IAP
	errTok := makeToken("https://cloud.google.com/iap", "test-aud", "sub-err", "err@google.com", time.Now().Add(1*time.Hour), "ES256")
	w = httptest.NewRecorder()
	reqErr, _ := http.NewRequest("GET", "/iap-mw", nil)
	reqErr.Header.Set("X-Goog-IAP-JWT-Assertion", errTok)
	r.ServeHTTP(w, reqErr)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
