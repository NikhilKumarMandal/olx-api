package handlers

import (
	"database/sql"
	"net/http"
)

func List(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request){

		w.Write([]byte("all okay"))
	}
}