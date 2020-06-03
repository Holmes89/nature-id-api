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
const predictBaseURL = "/v1/predict"

type predictHandler struct {
	service internal.Predictor
}

func MakeV1PredictHandler(mr *mux.Router, service internal.Predictor) http.Handler {

	r := mr.PathPrefix(predictBaseURL).Subrouter()

	h := &predictHandler{
		service: service,
	}

	r.HandleFunc("/", h.Predict).Methods("POST")

	return r
}

func (h *predictHandler) Predict(w http.ResponseWriter, r *http.Request) {

	//TODO check extension (jpg)
	logrus.Info("received prediction request")
	file, _, err := r.FormFile("file")
	if err != nil {
		makeError(w, http.StatusBadRequest, "Unable to parse form: "+err.Error(), "create")
		return
	}
	if file == nil {
		makeError(w, http.StatusBadRequest, "File missing from form", "create")
		return
	}
	logrus.Info("starting prediction")
	labels, err := h.service.Predict(file)
	if err != nil {
		makeError(w, http.StatusInternalServerError, err.Error(), "predict")
		return
	}
	logrus.Info("prediction complete")

	sort.Sort(labels)
	w.WriteHeader(http.StatusCreated)
	encodeResponse(r.Context(), w, labels)
}

func makeError(w http.ResponseWriter, code int, message string, method string) {
	logrus.WithFields(
		logrus.Fields{
			"type":   code,
			"method": method,
		}).Error(strings.ToLower(message))
	http.Error(w, message, code)
}

func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response) //TODO check error and handle?
}
