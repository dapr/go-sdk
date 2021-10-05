package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStoppingUnstartedService(t *testing.T) {
	s := newServer("", nil)
	assert.NotNil(t, s)
	err := s.Stop()
	assert.NoError(t, err)
}

func TestStoppingStartedService(t *testing.T) {
	s := newServer(":3333", nil)
	assert.NotNil(t, s)

	go func() {
		if err := s.Start(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()
	// Wait for the server to start
	time.Sleep(200 * time.Millisecond)
	assert.NoError(t, s.Stop())
}

func TestStartingStoppedService(t *testing.T) {
	s := newServer(":3333", nil)
	assert.NotNil(t, s)
	stopErr := s.Stop()
	assert.NoError(t, stopErr)

	startErr := s.Start()
	assert.Error(t, startErr, "expected starting a stopped server to raise an error")
	assert.Equal(t, startErr.Error(), http.ErrServerClosed.Error())
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
