package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"regexp"

	"github.com/stevestotter/assignment-server/assignment"

	validator "github.com/go-playground/validator/v10"
	"github.com/julienschmidt/httprouter"
)

var validate *validator.Validate

// API routes and handles assignment requests
type API struct {
	Port                string
	AssignmentSubmitter assignment.Submitter

	server *http.Server
}

// Start initialises and runs the API in a separate goroutine (non-blocking)
func (api *API) Start() error {
	validate = validator.New()
	validate.RegisterValidation("currency", validateCurrency)

	router := httprouter.New()
	router.POST("/buy", api.buyHandler)
	router.POST("/sell", api.sellHandler)

	api.server = &http.Server{Addr: fmt.Sprintf(":%s", api.Port), Handler: router}

	ln, err := net.Listen("tcp", api.server.Addr)
	if err != nil {
		return err
	}

	go func() {
		// TODO: Change logger to critical
		log.Printf("Server stopped: %s", api.server.Serve(ln))
	}()

	return nil
}

func validateCurrency(fl validator.FieldLevel) bool {
	res, _ := regexp.MatchString(`^([0-9])*\.([0-9]{2})$`, fl.Field().String())
	return res
}

func (api *API) buyHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	api.assignmentHandler(w, r, assignment.Buy)
}

func (api *API) sellHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	api.assignmentHandler(w, r, assignment.Sell)
}

func (api *API) assignmentHandler(w http.ResponseWriter, r *http.Request, t assignment.Type) {
	assignment := assignment.Assignment{}
	body := json.NewDecoder(r.Body)
	if err := body.Decode(&assignment); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := validate.Struct(&assignment); err != nil {
		// TODO: Change logger
		log.Printf("Validation failed on assignment: %s\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err := api.AssignmentSubmitter.SubmitAssignment(assignment, t)
	if err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
