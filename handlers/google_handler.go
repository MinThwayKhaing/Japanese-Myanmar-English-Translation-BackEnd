package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GoogleAuthRequest struct {
	IdToken string `json:"idToken"`
}

func (h *UserHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	log.Println("[DEBUG] GoogleLogin endpoint called")
	var req GoogleAuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.IdToken == "" {
		log.Println("[ERROR] GoogleLogin: invalid request body")
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	log.Println("[DEBUG] GoogleLogin: token length:", len(req.IdToken))

	token, user, err := h.service.GoogleLogin(r.Context(), req.IdToken)
	if err != nil {
		log.Println("[ERROR] GoogleLogin failed:", err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	if user == nil {
		log.Println("[ERROR] GoogleLogin: user is nil")
		http.Error(w, "Login failed", http.StatusInternalServerError)
		return
	}

	log.Println("[DEBUG] GoogleLogin successful for:", user.Email)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token": token,
		"user":  user,
	})
}

func (h *UserHandler) GoogleRegister(w http.ResponseWriter, r *http.Request) {
	log.Println("[DEBUG] GoogleRegister endpoint called")
	var req GoogleAuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.IdToken == "" {
		log.Println("[ERROR] GoogleRegister: invalid request body")
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	log.Println("[DEBUG] GoogleRegister: token length:", len(req.IdToken))

	token, user, err := h.service.GoogleRegister(r.Context(), req.IdToken)
	if err != nil {
		log.Println("[ERROR] GoogleRegister failed:", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if user == nil {
		log.Println("[ERROR] GoogleRegister: user is nil")
		http.Error(w, "Registration failed", http.StatusInternalServerError)
		return
	}

	log.Println("[DEBUG] GoogleRegister successful for:", user.Email)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token": token,
		"user":  user,
	})
}

func (h *UserHandler) LinkGoogle(w http.ResponseWriter, r *http.Request, userID primitive.ObjectID) {
	log.Println("[DEBUG] LinkGoogle endpoint called for userID:", userID.Hex())
	var req GoogleAuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.IdToken == "" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	updatedUser, err := h.service.LinkGoogle(r.Context(), userID, req.IdToken)
	if err != nil {
		log.Println("[ERROR] LinkGoogle failed:", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": err.Error()})
		return
	}

	log.Println("[DEBUG] LinkGoogle successful for:", updatedUser.Email)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Account linked with Google successfully",
		"user":    updatedUser,
	})
}
