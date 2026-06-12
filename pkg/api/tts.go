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
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	tts "cloud.google.com/go/texttospeech/apiv1"
	ttspb "cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
)

type TTSRequest struct {
	Text string `json:"text" binding:"required"`
}

type TTSHandler struct {
	client *tts.Client
}

func NewTTSHandler() *TTSHandler {
	client, err := tts.NewClient(context.Background())
	if err != nil {
		log.Printf("WARNING: Failed to initialize Google Text-to-Speech client: %v. Speech synthesis will be degraded.", err)
		return &TTSHandler{client: nil}
	}
	log.Println("[TTS] Google Text-to-Speech client successfully initialized and cached globally.")
	return &TTSHandler{client: client}
}

func (h *TTSHandler) Close() {
	if h.client != nil {
		h.client.Close()
	}
}

func (h *TTSHandler) Synthesize(c *gin.Context) {
	if h.client == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Google Text-to-Speech service is offline or not initialized"})
		return
	}

	var req TTSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()

	// Configure the synthesis request
	// Journey voices are the most premium, natural, and realistic voices available in Google Cloud
	ttsReq := &ttspb.SynthesizeSpeechRequest{
		Input: &ttspb.SynthesisInput{
			InputSource: &ttspb.SynthesisInput_Text{Text: req.Text},
		},
		Voice: &ttspb.VoiceSelectionParams{
			LanguageCode: "en-US",
			Name:         "en-US-Journey-F", // Beautiful, highly natural Journey Female voice
		},
		AudioConfig: &ttspb.AudioConfig{
			AudioEncoding: ttspb.AudioEncoding_MP3,
		},
	}

	resp, err := h.client.SynthesizeSpeech(ctx, ttsReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to synthesize speech: " + err.Error()})
		return
	}

	// Stream the raw MP3 bytes back to the frontend
	c.Data(http.StatusOK, "audio/mpeg", resp.AudioContent)
}
