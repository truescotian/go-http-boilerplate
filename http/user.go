package http

import (
	"encoding/json"
	"net/http"

	app "github.com/truescotian/golang-rest-projstructure"

	"github.com/gorilla/mux"
)

func (s *Server) registerUserRoutes(r *mux.Router) {
	r.HandleFunc("/users", s.handleUserCreate).Methods("POST")
}

func (s *Server) handleUserCreate(w http.ResponseWriter, r *http.Request) {
	var user app.User

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err := s.UserService.CreateUser(r.Context(), &user)
	if err != nil {
		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
		return
	}

	if err := json.NewEncoder(w).Encode(user); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
}
