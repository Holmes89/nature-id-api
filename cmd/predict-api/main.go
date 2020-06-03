package main

import (
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"nature-id-api/internal"
	"nature-id-api/internal/handlers/rest"
	"nature-id-api/internal/predictor"
	"nature-id-api/internal/storage"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

const (
	defaultPort = "8080"
)

func main() {

	//Create server
	port := ":" + internal.GetEnv("PORT", defaultPort)

	router := mux.NewRouter()

	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "PATCH", "OPTIONS"})
	cors := handlers.CORS(originsOk, headersOk, methodsOk)

	router.Use(cors, endpointLogging)

	bucket, err := storage.NewGCPBucketStorage(internal.LoadBucketConfig())
	if err != nil {
		logrus.WithError(err).Fatal("unable to create bucket")
	}
	defer bucket.Close()
	pred, err := predictor.NewTensorflowPredictor(bucket)
	if err != nil {
		logrus.WithField("err", err).Fatal("unable to start service")
	}

	rest.MakeV1Handler(router, pred)

	errs := make(chan error, 2) // This is used to handle and log the reason why the application quit.
	go func() {
		logrus.WithFields(
			logrus.Fields{
				"transport": "http",
				"port":      port,
			}).Info("server started")
		errs <- http.ListenAndServe(port, (cors)(router))
	}()
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	logrus.WithField("error", <-errs).Error("terminated")
}

func endpointLogging(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logrus.WithFields(logrus.Fields{"uri": r.URL.String(), "method": r.Method}).Info("endpoint")
		if r.Method == "OPTIONS" {
			return
		}

		h.ServeHTTP(w, r)
	})
}
