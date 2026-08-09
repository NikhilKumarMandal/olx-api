package main

import (
	"log"
	"net/http"
)

func main(){

	http.HandleFunc("GET /healthz", func(w http.ResponseWriter,r  *http.Request){
		w.Header().Set("Content-Type","application/json")

		w.Write([]byte("all ok"))
	})


	err := http.ListenAndServe(":8090",nil)
	if err != nil {
		log.Fatalf("server failed: %v",err)
	}
}