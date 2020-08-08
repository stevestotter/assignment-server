package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"regexp"
	"stevestotter/assignment-server/event"

	validator "github.com/go-playground/validator/v10"
	"github.com/julienschmidt/httprouter"
)

var validate *validator.Validate

type API struct {
	Port         string
	MessageQueue event.Publisher

	server *http.Server
}

// Start initialises and runs the API in a separate goroutine (non-blocking)
func (a *API) Start() error {
	validate = validator.New()
	validate.RegisterValidation("currency", validateCurrency)

	router := httprouter.New()
	router.POST("/buy", a.buyHandler)
	router.POST("/sell", a.sellHandler)

	a.server = &http.Server{Addr: fmt.Sprintf(":%s", a.Port), Handler: router}

	ln, err := net.Listen("tcp", a.server.Addr)
	if err != nil {
		return err
	}

	go func() {
		// TODO: Change logger to critical
		log.Printf("Server stopped: %s", a.server.Serve(ln))
	}()

	return nil
}

func validateCurrency(fl validator.FieldLevel) bool {
	res, _ := regexp.MatchString(`^([0-9])*\.([0-9]{2})$`, fl.Field().String())
	return res
}

type Assignment struct {
	Price    string `json:"price" validate:"required,currency"`
	Quantity string `json:"quantity" validate:"required,numeric,excludes=-"`
}

func (a *API) buyHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	a.assignmentHandler(event.TopicBuyerAssignment, w, r)
}

func (a *API) sellHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	a.assignmentHandler(event.TopicSellerAssignment, w, r)
}

func (a *API) assignmentHandler(eventTopic string, w http.ResponseWriter, r *http.Request) {
	assignment := &Assignment{}
	body := json.NewDecoder(r.Body)
	if err := body.Decode(&assignment); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := validate.Struct(assignment); err != nil {
		// TODO: Change logger
		log.Printf("Validation failed on assignment: %s\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	bytes, err := json.Marshal(assignment)
	if err != nil {
		apiErr := ErrorServer("Marshalling assignment to JSON failed")
		log.Printf("%s: %s\n", apiErr.Detail, err)
		apiErr.WriteJSON(w)
		return
	}

	if err := a.MessageQueue.Publish(bytes, eventTopic); err != nil {
		apiErr := ErrorServer("Failed to publish assignment to message queue")
		log.Printf("%s: %s\n", apiErr.Detail, err)
		apiErr.WriteJSON(w)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
