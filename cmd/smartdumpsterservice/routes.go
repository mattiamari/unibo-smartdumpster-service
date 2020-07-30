package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func RootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Smart Dumpster Service"))
}

func DumpstersHandler(w http.ResponseWriter, r *http.Request) {
	dumpsters := []Dumpster{}
	err := db.Select(&dumpsters, getDumpstersQuery(""))
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

	dumpster := DumpsterDetails{}
	err := db.Get(&dumpster, getDumpstersQuery("and d.id = $1"), params["dumpsterId"])
	if err != nil {
		log.Print(err.Error())
		w.WriteHeader(http.StatusNotFound)
		return
	}

	dumps, err := getDumps(dumpster.ID, 30)
	if err != nil {
		log.Print(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	weights, err := getWeights(dumpster.ID, 30)
	if err != nil {
		log.Print(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	dumpster.DumpHistory = dumps
	dumpster.WeightHistory = weights

	json, _ := json.Marshal(map[string]interface{}{"dumpster": dumpster})
	w.WriteHeader(http.StatusOK)
	w.Write(json)
}

func AvailabilityHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	availability, err := isAvailable(params["dumpsterId"])
	if err != nil {
		log.Print(err.Error())
		w.WriteHeader(http.StatusNotFound)
		return
	}

	json, _ := json.Marshal(map[string]interface{}{"available": availability})
	w.WriteHeader(http.StatusOK)
	w.Write(json)
}

func AvailabilityUpdateHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	availability := Availability{}
	json.NewDecoder(r.Body).Decode(&availability)

	_, err := db.Exec("update dumpster set available=$1 where id=$2", availability.Available, params["dumpsterId"])

	if err != nil {
		log.Print(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func DumpHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	dump := Dump{}
	json.NewDecoder(r.Body).Decode(&dump)

	_, err := db.Exec("insert into dump (id_user, id_dumpster, type) values ($1, $2, $3)", dump.UserID, params["dumpsterId"], dump.Type)

	if err != nil {
		log.Print(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func WeightHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	weight := Weight{}
	json.NewDecoder(r.Body).Decode(&weight)

	_, err := db.Exec("insert into weight (id_dumpster, weight) values ($1, $2)", params["dumpsterId"], weight.Weight)

	if err != nil {
		log.Print(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func TokenHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	available, err := isAvailable(params["dumpsterId"])
	if err != nil {
		log.Print(err.Error())
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if !available {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// generate token, which is "<dumpster_id>_<timestamp>"
	token := fmt.Sprintf("%s_%d", params["dumpsterId"], time.Now().Unix())

	// generate signature from token and private key
	signature, _ := signer.Sign(nil, []byte(token))
	encodedSig := base64.StdEncoding.EncodeToString(signature.Blob)

	// append base64 encoded signature to the token
	token = fmt.Sprintf("%s_%s", token, encodedSig)

	json, _ := json.Marshal(map[string]string{"token": token})
	w.WriteHeader(http.StatusOK)
	w.Write(json)
}

func HeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

func isAvailable(dumpsterID string) (bool, error) {
	var availability bool
	err := db.Get(&availability, "select available from dumpster where id=$1", dumpsterID)
	if err != nil {
		return false, err
	}

	return availability, nil
}

func getCurrentWeight(dumpsterID string) (int, error) {
	var weight int
	err := db.Get(&weight, "select weight from weight where id_dumpster=$1 order by created_at desc limit 1")
	if err != nil {
		return -1, err
	}

	return weight, nil
}

func getDumps(dumpsterID string, days int) ([]Dump, error) {
	dumps := make([]Dump, 0)
	err := db.Select(&dumps,
		"select * from dump where id_dumpster=$1 and created_at >= $2 order by created_at",
		dumpsterID,
		time.Now().AddDate(0, 0, -1*days))
	if err != nil {
		return nil, err
	}

	return dumps, nil
}

func getWeights(dumpsterID string, days int) ([]Weight, error) {
	weights := make([]Weight, 0)
	err := db.Select(&weights,
		`select t1.id_dumpster, t1.weight, t1.created_at, count(d) as dumps_since_last_emptied
		from (
		select w.*, max(w2.created_at) as last_emptied from weight w
		left join weight w2 on w.id_dumpster = w.id_dumpster and w2.weight = 0 and w2.created_at <= w.created_at
		where w.id_dumpster = $1 and w.created_at >= $2
		group by w.id_dumpster, w.weight, w.created_at
		) t1
		
		left join dump d on d.id_dumpster = t1.id_dumpster and d.created_at between t1.last_emptied and t1.created_at
		group by t1.id_dumpster, t1.weight, t1.created_at, t1.last_emptied
		order by t1.created_at`,
		dumpsterID, time.Now().AddDate(0, 0, -1*days))
	if err != nil {
		return nil, err
	}

	return weights, nil
}

func getDumpstersQuery(filter string) string {
	return fmt.Sprintf(
		`select d.*, current_weight, last_emptied, count(du) as dumps_since_last_emptied from dumpster d

		-- current weight
		left join (
		select w.id_dumpster, w.weight as current_weight from weight w
		join (select id_dumpster, max(created_at) as last_weight from weight
		group by id_dumpster) w2 on w2.id_dumpster = w.id_dumpster
		where w.created_at = w2.last_weight) as cw on cw.id_dumpster = d.id
		
		-- date last emptied
		left join (
		select id_dumpster, max(created_at) as last_emptied from weight
		where weight = 0
		group by id_dumpster) as le on le.id_dumpster = d.id
		
		-- dumps since last emptied
		left join dump du on du.id_dumpster = d.id and du.created_at >= last_emptied
		
		where 1=1 %s
		group by d.id, current_weight, last_emptied
		order by d.name`, filter)
}
