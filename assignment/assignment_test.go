// +build ignore

package assignment

import (
	"encoding/json"
	"errors"
	"stevestotter/assignment-server/event"
	mock_event "stevestotter/assignment-server/mocks/event"
	"strconv"
	"sync"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestAssignmentGeneratorGeneratesNewSellAssignmentOnBuyTrade(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	buyTrade := []byte(`{"assignmentId": 123, "price": "2.24", "quantity": "0.5"}`)
	tradeChan := make(chan []byte)

	mockListenPublisher := mock_event.NewMockListenPublisher(ctrl)
	mockListenPublisher.EXPECT().
		Subscribe(event.TopicBuyerTrade, event.GroupBuyer).
		Times(1).
		Return(tradeChan, nil)

	mockListenPublisher.EXPECT().
		Subscribe(event.TopicSellerTrade, event.GroupSeller).
		Times(1).
		Return(nil, nil)

	var wg sync.WaitGroup
	wg.Add(1)
	defer wg.Wait()
	mockListenPublisher.EXPECT().
		Publish(gomock.Any(), event.TopicSellerAssignment).
		Times(1).
		Do(func(message []byte, topic string) {
			assignment := &Assignment{}
			err := json.Unmarshal(message, &assignment)
			assert.NoError(t, err)
			assert.Equal(t, "0.5", assignment.Quantity)

			fPrice, err := strconv.ParseFloat(assignment.Price, 64)
			assert.NoError(t, err)
			assert.Greater(t, fPrice, 2.24*1.01)
			assert.Less(t, fPrice, 2.24*1.02)

			wg.Done()
		}).
		Return(nil)

	g := Generator{
		MessageQueue:        mockListenPublisher,
		PercentageChangeMin: 1,
		PercentageChangeMax: 2,
	}

	go g.GenerateFromTrades()

	tradeChan <- buyTrade
}

func TestAssignmentGeneratorReturnsErrorWhenFailureToSubscribeToBuyTopic(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedErr := errors.New("Subscribe Error")

	mockListenPublisher := mock_event.NewMockListenPublisher(ctrl)
	mockListenPublisher.EXPECT().
		Subscribe(event.TopicBuyerTrade, event.GroupBuyer).
		Times(1).
		Return(nil, expectedErr)

	g := Generator{
		MessageQueue:        mockListenPublisher,
		PercentageChangeMin: 1,
		PercentageChangeMax: 2,
	}

	err := g.GenerateFromTrades()
	assert.Equal(t, expectedErr, err)
}

func TestAssignmentGeneratorGeneratesNewPurchaseAssignmentOnSellTrade(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sellTrade := []byte(`{"assignmentId": 123, "price": "2.24", "quantity": "0.5"}`)
	tradeChan := make(chan []byte)

	mockListenPublisher := mock_event.NewMockListenPublisher(ctrl)
	mockListenPublisher.EXPECT().
		Subscribe(event.TopicSellerTrade, event.GroupSeller).
		Times(1).
		Return(tradeChan, nil)

	mockListenPublisher.EXPECT().
		Subscribe(event.TopicBuyerTrade, event.GroupBuyer).
		Times(1).
		Return(nil, nil)

	var wg sync.WaitGroup
	wg.Add(1)
	defer wg.Wait()
	mockListenPublisher.EXPECT().
		Publish(gomock.Any(), event.TopicBuyerAssignment).
		Times(1).
		Do(func(message []byte, topic string) {
			assignment := &Assignment{}
			err := json.Unmarshal(message, &assignment)
			assert.NoError(t, err)
			assert.Equal(t, "0.5", assignment.Quantity)

			fPrice, err := strconv.ParseFloat(assignment.Price, 64)
			assert.NoError(t, err)
			assert.Greater(t, fPrice, 2.24*0.98)
			assert.Less(t, fPrice, 2.24*0.99)

			wg.Done()
		}).
		Return(nil)

	g := Generator{
		MessageQueue:        mockListenPublisher,
		PercentageChangeMin: 1,
		PercentageChangeMax: 2,
	}

	go g.GenerateFromTrades()

	tradeChan <- sellTrade
}

func TestAssignmentGeneratorReturnsErrorWhenFailureToSubscribeToSellTopic(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedErr := errors.New("Subscribe Error")

	mockListenPublisher := mock_event.NewMockListenPublisher(ctrl)
	mockListenPublisher.EXPECT().
		Subscribe(event.TopicSellerTrade, event.GroupSeller).
		Times(1).
		Return(nil, expectedErr)

	mockListenPublisher.EXPECT().
		Subscribe(event.TopicBuyerTrade, event.GroupBuyer).
		Times(1).
		Return(nil, nil)

	g := Generator{
		MessageQueue:        mockListenPublisher,
		PercentageChangeMin: 1,
		PercentageChangeMax: 2,
	}

	err := g.GenerateFromTrades()
	assert.Equal(t, expectedErr, err)
}
