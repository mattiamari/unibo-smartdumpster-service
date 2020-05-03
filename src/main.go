package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
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
	ID        string `db:"id" json:"id"`
	Name      string `db:"name" json:"name"`
	Available bool   `db:"available" json:"available"`
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

var db *sqlx.DB

func handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Smart Dumpster Service"))
}

func handleDumpsters(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dumpsters := []Dumpster{}
	err := db.Select(&dumpsters, "select * from dumpster")
	if err != nil {
		log.Print(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	json, _ := json.Marshal(dumpsters)
	w.WriteHeader(http.StatusOK)
	w.Write(json)
}

func handleDumpster(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"dumpster": {"id": "%s"}}`, params["dumpsterId"])))
}

func main() {
	log.Println("hi! :)")

	dbConnStr := "user=smartdumpster password=smartdumpster dbname=smartdumpster host=localhost sslmode=disable"
	var err error
	db, err = sqlx.Connect("postgres", dbConnStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := mux.NewRouter()
	r.HandleFunc("/", handleRoot).Methods(http.MethodGet)

	api := r.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/", handleRoot).Methods(http.MethodGet)
	api.HandleFunc("/dumpsters", handleDumpsters).Methods(http.MethodGet)
	api.HandleFunc("/dumpster/{dumpsterId}", handleDumpster).Methods(http.MethodGet)

	log.Fatal(http.ListenAndServe(":8080", r))
}
