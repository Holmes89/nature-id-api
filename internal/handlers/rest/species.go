package rest

import (
	"github.com/gorilla/mux"
	"nature-id-api/internal"
	"net/http"
)

const speciesBaseURL = "/v1/species"

type speciesHandler struct {
	service internal.SpeciesFinder
}

func MakeV1SpeciesHandler(mr *mux.Router, service internal.SpeciesFinder) http.Handler {

	r := mr.PathPrefix(speciesBaseURL).Subrouter()

	h := &speciesHandler{
		service: service,
	}

	r.HandleFunc("/{name}", h.Find).Methods("GET")

	return r
}

func (h *speciesHandler) Find(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	speciesName := vars["name"]

	res, err := h.service.FindMetaData(speciesName)
	if err != nil {
		makeError(w, http.StatusBadRequest, "unable to find results", "get")
		return
	}

	encodeResponse(r.Context(), w, res)
}

