package main

import (
	"encoding/json"
	"net/http"
)

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithIndentJSON(w, code, map[string]string{"error": message})
}

func respondWithIndentJSON(w http.ResponseWriter, code int, payload interface{}) {
	data, err := json.MarshalIndent(payload, "", "	")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	w.Write(data)
}
