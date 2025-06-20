package user

import (
	"encoding/json"
	"net/http"
	"regexp"
)

type CreateUserRequest struct {
	Username   string `json:"username"`
	ServerName string `json:"server_name"`
}

var validName = regexp.MustCompile(`^[a-zA-Z0-9.\-_]+$`)

func isValidName(name string) bool {
	return validName.MatchString(name)
}

func writeJSONError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "Invalid request", http.StatusBadRequest)
		return
	}
	if req.Username == "" || req.ServerName == "" {
		writeJSONError(w, "Username and server_name required", http.StatusBadRequest)
		return
	}
	if !isValidName(req.Username) {
		writeJSONError(w, "Invalid username: only letters, numbers, ., -, _ allowed", http.StatusBadRequest)
		return
	}
	if !isValidName(req.ServerName) {
		writeJSONError(w, "Invalid server_name: only letters, numbers, ., -, _ allowed", http.StatusBadRequest)
		return
	}
	if err := CreateHome(req.ServerName, req.Username); err != nil {
		writeJSONError(w, "Failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "User and directories created"})
}
