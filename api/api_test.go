package api_test

import (
	"bytes"
	"net/http"
	"stevestotter/assignment-server/api"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuyWhenValidPriceAndQuantity(t *testing.T) {
	s := api.Server{Port: "1001"}
	go s.Start()

	reqJSON := `{ "price": "2.24", "quantity": "0.5" }`

	req, err := http.NewRequest("POST",
		"http://localhost:1001/buy",
		bytes.NewBufferString(reqJSON),
	)
	assert.NoError(t, err)

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// TODO: Test interaction with MockPublisher

}

func TestBuyReturnsBadRequestWhenInvalidPriceAndQuantity(t *testing.T) {
	s := api.Server{Port: "1001"}
	go s.Start()
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

// TODO: Add Sell tests
