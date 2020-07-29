package api

import (
	"encoding/json"
	"net/http"
)

const (
	errServer = 1000
)

func ErrorServer(detail string) Error {
	return Error{
		Title:  "Unexpected Error",
		Detail: detail,
		Status: 500,
		Code:   errServer,
	}
}

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
