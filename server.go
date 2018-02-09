package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/kebabmane/notely-api/app"
	"github.com/kebabmane/notely-api/controller"
	"github.com/kebabmane/notely-api/middlewares"
	"github.com/kebabmane/notely-api/model"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/cors"
	"github.com/zbindenren/negroni-prometheus"
)

func main() {
	// load application configurations in not production

	if os.Getenv("ENV") == "PRODUCTION" {
		fmt.Println("your running in production, did you know that?")
	} else {
		fmt.Println("your running in dev/test, did you know that?")
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}

	// migrate the database
	model.Init()

	// create the logger
	logger := logrus.New()

	// CORS middleware setup
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedHeaders:   []string{"Access-Control-Allow-Origin", "Content-Type", "Origin", "Accept-Encoding", "Accept-Language", "Authorization"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "OPTIONS", "DELETE"},
		AllowCredentials: true,
	})

	// set up router
	r := mux.NewRouter().StrictSlash(true)

	r.HandleFunc("/health", controller.HealthFunction).Methods("GET")

	r.Handle("/metrics", prometheus.Handler())

	// s is a subrouter to handle question routes
	api := r.PathPrefix("/v1").Subrouter()

	// categories routes
	api.HandleFunc("/categories/", controller.FetchAllCategories).Methods("GET")
	api.HandleFunc("/categories/", controller.CreateCategory).Methods("POST")
	api.HandleFunc("/categories/{id}", controller.FetchSingleCategory).Methods("GET")
	api.HandleFunc("/categories/{id}", controller.UpdateCategory).Methods("PUT")

	// muxRouter uses Negroni handles the middleware for authorization
	muxRouter := http.NewServeMux()
	muxRouter.Handle("/", r)
	muxRouter.Handle("/api/", negroni.New(
		negroni.HandlerFunc(middlewares.CheckJWT()),
		negroni.Wrap(api),
	))

	// Negroni handles the middleware chaining with next
	n := negroni.Classic()

	m := negroniprometheus.NewMiddleware("notely-api")

	// Use promethus for service stuff
	n.Use(m)

	// Use CORS
	n.Use(c)

	// handle routes with the muxRouter
	n.UseHandler(muxRouter)

	// start the server
	address := fmt.Sprintf(":%v", os.Getenv("PORT"))
	logger.Infof("server %v is started at %v\n", app.Version, address)
	panic(http.ListenAndServe(address, handlers.RecoveryHandler()(n)))

}
