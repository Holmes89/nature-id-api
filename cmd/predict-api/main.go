package main

import (
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"nature-id-api/internal/connection"
	"nature-id-api/internal/handlers/rest"
	"nature-id-api/internal/predictor"
	"nature-id-api/internal/speciesfinder"
	"nature-id-api/internal/speciesfinder/cache"
	"nature-id-api/internal/speciesfinder/client/wiki"
	"nature-id-api/internal/speciesfinder/client/wolframalpha"
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
	port := ":" + GetEnv("PORT", defaultPort)

	router := mux.NewRouter()

	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "PATCH", "OPTIONS"})
	cors := handlers.CORS(originsOk, headersOk, methodsOk)

	router.Use(cors, endpointLogging)

	bucket, err := storage.NewGCPBucketStorage(storage.LoadBucketConfig())
	if err != nil {
		logrus.WithError(err).Fatal("unable to create bucket")
	}
	defer bucket.Close()

	redisConn := connection.NewRedisClientDefault()
	defer redisConn.Close()

	speciesCache := cache.NewRedisCache(redisConn)
	clients := []speciesfinder.Client{wolframalpha.NewClient(), wiki.NewClient()}
	speciesService := speciesfinder.NewSpeciesFinderService(speciesCache, clients)

	modelConfig := predictor.LoadModelConfig()
	pred, err := predictor.NewTensorflowPredictor(bucket, modelConfig.GetModelPath(), modelConfig.GetLabelFilePath())
	if err != nil {
		logrus.WithField("err", err).Fatal("unable to start service")
	}

	rest.MakeV1PredictHandler(router, pred)
	rest.MakeV1SpeciesHandler(router, speciesService)

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

func GetEnv(env, fallback string) string {
	e := os.Getenv(env)
	if e == "" {
		return fallback
	}
	return e
}