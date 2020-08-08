package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"stevestotter/assignment-server/event"
	mock_event "stevestotter/assignment-server/mocks/event"

	"testing"

	"github.com/golang/mock/gomock"

	"github.com/stretchr/testify/assert"
)

const apiPort string = "1005"

func TestBuyWhenValidPriceAndQuantity(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	reqJSON := `{"price": "2.24", "quantity": "0.5"}`

	mockPublisher := mock_event.NewMockPublisher(ctrl)
	mockPublisher.EXPECT().
		Publish(gomock.Any(), event.TopicBuyerAssignment).
		Times(1).
		Return(nil).
		Do(func(b []byte, topic string) {
			assert.JSONEq(t, reqJSON, string(b))
		})

	a := API{Port: apiPort, MessageQueue: mockPublisher}
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
	defer resp.Body.Close()
	assert.NoError(t, err)

	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
}

func TestBuyReturnsBadRequestWhenInvalidPriceAndQuantity(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPublisher := mock_event.NewMockPublisher(ctrl)
	a := API{Port: apiPort, MessageQueue: mockPublisher}

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
			defer resp.Body.Close()
			assert.NoError(t, err)

			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		})
	}
}

func TestBuyReturnsServerErrorIfMessageQueueReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	reqJSON := `{"price": "2.24", "quantity": "0.5"}`

	mockPublisher := mock_event.NewMockPublisher(ctrl)
	mockPublisher.EXPECT().
		Publish(gomock.Any(), gomock.Any()).
		Times(1).
		Return(errors.New("queue error"))

	a := API{Port: apiPort, MessageQueue: mockPublisher}

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
	defer resp.Body.Close()
	assert.NoError(t, err)

	expectedErr, _ := json.Marshal(ErrorServer("Failed to publish assignment to message queue"))
	body, _ := ioutil.ReadAll(resp.Body)

	assert.JSONEq(t, string(expectedErr), string(body))

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestSellWhenValidPriceAndQuantity(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	reqJSON := `{"price": "2.24", "quantity": "0.5"}`

	mockPublisher := mock_event.NewMockPublisher(ctrl)
	mockPublisher.EXPECT().
		Publish(gomock.Any(), event.TopicSellerAssignment).
		Times(1).
		Return(nil).
		Do(func(b []byte, topic string) {
			assert.JSONEq(t, reqJSON, string(b))
		})

	a := API{Port: apiPort, MessageQueue: mockPublisher}

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
	defer resp.Body.Close()
	assert.NoError(t, err)

	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
}

func TestSellReturnsBadRequestWhenInvalidPriceAndQuantity(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPublisher := mock_event.NewMockPublisher(ctrl)
	a := API{Port: apiPort, MessageQueue: mockPublisher}

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
			defer resp.Body.Close()
			assert.NoError(t, err)

			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		})
	}
}

func TestSellReturnsServerErrorIfMessageQueueReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	reqJSON := `{"price": "2.24", "quantity": "0.5"}`

	mockPublisher := mock_event.NewMockPublisher(ctrl)
	mockPublisher.EXPECT().
		Publish(gomock.Any(), gomock.Any()).
		Times(1).
		Return(errors.New("queue error"))

	a := API{Port: apiPort, MessageQueue: mockPublisher}

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
	defer resp.Body.Close()
	assert.NoError(t, err)

	expectedErr, _ := json.Marshal(ErrorServer("Failed to publish assignment to message queue"))
	body, _ := ioutil.ReadAll(resp.Body)

	assert.JSONEq(t, string(expectedErr), string(body))

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}
