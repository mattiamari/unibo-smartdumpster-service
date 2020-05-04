package main

import (
	"encoding/json"
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

var db *sqlx.DB

func RootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Smart Dumpster Service"))
}

func DumpstersHandler(w http.ResponseWriter, r *http.Request) {
	dumpsters := []Dumpster{}
	err := db.Select(&dumpsters, "select * from dumpster")
	if err != nil {
		log.Print(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	json, _ := json.Marshal(map[string]interface{}{"dumpsters": dumpsters})
	w.WriteHeader(http.StatusOK)
	w.Write(json)
}

func DumpsterHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	dumpster := Dumpster{}
	err := db.Get(&dumpster, "select * from dumpster where id=$1", params["dumpsterId"])
	if err != nil {
		log.Print(err.Error())
		w.WriteHeader(http.StatusNotFound)
		return
	}

	json, _ := json.Marshal(map[string]interface{}{"dumpster": dumpster})
	w.WriteHeader(http.StatusOK)
	w.Write(json)
}

func AvailabilityHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	if r.Method == http.MethodGet {
		var availability bool
		err := db.Get(&availability, "select available from dumpster where id=$1", params["dumpsterId"])
		if err != nil {
			log.Print(err.Error())
			w.WriteHeader(http.StatusNotFound)
			return
		}

		json, _ := json.Marshal(map[string]interface{}{"available": availability})
		w.WriteHeader(http.StatusOK)
		w.Write(json)
	}

	type Availability struct {
		Available bool `json:"available"`
	}

	if r.Method == http.MethodPost {
		availability := Availability{}
		decoder := json.NewDecoder(r.Body)
		decoder.Decode(&availability)

		db.Exec("update dumpster set available=$1 where id=$2", availability.Available, params["dumpsterId"])

		w.WriteHeader(http.StatusCreated)
	}
}

func HeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
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
	r.HandleFunc("/", RootHandler).Methods(http.MethodGet)

	api := r.PathPrefix("/api/v1").Subrouter()
	api.Use(HeadersMiddleware)
	api.HandleFunc("/", RootHandler).Methods(http.MethodGet)
	api.HandleFunc("/dumpsters", DumpstersHandler).Methods(http.MethodGet)
	api.HandleFunc("/dumpster/{dumpsterId}", DumpsterHandler).Methods(http.MethodGet)
	api.HandleFunc("/dumpster/{dumpsterId}/availability", AvailabilityHandler).Methods(http.MethodGet)
	api.HandleFunc("/dumpster/{dumpsterId}/availability", AvailabilityHandler).Methods(http.MethodPost).Headers("Content-Type", "application/json")
	api.Use(mux.CORSMethodMiddleware(api))

	log.Fatal(http.ListenAndServe(":8080", r))
}
