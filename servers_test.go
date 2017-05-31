package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func Router() *mux.Router {

	r := mux.NewRouter()
	r.HandleFunc("/devices/{id}", webSocketHandler)
	r.HandleFunc("/devWS", webSocketHandler)
	return r
}

func TestCreateEndpoint(t *testing.T) {
	request, _ := http.NewRequest("GET", "/devices", nil)
	response := httptest.NewRecorder()
	Router().ServeHTTP(response, request)
	assert.Equal(t, 200, response.Code, "OK response is expected")
}
