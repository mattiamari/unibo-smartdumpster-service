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

	dumpster := Dumpster{}
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

	// generate token, which is "<dumpster_id>|<timestamp>"
	token := fmt.Sprintf("%s|%d", params["dumpsterId"], time.Now().Unix())

	// generate signature from token and private key
	signature, _ := signer.Sign(nil, []byte(token))
	encodedSig := base64.StdEncoding.EncodeToString(signature.Blob)

	// append base64 encoded signature to the token
	token = fmt.Sprintf("%s|%s", token, encodedSig)

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
	var dumps []Dump
	err := db.Select(&dumps,
		"select * from dump where id_dumpster=$1 and created_at >= $2",
		dumpsterID,
		time.Now().AddDate(0, 0, -1*days))
	if err != nil {
		return nil, err
	}

	return dumps, nil
}

func getWeights(dumpsterID string, days int) ([]Weight, error) {
	var weights []Weight
	err := db.Select(&weights,
		"select * from weight where id_dumpster=$1 and created_at >= $2",
		dumpsterID,
		time.Now().AddDate(0, 0, -1*days))
	if err != nil {
		return nil, err
	}

	return weights, nil
}

func getDumpstersQuery(filter string) string {
	return fmt.Sprintf(
		`select ta.*,
		(select count(*) from dump
			where id_dumpster = ta.id and created_at >= ta.last_emptied) dumps_since_last_emptied
		from (
			select d.*, max(w.created_at) as last_emptied
			from dumpster d
			left join weight w on w.id_dumpster = d.id
			where weight = 0 %s
			group by d.id
			order by d.name) ta`, filter)
}
