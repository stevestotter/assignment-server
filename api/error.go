package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/stevestotter/assignment-server/event"
)

const (
	errUnexpected  = 1000
	errSubmitError = 1001
)

// ErrorUnexpected is a detailed HTTP 500 message for unexpected errors
func ErrorUnexpected(detail string) Error {
	return Error{
		Title:  "Unexpected Error",
		Detail: detail,
		Status: http.StatusInternalServerError,
		Code:   errUnexpected,
	}
}

// Error is a JSON error that adheres to the JSON API spec
type Error struct {
	Title  string `json:"title"`
	Detail string `json:"detail"`
	Status int    `json:"status,string"`
	Code   int    `json:"code,string"`
}

// WriteJSON writes a JSON representation of the error to the HTTP ResponseWriter
func (e *Error) WriteJSON(w http.ResponseWriter) error {
	b, err := json.Marshal(e)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.Status)
	w.Write(b)
	return nil
}

func handleError(w http.ResponseWriter, err error) {
	var apiErr Error

	switch err {
	case event.ErrQueueWrite:
		apiErr = Error{
			Title:  "Could not submit event",
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
			Code:   errSubmitError,
		}
	default:
		apiErr = ErrorUnexpected(err.Error())
	}

	log.Printf("%s: %s\n", apiErr.Title, err)
	apiErr.WriteJSON(w)
}
