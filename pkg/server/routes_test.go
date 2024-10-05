package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cisco-eti/sre-go-helloworld/pkg/models"
	etilogger "wwwin-github.cisco.com/eti/sre-go-logger"
)

func getDeviceZone(t *testing.T, path string, want string, ret string) {
	t.Run(ret, func(t *testing.T) {
		request, err := http.NewRequest(http.MethodGet, "/v1/device/"+path, nil)
		require.NoError(t, err)

		response := httptest.NewRecorder()
		request.Header.Set("Authorization", "Bearer 123456")

		s := New(etilogger.NewNop(), nil, nil)
		router := s.Router()
		router.ServeHTTP(response, request)

		apiRes := models.APIResponse{}
		err = json.Unmarshal(response.Body.Bytes(), &apiRes)
		require.NoError(t, err, "response body: %s", string(response.Body.Bytes()))

		assert.Equal(t, want, apiRes.Data)
	})
}

func TestRoutes_GetDeviceZones(t *testing.T) {
	getDeviceZone(t, "A", "Plumbing", "returns Device_A zone")
	getDeviceZone(t, "B", "Gardening", "returns Device_B zone")
}
