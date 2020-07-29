package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	mock_event "stevestotter/assignment-server/mocks/event"

	"testing"

	"github.com/golang/mock/gomock"

	"github.com/stretchr/testify/assert"
)

func TestBuyWhenValidPriceAndQuantity(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	reqJSON := `{"price": "2.24", "quantity": "0.5"}`

	mockPublisher := mock_event.NewMockPublisher(ctrl)
	mockPublisher.EXPECT().
		Publish(gomock.Any()).
		Times(1).
		Return(nil).
		Do(func(b []byte) {
			assert.JSONEq(t,
				`{"price": "2.24", "quantity": "0.5", "type": "buy"}`,
				string(b))
		})

	a := API{Port: "1001", MessageQueue: mockPublisher}
	go a.Start()
	t.Cleanup(func() {
		a.server.Close()
	})

	req, err := http.NewRequest("POST",
		"http://localhost:1001/buy",
		bytes.NewBufferString(reqJSON),
	)
	assert.NoError(t, err)

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
}

func TestBuyReturnsBadRequestWhenInvalidPriceAndQuantity(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPublisher := mock_event.NewMockPublisher(ctrl)
	a := API{Port: "1001", MessageQueue: mockPublisher}

	go a.Start()
	t.Cleanup(func() {
		a.server.Close()
	})

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
				"http://localhost:1001/buy",
				bytes.NewBufferString(tc.reqJSON),
			)
			assert.NoError(t, err)

			resp, err := httpClient.Do(req)
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
		Publish(gomock.Any()).
		Times(1).
		Return(errors.New("queue error"))

	a := API{Port: "1001", MessageQueue: mockPublisher}
	go a.Start()
	t.Cleanup(func() {
		a.server.Close()
	})

	req, err := http.NewRequest("POST",
		"http://localhost:1001/buy",
		bytes.NewBufferString(reqJSON),
	)
	assert.NoError(t, err)

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	assert.NoError(t, err)

	expectedErr, _ := json.Marshal(ErrorServer("Failed to publish assignment to message queue"))
	body, _ := ioutil.ReadAll(resp.Body)

	assert.JSONEq(t, string(expectedErr), string(body))

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

// TODO: Add Sell tests
func TestSellWhenValidPriceAndQuantity(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	reqJSON := `{"price": "2.24", "quantity": "0.5"}`

	mockPublisher := mock_event.NewMockPublisher(ctrl)
	mockPublisher.EXPECT().
		Publish(gomock.Any()).
		Times(1).
		Return(nil).
		Do(func(b []byte) {
			assert.JSONEq(t,
				`{"price": "2.24", "quantity": "0.5", "type": "sell"}`,
				string(b))
		})

	a := API{Port: "1001", MessageQueue: mockPublisher}
	go a.Start()
	t.Cleanup(func() {
		a.server.Close()
	})

	req, err := http.NewRequest("POST",
		"http://localhost:1001/sell",
		bytes.NewBufferString(reqJSON),
	)
	assert.NoError(t, err)

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
}

func TestSellReturnsBadRequestWhenInvalidPriceAndQuantity(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPublisher := mock_event.NewMockPublisher(ctrl)
	a := API{Port: "1001", MessageQueue: mockPublisher}

	go a.Start()
	t.Cleanup(func() {
		a.server.Close()
	})

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
				"http://localhost:1001/sell",
				bytes.NewBufferString(tc.reqJSON),
			)
			assert.NoError(t, err)

			resp, err := httpClient.Do(req)
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
		Publish(gomock.Any()).
		Times(1).
		Return(errors.New("queue error"))

	a := API{Port: "1001", MessageQueue: mockPublisher}
	go a.Start()
	t.Cleanup(func() {
		a.server.Close()
	})

	req, err := http.NewRequest("POST",
		"http://localhost:1001/sell",
		bytes.NewBufferString(reqJSON),
	)
	assert.NoError(t, err)

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	assert.NoError(t, err)

	expectedErr, _ := json.Marshal(ErrorServer("Failed to publish assignment to message queue"))
	body, _ := ioutil.ReadAll(resp.Body)

	assert.JSONEq(t, string(expectedErr), string(body))

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}
