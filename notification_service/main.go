package main

import (
	"github.com/gorilla/mux"
	"net/http"
)

func main () {
	router := mux.NewRouter()

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	}).Methods(http.MethodGet)

	http.ListenAndServe(":80", router)
}