package http

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStoppingUnstartedService(t *testing.T) {
	s := newServer("", nil)
	assert.NotNil(t, s)
	err := s.Stop()
	assert.NoError(t, err)
}

func TestSettingOptions(t *testing.T) {
	req, err := http.NewRequest(http.MethodOptions, "/", nil)
	assert.NoErrorf(t, err, "error creating request")
	w := httptest.NewRecorder()
	setOptions(w, req)
	resp := w.Result()
	defer resp.Body.Close()
	assert.NotNil(t, resp)
	assert.Equal(t, "*", resp.Header.Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "POST,OPTIONS", resp.Header.Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "authorization, origin, content-type, accept", resp.Header.Get("Access-Control-Allow-Headers"))
	assert.Equal(t, "POST,OPTIONS", resp.Header.Get("Allow"))
}

func testRequest(t *testing.T, s *Server, r *http.Request, expectedStatusCode int) {
	rr := httptest.NewRecorder()
	s.mux.ServeHTTP(rr, r)
	resp := rr.Result()
	defer resp.Body.Close()
	assert.NotNil(t, resp)
	assert.Equal(t, expectedStatusCode, resp.StatusCode)
}

func testRequestWithResponseBody(t *testing.T, s *Server, r *http.Request, expectedStatusCode int, expectedBody []byte) {
	rr := httptest.NewRecorder()
	s.mux.ServeHTTP(rr, r)
	rez := rr.Result()
	defer rez.Body.Close()
	rspBody, err := io.ReadAll(rez.Body)
	assert.Nil(t, err)
	assert.NotNil(t, rez)
	assert.Equal(t, expectedStatusCode, rez.StatusCode)
	assert.Equal(t, expectedBody, rspBody)
}
