package rest

import (
	"context"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"nature-id-api/internal"
	"net/http"
	"sort"
	"strings"
)

// BaseURL for http endpoint
const baseURL = "/v1/predict"

type handler struct {
	service internal.Predictor
}

func MakeV1Handler(mr *mux.Router, service internal.Predictor) http.Handler {

	r := mr.PathPrefix(baseURL).Subrouter()

	h := &handler{
		service: service,
	}

	r.HandleFunc("/", h.Predict).Methods("POST")

	return r
}

func (h *handler) Predict(w http.ResponseWriter, r *http.Request) {

	//TODO check extenstion (jpg)
	logrus.Info("received prediction request")
	file, _, err := r.FormFile("file")
	if err != nil {
		h.makeError(w, http.StatusBadRequest, "Unable to parse form: "+err.Error(), "create")
		return
	}
	if file == nil {
		h.makeError(w, http.StatusBadRequest, "File missing from form", "create")
		return
	}
	logrus.Info("starting prediction")
	labels, err := h.service.Predict(file)
	if err != nil {
		h.makeError(w, http.StatusInternalServerError, err.Error(), "predict")
		return
	}
	logrus.Info("prediction complete")

	sort.Sort(labels)
	w.WriteHeader(http.StatusCreated)
	h.encodeResponse(r.Context(), w, labels)
}

func (h *handler) makeError(w http.ResponseWriter, code int, message string, method string) {
	logrus.WithFields(
		logrus.Fields{
			"type":   code,
			"method": method,
		}).Error(strings.ToLower(message))
	http.Error(w, message, code)
}

func (h *handler) encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response) //TODO check error and handle?
}
