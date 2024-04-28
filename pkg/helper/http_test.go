package helper_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kyverno/policy-reporter/pkg/helper"
)

func TestSendJSONResponse(t *testing.T) {
	t.Run("success response", func(t *testing.T) {
		w := httptest.NewRecorder()

		helper.SendJSONResponse(w, []string{"default", "user"}, nil)

		assert.Equal(t, http.StatusOK, w.Code)

		resp := make([]string, 0, 2)

		json.NewDecoder(w.Body).Decode(&resp)

		assert.Equal(t, []string{"default", "user"}, resp)
	})

	t.Run("error response", func(t *testing.T) {
		w := httptest.NewRecorder()

		helper.SendJSONResponse(w, nil, errors.New("error"))

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		resp := make(map[string]string, 0)

		json.NewDecoder(w.Body).Decode(&resp)

		assert.Equal(t, map[string]string{"message": "error"}, resp)
	})
}
