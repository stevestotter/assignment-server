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

var validate *validator.Validate

type Server struct {
	Port         string
	MessageQueue event.Publisher
}

// Start initialises and runs the API
func (s *Server) Start() {
	validate = validator.New()
	validate.RegisterValidation("currency", validateCurrency)

	router := httprouter.New()
	router.POST("/buy", s.buyHandler)
	router.POST("/sell", s.sellHandler)

	// TODO: change from Fatal? As shouldn't bring down entire assignment server...
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", s.Port), router))
}

func validateCurrency(fl validator.FieldLevel) bool {
	res, _ := regexp.MatchString(`^([0-9])*\.([0-9]{2})$`, fl.Field().String())
	fmt.Printf("%s", fl.Field().String())
	return res
}

type Assignment struct {
	Price    string `json:"price" validate:"required,currency"`
	Quantity string `json:"quantity" validate:"required,numeric,excludes=-"`
}

func (s *Server) buyHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	assignment := &Assignment{}
	body := json.NewDecoder(r.Body)
	if err := body.Decode(&assignment); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := validate.Struct(assignment); err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

}

func (s *Server) sellHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

}
