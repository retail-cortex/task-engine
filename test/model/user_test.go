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

package model_test

import (
	"encoding/json"
	"testing"

	"github.com/rmcguinness/gemini_task_engine/pkg/model"
	"github.com/stretchr/testify/assert"
)

func TestUser_MetadataDirectMethods(t *testing.T) {
	t.Run("empty metadata defaults", func(t *testing.T) {
		u := &model.User{}

		gender, err := u.GetGender()
		assert.NoError(t, err)
		assert.Equal(t, model.Gender(""), gender)

		assVoice, err := u.GetAssistantVoice()
		assert.NoError(t, err)
		assert.Equal(t, "", assVoice)

		myVoice, err := u.GetMyVoice()
		assert.NoError(t, err)
		assert.Equal(t, "", myVoice)

		customVoice, err := u.GetCustomVoice()
		assert.NoError(t, err)
		assert.False(t, customVoice)
	})

	t.Run("get and set metadata fields", func(t *testing.T) {
		u := &model.User{}

		err := u.SetGender(model.GenderFemale)
		assert.NoError(t, err)
		g, err := u.GetGender()
		assert.NoError(t, err)
		assert.Equal(t, model.GenderFemale, g)

		err = u.SetAssistantVoice("en-US-Journey-F")
		assert.NoError(t, err)
		av, err := u.GetAssistantVoice()
		assert.NoError(t, err)
		assert.Equal(t, "en-US-Journey-F", av)

		err = u.SetMyVoice("custom-clone-id-123")
		assert.NoError(t, err)
		mv, err := u.GetMyVoice()
		assert.NoError(t, err)
		assert.Equal(t, "custom-clone-id-123", mv)

		err = u.SetCustomVoice(true)
		assert.NoError(t, err)
		cv, err := u.GetCustomVoice()
		assert.NoError(t, err)
		assert.True(t, cv)
	})

	t.Run("preserves unrelated metadata fields", func(t *testing.T) {
		u := &model.User{
			Metadata: model.JSONB(`{"unrelated_key": "some_value", "gender": "Male"}`),
		}

		// Verify existing unrelated key is present by unmarshalling manually
		var raw map[string]interface{}
		err := json.Unmarshal(u.Metadata, &raw)
		assert.NoError(t, err)
		assert.Equal(t, "some_value", raw["unrelated_key"])
		assert.Equal(t, "Male", raw["gender"])

		// Update AssistantVoice
		err = u.SetAssistantVoice("en-GB-Standard-A")
		assert.NoError(t, err)

		// Verify gender and unrelated key are still there
		g, err := u.GetGender()
		assert.NoError(t, err)
		assert.Equal(t, model.GenderMale, g)

		err = json.Unmarshal(u.Metadata, &raw)
		assert.NoError(t, err)
		assert.Equal(t, "some_value", raw["unrelated_key"])
		assert.Equal(t, "en-GB-Standard-A", raw["assistant_voice"])
	})

}

func TestUser_LanguageRelationship(t *testing.T) {
	t.Run("empty relationship defaults to nil", func(t *testing.T) {
		u := &model.User{}
		assert.Nil(t, u.GetPreferredLanguage())
		assert.Nil(t, u.PreferredLanguageID)
	})

	t.Run("set preferred language relation", func(t *testing.T) {
		u := &model.User{}
		lang := &model.Language{
			ID:   "a0000000-0000-0000-0000-000000000001",
			Code: model.LanguageEnglish,
			Name: "English (United States)",
		}

		u.SetPreferredLanguage(lang)
		assert.Equal(t, lang, u.GetPreferredLanguage())
		assert.Equal(t, &lang.ID, u.PreferredLanguageID)

		// Set to nil
		u.SetPreferredLanguage(nil)
		assert.Nil(t, u.GetPreferredLanguage())
		assert.Nil(t, u.PreferredLanguageID)
	})
}

func TestUser_ConsentRecording(t *testing.T) {
	t.Run("set consent recording", func(t *testing.T) {
		recordingBytes := []byte{0x01, 0x02, 0x03, 0x04}
		u := &model.User{
			ConsentRecording: recordingBytes,
		}

		assert.Equal(t, recordingBytes, u.ConsentRecording)
	})
}
