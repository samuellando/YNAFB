package handler

import (
	"net/http"
	"io"
	"encoding/json"
	"samuellando.com/YNAFB/data"
	"log"
)

type Login struct {
	Queires *data.Queries
}

func (b Login) CreateLogin(w http.ResponseWriter, req *http.Request) {
	log.Println("Create login")
	// Parse the input
	params := data.CreateLoginParams{}
	data, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(data, &params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Query the db
	budget, err := b.Queires.CreateLogin(req.Context(), params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Send response
	res, err := json.Marshal(budget)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(res)
}
