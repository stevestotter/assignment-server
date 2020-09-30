// +build ignore

package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"stevestotter/assignment-server/assignment"
	mock_assignment "stevestotter/assignment-server/mocks/assignment"

	"testing"

	"github.com/golang/mock/gomock"

	"github.com/stretchr/testify/assert"
)

const apiPort string = "1005"

func TestBuyWhenValidPriceAndQuantity(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	reqJSON := `{"price": "2.24", "quantity": "0.5"}`
	expAssignment := assignment.Assignment{
		Price:    "2.24",
		Quantity: "0.5",
	}

	mockSubmitter := mock_assignment.NewMockSubmitter(ctrl)
	mockSubmitter.EXPECT().
		SubmitAssignment(expAssignment, assignment.Buy).
		Times(1).
		Return(nil)

	a := API{Port: apiPort, AssignmentSubmitter: mockSubmitter}
	err := a.Start()
	defer a.server.Shutdown(context.Background())
	assert.NoError(t, err)

	req, err := http.NewRequest("POST",
		fmt.Sprintf("http://localhost:%s/buy", apiPort),
		bytes.NewBufferString(reqJSON),
	)
	req.Close = true
	assert.NoError(t, err)

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
}

func TestBuyReturnsBadRequestWhenInvalidPriceAndQuantity(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSubmitter := mock_assignment.NewMockSubmitter(ctrl)
	a := API{Port: apiPort, AssignmentSubmitter: mockSubmitter}

	err := a.Start()
	defer a.server.Shutdown(context.Background())
	assert.NoError(t, err)

	httpClient := &http.Client{}

	tests := map[string]struct {
		reqJSON string
	}{
		"Price not a number":               {reqJSON: `{ "price": "2.a4", "quantity": "0.5" }`},
		"Quantity not a number":            {reqJSON: `{ "price": "2.34", "quantity": "0.+" }`},
		"Price negative":                   {reqJSON: `{ "price": "-2.34", "quantity": "0.5" }`},
		"Quantity negative":                {reqJSON: `{ "price": "2.34", "quantity": "-0.5" }`},
		"Price more than 2 decimal places": {reqJSON: `{ "price": "2.345", "quantity": "0.5" }`},
		"Price less than 2 decimal places": {reqJSON: `{ "price": "2.3", "quantity": "0.5" }`},
		"Price not specified":              {reqJSON: `{ "quantity": "0.5" }`},
		"Quantity not specified":           {reqJSON: `{ "price": "0.5" }`},
		"Invalid JSON":                     {reqJSON: `{ "price": "0.5", quantity: "0.2" }`},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			req, err := http.NewRequest("POST",
				fmt.Sprintf("http://localhost:%s/buy", apiPort),
				bytes.NewBufferString(tc.reqJSON),
			)
			req.Close = true
			assert.NoError(t, err)

			resp, err := httpClient.Do(req)
			assert.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		})
	}
}

func TestBuyReturnsServerErrorIfMessageQueueReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	reqJSON := `{"price": "2.24", "quantity": "0.5"}`

	mockSubmitter := mock_assignment.NewMockSubmitter(ctrl)
	mockSubmitter.EXPECT().
		SubmitAssignment(gomock.Any(), gomock.Any()).
		Times(1).
		Return(errors.New("queue error"))

	a := API{Port: apiPort, AssignmentSubmitter: mockSubmitter}

	err := a.Start()
	defer a.server.Shutdown(context.Background())
	assert.NoError(t, err)

	req, err := http.NewRequest("POST",
		fmt.Sprintf("http://localhost:%s/buy", apiPort),
		bytes.NewBufferString(reqJSON),
	)
	req.Close = true
	assert.NoError(t, err)

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	expectedErr, _ := json.Marshal(ErrorServer("Failed to submit assignment"))
	body, _ := ioutil.ReadAll(resp.Body)

	assert.JSONEq(t, string(expectedErr), string(body))

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestSellWhenValidPriceAndQuantity(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	reqJSON := `{"price": "2.24", "quantity": "0.5"}`
	expAssignment := assignment.Assignment{
		Price:    "2.24",
		Quantity: "0.5",
	}

	mockSubmitter := mock_assignment.NewMockSubmitter(ctrl)
	mockSubmitter.EXPECT().
		SubmitAssignment(expAssignment, assignment.Sell).
		Times(1).
		Return(nil)

	a := API{Port: apiPort, AssignmentSubmitter: mockSubmitter}

	err := a.Start()
	defer a.server.Shutdown(context.Background())
	assert.NoError(t, err)

	req, err := http.NewRequest("POST",
		fmt.Sprintf("http://localhost:%s/sell", apiPort),
		bytes.NewBufferString(reqJSON),
	)
	req.Close = true
	assert.NoError(t, err)

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
}

func TestSellReturnsBadRequestWhenInvalidPriceAndQuantity(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSubmitter := mock_assignment.NewMockSubmitter(ctrl)
	a := API{Port: apiPort, AssignmentSubmitter: mockSubmitter}

	err := a.Start()
	defer a.server.Shutdown(context.Background())
	assert.NoError(t, err)

	httpClient := &http.Client{}

	tests := map[string]struct {
		reqJSON string
	}{
		"Price not a number":               {reqJSON: `{ "price": "2.a4", "quantity": "0.5" }`},
		"Quantity not a number":            {reqJSON: `{ "price": "2.34", "quantity": "0.+" }`},
		"Price negative":                   {reqJSON: `{ "price": "-2.34", "quantity": "0.5" }`},
		"Quantity negative":                {reqJSON: `{ "price": "2.34", "quantity": "-0.5" }`},
		"Price more than 2 decimal places": {reqJSON: `{ "price": "2.345", "quantity": "0.5" }`},
		"Price less than 2 decimal places": {reqJSON: `{ "price": "2.3", "quantity": "0.5" }`},
		"Price not specified":              {reqJSON: `{ "quantity": "0.5" }`},
		"Quantity not specified":           {reqJSON: `{ "price": "0.5" }`},
		"Invalid JSON":                     {reqJSON: `{ "price": "0.5", quantity: "0.2" }`},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			req, err := http.NewRequest("POST",
				fmt.Sprintf("http://localhost:%s/sell", apiPort),
				bytes.NewBufferString(tc.reqJSON),
			)
			req.Close = true
			assert.NoError(t, err)

			resp, err := httpClient.Do(req)
			assert.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		})
	}
}

func TestSellReturnsServerErrorIfMessageQueueReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	reqJSON := `{"price": "2.24", "quantity": "0.5"}`

	mockSubmitter := mock_assignment.NewMockSubmitter(ctrl)
	mockSubmitter.EXPECT().
		SubmitAssignment(gomock.Any(), gomock.Any()).
		Times(1).
		Return(errors.New("queue error"))

	a := API{Port: apiPort, AssignmentSubmitter: mockSubmitter}

	err := a.Start()
	defer a.server.Shutdown(context.Background())
	assert.NoError(t, err)

	req, err := http.NewRequest("POST",
		fmt.Sprintf("http://localhost:%s/sell", apiPort),
		bytes.NewBufferString(reqJSON),
	)
	req.Close = true
	assert.NoError(t, err)

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	expectedErr, _ := json.Marshal(ErrorServer("Failed to submit assignment"))
	body, _ := ioutil.ReadAll(resp.Body)

	assert.JSONEq(t, string(expectedErr), string(body))

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}
