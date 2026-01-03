package main

import "net/http"

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, req *http.Request) {
	if cfg.platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	err := cfg.db.Reset(req.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to reset the database: ", err)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("All Users dropped."))
}
