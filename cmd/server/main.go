package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/martskins/zipsrv/cmd/server/handler"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/zip/create", handler.HandleZipRequest).Methods("POST")
	r.HandleFunc("/zip/download/{tkn}", handler.HandleGetZip).Methods("GET")
	http.Handle("/", r)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
