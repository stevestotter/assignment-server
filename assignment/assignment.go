package assignment

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"stevestotter/assignment-server/event"
	"strconv"
)

// Assignment is a directive given to agents (buy/sell) in a market
type Assignment struct {
	Price    string `json:"price" validate:"required,currency"`
	Quantity string `json:"quantity" validate:"required,numeric,excludes=-"`
}

// Generator generates new assignments
type Generator struct {
	MessageQueue        event.ListenPublisher
	PercentageChangeMin float64
	PercentageChangeMax float64
}

func randomFloat64(min, max float64) float64 {
	return min + (rand.Float64() * (max - min))
}

// GenerateFromTrades listens to trades and generates new assignments
// based off of their trade value. This function is blocking.
func (a *Generator) GenerateFromTrades() error {
	buyTrades, err := a.MessageQueue.Subscribe(event.TopicBuyerTrade, event.GroupBuyer)
	if err != nil {
		return err
	}

	sellTrades, err := a.MessageQueue.Subscribe(event.TopicSellerTrade, event.GroupSeller)
	if err != nil {
		return err
	}

	go func() {
		for b := range buyTrades {
			// TODO: Change logger
			log.Printf("Got buy trade: %s", string(b))

			trade := &event.Trade{}
			if err := json.Unmarshal(b, &trade); err != nil {
				log.Printf("Error unmarshalling buy trade from queue: %s", err)
				continue
			}

			percentIncrease := randomFloat64(a.PercentageChangeMin, a.PercentageChangeMax)

			err := a.submitNewAssignmentFromTrade(trade, percentIncrease, event.TopicSellerAssignment)
			if err != nil {
				log.Printf("Error submitting new assignment: %s", err)
				continue
			}
		}
	}()

	for s := range sellTrades {
		// TODO: Change logger
		log.Printf("Got sell trade: %s", string(s))

		trade := &event.Trade{}
		if err := json.Unmarshal(s, &trade); err != nil {
			log.Printf("Error unmarshalling sell trade from queue: %s", err)
			continue
		}

		percentDecrease := -1 * randomFloat64(a.PercentageChangeMin, a.PercentageChangeMax)

		err := a.submitNewAssignmentFromTrade(trade, percentDecrease, event.TopicBuyerAssignment)
		if err != nil {
			log.Printf("Error submitting new assignment: %s", err)
			continue
		}
	}

	return nil
}

func (a *Generator) submitNewAssignmentFromTrade(trade *event.Trade, percentChange float64, topic string) error {
	floatPrice, err := strconv.ParseFloat(trade.Price, 64)
	newPrice := floatPrice * ((100 + percentChange) / 100)
	// normalise to 0.01 precision for currencies
	newPrice = (math.Floor(newPrice*100 + 0.5)) / 100

	newAssignment := Assignment{
		Price:    fmt.Sprintf("%.2f", newPrice),
		Quantity: trade.Quantity,
	}

	assignmentBytes, err := json.Marshal(newAssignment)
	if err != nil {
		return fmt.Errorf("Failed to marshal new assignment: %s", err)
	}

	err = a.MessageQueue.Publish(assignmentBytes, topic)
	if err != nil {
		return fmt.Errorf("Failed to publish assignment: %s", err)
	}

	return nil
}
