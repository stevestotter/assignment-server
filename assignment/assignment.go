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

//go:generate go run -mod=mod github.com/golang/mock/mockgen --build_flags=-mod=mod --source=assignment.go --destination=../mocks/assignment/assignment.go

// Assignment is a directive given to agents (buy/sell) in a market
type Assignment struct {
	Price    string `json:"price" validate:"required,currency"`
	Quantity string `json:"quantity" validate:"required,numeric,excludes=-"`
}

// Type defines the type of assignment - either buy or sell
type Type int

const (
	// Buy is a type of assignment
	Buy Type = iota
	// Sell is a type of assignment
	Sell
)

// Submitter defines the ability to submit an assignment
type Submitter interface {
	SubmitAssignment(a Assignment, t Type) error
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
func (g *Generator) GenerateFromTrades() error {
	buyTrades, err := g.MessageQueue.Subscribe(event.TopicBuyerTrade, event.GroupBuyer)
	if err != nil {
		return err
	}

	sellTrades, err := g.MessageQueue.Subscribe(event.TopicSellerTrade, event.GroupSeller)
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

			percentIncrease := randomFloat64(g.PercentageChangeMin, g.PercentageChangeMax)

			err := g.submitNewAssignmentFromTrade(trade, percentIncrease, Sell)
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

		percentDecrease := -1 * randomFloat64(g.PercentageChangeMin, g.PercentageChangeMax)

		err := g.submitNewAssignmentFromTrade(trade, percentDecrease, Buy)
		if err != nil {
			log.Printf("Error submitting new assignment: %s", err)
			continue
		}
	}

	return nil
}

func (g *Generator) submitNewAssignmentFromTrade(trade *event.Trade, percentChange float64, t Type) error {
	floatPrice, err := strconv.ParseFloat(trade.Price, 64)
	if err != nil {
		return fmt.Errorf("Failed to parse price into float: %s", err)
	}

	newPrice := floatPrice * ((100 + percentChange) / 100)
	// normalise to 0.01 precision for currencies
	newPrice = (math.Floor(newPrice*100 + 0.5)) / 100

	newAssignment := Assignment{
		Price:    fmt.Sprintf("%.2f", newPrice),
		Quantity: trade.Quantity,
	}

	return g.SubmitAssignment(newAssignment, t)
}

// SubmitAssignment submits an assignment of type t to a message queue
func (g *Generator) SubmitAssignment(a Assignment, t Type) error {
	assignmentBytes, err := json.Marshal(a)
	if err != nil {
		return fmt.Errorf("Failed to marshal new assignment: %s", err)
	}

	var topic string
	switch t {
	case Buy:
		topic = event.TopicBuyerAssignment
	case Sell:
		topic = event.TopicSellerAssignment
	default:
		return fmt.Errorf("Unknown type of assignment given, expected BUY or SELL")
	}

	err = g.MessageQueue.Publish(assignmentBytes, topic)
	if err != nil {
		return fmt.Errorf("Failed to publish assignment: %s", err)
	}

	return nil
}
