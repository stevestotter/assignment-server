package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"stevestotter/assignment-server/event"

	validator "github.com/go-playground/validator/v10"
	"github.com/julienschmidt/httprouter"
)

const (
	buyAssignmentType  = "buy"
	sellAssignmentType = "sell"
)

var validate *validator.Validate

type API struct {
	Port         string
	MessageQueue event.Publisher

	server *http.Server
}

// Start initialises and runs the API
func (a *API) Start() {
	validate = validator.New()
	validate.RegisterValidation("currency", validateCurrency)

	router := httprouter.New()
	router.POST("/buy", a.buyHandler)
	router.POST("/sell", a.sellHandler)

	a.server = &http.Server{Addr: fmt.Sprintf(":%s", a.Port), Handler: router}

	// TODO: Bring in logger to make this a critical level
	log.Printf("Server stopped: %s", a.server.ListenAndServe())
}

func validateCurrency(fl validator.FieldLevel) bool {
	res, _ := regexp.MatchString(`^([0-9])*\.([0-9]{2})$`, fl.Field().String())
	return res
}

type Assignment struct {
	Price    string `json:"price" validate:"required,currency"`
	Quantity string `json:"quantity" validate:"required,numeric,excludes=-"`
	Type     string `json:"type" validate:"isdefault"`
}

func (a *API) buyHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	a.assignmentHandler(buyAssignmentType, w, r)
}

func (a *API) sellHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	a.assignmentHandler(sellAssignmentType, w, r)
}

func (a *API) assignmentHandler(assignmentType string, w http.ResponseWriter, r *http.Request) {
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

	assignment.Type = assignmentType

	bytes, err := json.Marshal(assignment)
	if err != nil {
		apiErr := ErrorServer("Marshalling assignment to JSON failed")
		log.Printf("%s: %s\n", apiErr.Detail, err)
		apiErr.WriteJSON(w)
		return
	}

	if err := a.MessageQueue.Publish(bytes); err != nil {
		apiErr := ErrorServer("Failed to publish assignment to message queue")
		log.Printf("%s: %s\n", apiErr.Detail, err)
		apiErr.WriteJSON(w)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
