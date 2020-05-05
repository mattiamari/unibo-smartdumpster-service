package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/ssh"
)

/*
	GET /dumpsters : dumpsters list, with availability, dumps count since last emptied and weight
	GET /dumpster/<uuid> : dumpster info + some stats
	GET /dumpster/<uuid>/token : generate and return new token
	POST /dumpster/<uuid>/availability : change availability
	POST /dumpster/<uuid>/dump : register dump
	POST /dumpster/<uuid>/weight : register weight change
*/

type DumpType string

const (
	Paper    DumpType = "paper"
	Plastic           = "plastic"
	Unsorted          = "unsorted"
)

type Dumpster struct {
	ID          string `db:"id" json:"id"`
	Name        string `db:"name" json:"name"`
	Available   bool   `db:"available" json:"available"`
	WeightLimit int    `db:"weight_limit" json:"weight_limit"`
}

type Dump struct {
	DumpsterID string    `db:"id_dumpster" json:"dumpster_id"`
	UserID     string    `db:"id_user" json:"user_id"`
	Type       DumpType  `db:"type" json:"dump_type"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}

type Weight struct {
	DumpsterID string    `db:"id_dumpster" json:"dumpster_id"`
	Weight     int       `db:"weight" json:"weight"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}

type Availability struct {
	Available bool `json:"available"`
}

var db *sqlx.DB
var signer ssh.Signer

func main() {
	log.Println("hi! :)")
	var err error

	dbConnStr := "user=smartdumpster password=smartdumpster dbname=smartdumpster host=localhost sslmode=disable"
	db, err = sqlx.Connect("postgres", dbConnStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	key, err := ioutil.ReadFile("test-key")
	if err != nil {
		log.Fatal("Unable to read key file")
	}

	signer, err = ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatal("Unable to parse private key")
	}

	r := mux.NewRouter()
	r.HandleFunc("/", RootHandler).Methods(http.MethodGet)

	api := r.PathPrefix("/api/v1").Subrouter()
	api.Use(HeadersMiddleware)
	api.HandleFunc("/", RootHandler).Methods(http.MethodGet)
	api.HandleFunc("/dumpsters", DumpstersHandler).Methods(http.MethodGet)
	api.HandleFunc("/dumpster/{dumpsterId}", DumpsterHandler).Methods(http.MethodGet)
	api.HandleFunc("/dumpster/{dumpsterId}/availability", AvailabilityHandler).Methods(http.MethodGet)
	api.HandleFunc("/dumpster/{dumpsterId}/availability", AvailabilityUpdateHandler).Methods(http.MethodPost).Headers("Content-Type", "application/json")
	api.HandleFunc("/dumpster/{dumpsterId}/dump", DumpHandler).Methods(http.MethodPost).Headers("Content-Type", "application/json")
	api.HandleFunc("/dumpster/{dumpsterId}/weight", WeightHandler).Methods(http.MethodPost).Headers("Content-Type", "application/json")
	api.HandleFunc("/dumpster/{dumpsterId}/token", TokenHandler).Methods(http.MethodGet)

	api.Use(mux.CORSMethodMiddleware(api))

	log.Fatal(http.ListenAndServe(":8080", r))
}