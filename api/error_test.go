package api

import (
	"io/ioutil"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteJSONReturnsErrorAsJSON(t *testing.T) {
	err := &Error{
		Title:  "a-title",
		Detail: "some-detail",
		Status: 404,
		Code:   123,
	}

	expJSON := `{ 
		"title": "a-title",
		"detail": "some-detail",
		"status": "404",
		"code": "123" 
	}`

	w := httptest.NewRecorder()

	err.WriteJSON(w)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	assert.JSONEq(t, expJSON, string(body))
	assert.Equal(t, 404, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
}
