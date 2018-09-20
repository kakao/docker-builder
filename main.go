package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

var handler http.Handler

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	r := mux.NewRouter()
	r.HandleFunc("/", handleIndex).Methods("GET")
	r.HandleFunc("/build", handleBuild).Methods("POST")
	r.HandleFunc("/result", handleBuildResult).Methods("POST")

	handler = r
}

// Run start a server
func main() {
	http.Handle("/", handler)
	log.Printf("[INFO] server is starting at %s port\n", serverPort)
	if err := http.ListenAndServe(":"+serverPort, nil); err != nil {
		log.Fatalf("[ERROR] server is failed while starting at %s port\n", serverPort)
	}
}
