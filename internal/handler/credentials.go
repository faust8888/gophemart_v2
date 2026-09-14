package handler

import (
	"encoding/json"
	"net/http"
)

type credentialsRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func decodeCredentials(w http.ResponseWriter, r *http.Request) (credentialsRequest, error) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

	var req credentialsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return credentialsRequest{}, err
	}
	return req, nil
}
