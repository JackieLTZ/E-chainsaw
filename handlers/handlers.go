package handlers

import (
	"a/repos"
	"a/validators"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
)

type Handler struct {
	r *repos.UserRepository
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{r: repos.NewUserRepository(db)}
}

func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong\n"))
}

func (h *Handler) GetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.r.GetUsers(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := writeJSON(w, 200, users); err != nil {
		log.Printf("error writing JSON response: %v", err)
	}
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req validators.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	user, err := h.r.CreateUser(r.Context(), req.Name, req.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, user)
}
