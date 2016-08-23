package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/tidwall/buntdb"
)

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/{database}/{key}", get).Methods("GET")
	router.HandleFunc("/{database}/{key}", set).Methods("POST")
	router.HandleFunc("/{database}", create).Methods("PUT")

	log.Fatal(http.ListenAndServe(":8080", router))
}

func isJSONString(s string) bool {
	var js string
	return json.Unmarshal([]byte(s), &js) == nil

}

func isJSON(s string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}

func open(database string) (db *buntdb.DB, err error) {
	db, err = buntdb.Open(database + ".db")
	return
}

func writeError(w http.ResponseWriter, err error) {
	http.Error(w, fmt.Sprintf("%s", err), 400)
}

func writeString(w http.ResponseWriter, value *interface{}) {
	fmt.Fprintln(w, value)
}

func create(w http.ResponseWriter, r *http.Request) {
	log.Println("Responsing to CREATE request")
	log.Println(r.UserAgent())

	vars := mux.Vars(r)
	database := vars["database"]

	db, _ := open(database)
	defer db.Close()
	return
}

func set(w http.ResponseWriter, r *http.Request) {
	log.Println("Responsing to SET request")
	log.Println(r.UserAgent())

	vars := mux.Vars(r)
	database := vars["database"]
	key := vars["key"]

	db, _ := open(database)
	defer db.Close()

	err := db.Update(func(tx *buntdb.Tx) error {
		value, _ := ioutil.ReadAll(r.Body)
		_, _, err := tx.Set(key, string(value), nil)
		if err != nil {
			return err
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "completed")
		return nil
	})

	if err != nil {
		writeError(w, err)
	}
}

func get(w http.ResponseWriter, r *http.Request) {
	log.Println("Responsing to GET request")
	log.Println(r.UserAgent())

	vars := mux.Vars(r)
	database := vars["database"]
	key := vars["key"]

	db, _ := open(database)
	defer db.Close()

	err := db.View(func(tx *buntdb.Tx) error {

		value, err := tx.Get(key)
		if err != nil {
			return err
		}

		contentType := r.Header.Get("Content-Type")
		log.Println(fmt.Sprintf("content-type:%s", contentType))

		// application/json
		if contentType == "application/json" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, value)
			return nil
		}

		// auto-detect json
		if contentType == "" {
			if isJSONString(value) || isJSON(value) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, value)
				return nil
			}
		}

		fmt.Fprintln(w, value)
		return nil
	})

	if err != nil {
		writeError(w, err)
	}
}
