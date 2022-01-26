package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateEnvelopeIsPost(t *testing.T) {

	res, err := http.Get("http://localhost:8080/envelope")

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, http.StatusMethodNotAllowed, res.StatusCode)

}

func createEnvelopeClient(params map[string]interface{}) (*http.Response, []byte) {

	paramsJson, _ := json.Marshal(params)

	req := httptest.NewRequest(http.MethodPost, "/envelope", bytes.NewReader(paramsJson))
	w := httptest.NewRecorder()

	CreateEnvelopeHandler(w, req)

	res := w.Result()
	body, _ := ioutil.ReadAll(res.Body)
	defer res.Body.Close()

	return res, body
}

func TestCreateEnvelopeWithBlankJson(t *testing.T) {

	res, _ := createEnvelopeClient(map[string]interface{}{})

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

}
