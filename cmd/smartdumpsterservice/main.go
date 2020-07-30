package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/handlers"
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
	ID                    string    `db:"id" json:"id"`
	Name                  string    `db:"name" json:"name"`
	Available             bool      `db:"available" json:"available"`
	WeightLimit           int       `db:"weight_limit" json:"weight_limit"`
	LastEmptied           time.Time `db:"last_emptied" json:"last_emptied"`
	DumpsSinceLastEmptied int       `db:"dumps_since_last_emptied" json:"dumps_since_last_emptied"`
	CurrentWeight         int       `db:"current_weight" json:"current_weight"`
}

type DumpsterDetails struct {
	ID                    string    `db:"id" json:"id"`
	Name                  string    `db:"name" json:"name"`
	Available             bool      `db:"available" json:"available"`
	WeightLimit           int       `db:"weight_limit" json:"weight_limit"`
	LastEmptied           time.Time `db:"last_emptied" json:"last_emptied"`
	DumpsSinceLastEmptied int       `db:"dumps_since_last_emptied" json:"dumps_since_last_emptied"`
	CurrentWeight         int       `db:"current_weight" json:"current_weight"`
	DumpHistory           []Dump    `json:"dump_history"`
	WeightHistory         []Weight  `json:"weight_history"`
}

type Dump struct {
	DumpsterID string    `db:"id_dumpster" json:"-"`
	UserID     string    `db:"id_user" json:"user_id"`
	Type       DumpType  `db:"type" json:"dump_type"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}

type Weight struct {
	DumpsterID            string    `db:"id_dumpster" json:"-"`
	Weight                int       `db:"weight" json:"weight"`
	DumpsSinceLastEmptied int       `db:"dumps_since_last_emptied" json:"dumps_since_last_emptied"`
	CreatedAt             time.Time `db:"created_at" json:"created_at"`
}

type Availability struct {
	Available bool `json:"available"`
}

var db *sqlx.DB
var signer ssh.Signer

func main() {
	log.Println("hi! :)")
	var err error

	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")

	dbConnStr := fmt.Sprintf("user=%s password=%s dbname=smartdumpster host=%s sslmode=disable", dbUser, dbPassword, dbHost)
	db, err = sqlx.Connect("postgres", dbConnStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	key, err := ioutil.ReadFile("./signkey")
	if err != nil {
		log.Fatal("Unable to read key file")
	}

	signer, err = ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatal("Unable to parse private key")
	}

	r := mux.NewRouter()
	r.HandleFunc("/", RootHandler).Methods(http.MethodGet)

	dashboard := http.StripPrefix("/dashboard", http.FileServer(http.Dir("./dashboard/")))

	r.PathPrefix("/dashboard").Handler(dashboard)
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

	cors := handlers.CORS(
		handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "HEAD", "OPTIONS"}),
		handlers.AllowedOrigins([]string{"*"}))

	log.Fatal(http.ListenAndServe(":8080", cors(r)))
}
