package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"USDT_BackEnd/services"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserHandler struct {
	service *services.UserService
}

func NewUserHandler(service *services.UserService) *UserHandler {
	return &UserHandler{service: service}
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

// ------------------- Register -------------------
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	log.Println("[DEBUG] Register endpoint called")
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("[ERROR] Failed to decode Register request:", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	user, err := h.service.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		log.Println("[ERROR] Register failed:", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Println("[DEBUG] User registered successfully:", req.Email)
	json.NewEncoder(w).Encode(user)
}

// ------------------- Login -------------------
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	log.Println("[DEBUG] Login endpoint called")
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("[ERROR] Failed to decode Login request:", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	token, user, err := h.service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		log.Println("[ERROR] Login failed for email:", req.Email, "error:", err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	log.Println("[DEBUG] Login successful for email:", req.Email)
	response := map[string]interface{}{
		"token":        token,
		"role":         user.Role,
		"subscription": user.Subscription,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ------------------- Change Password -------------------
func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request, userID primitive.ObjectID) {
	log.Println("[DEBUG] ChangePassword endpoint called for userID:", userID.Hex())
	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("[ERROR] Failed to decode ChangePassword request:", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.ChangePassword(r.Context(), userID, req.CurrentPassword, req.NewPassword); err != nil {
		log.Println("[ERROR] ChangePassword failed for userID:", userID.Hex(), "error:", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Println("[DEBUG] Password updated successfully for userID:", userID.Hex())
	json.NewEncoder(w).Encode(map[string]string{"message": "Password updated successfully"})
}

// ------------------- Get Favorites -------------------
func (h *UserHandler) GetFavorites(w http.ResponseWriter, r *http.Request, userID primitive.ObjectID) {
	log.Println("[DEBUG] GetFavorites endpoint called for userID:", userID.Hex())

	favorites, err := h.service.GetFavoritesByUserID(r.Context(), userID)
	if err != nil {
		log.Println("[ERROR] GetFavorites failed for userID:", userID.Hex(), "error:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println("[DEBUG] Favorites fetched for userID:", userID.Hex(), "count:", len(favorites))
	json.NewEncoder(w).Encode(favorites)
}

// ------------------- Save Favorite -------------------
func (h *UserHandler) SaveFavorite(w http.ResponseWriter, r *http.Request, userID primitive.ObjectID) {
	log.Println("[DEBUG] SaveFavorite endpoint called userID:", userID.Hex())

	var data struct {
		WordID string `json:"wordId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil || data.WordID == "" {
		log.Println("[ERROR] Invalid SaveFavorite request:", err)
		http.Error(w, "Invalid word ID", http.StatusBadRequest)
		return
	}

	wordObjID, err := primitive.ObjectIDFromHex(data.WordID)
	if err != nil {
		log.Println("[ERROR] Invalid word ID format:", data.WordID)
		http.Error(w, "Invalid word ID format", http.StatusBadRequest)
		return
	}

	if err := h.service.SaveFavorite(r.Context(), userID, wordObjID); err != nil {
		log.Println("[ERROR] SaveFavorite failed for userID:", userID.Hex(), "wordID:", data.WordID, "error:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println("[DEBUG] Word added to favorites userID:", userID.Hex(), "wordID:", data.WordID)
	json.NewEncoder(w).Encode(map[string]string{"message": "Added to favorites"})
}

// ------------------- Remove Favorite -------------------
func (h *UserHandler) RemoveFavorite(w http.ResponseWriter, r *http.Request, userID primitive.ObjectID) {
	log.Println("[DEBUG] RemoveFavorite endpoint called userID:", userID.Hex())

	var data struct {
		WordID string `json:"wordId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil || data.WordID == "" {
		log.Println("[ERROR] Invalid RemoveFavorite request:", err)
		http.Error(w, "Invalid word ID", http.StatusBadRequest)
		return
	}

	wordObjID, err := primitive.ObjectIDFromHex(data.WordID)
	if err != nil {
		log.Println("[ERROR] Invalid word ID format:", data.WordID)
		http.Error(w, "Invalid word ID format", http.StatusBadRequest)
		return
	}

	if err := h.service.RemoveFavorite(r.Context(), userID, wordObjID); err != nil {
		log.Println("[ERROR] RemoveFavorite failed for userID:", userID.Hex(), "wordID:", data.WordID, "error:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println("[DEBUG] Word removed from favorites userID:", userID.Hex(), "wordID:", data.WordID)
	json.NewEncoder(w).Encode(map[string]string{"message": "Removed from favorites"})
}

// ------------------- Get Favorites Paginated -------------------
func (h *UserHandler) GetFavoritesPaginated(w http.ResponseWriter, r *http.Request, userID primitive.ObjectID) {
	log.Println("[DEBUG] GetFavoritesPaginated endpoint called userID:", userID.Hex())

	page := 1
	limit := 10
	if p := r.URL.Query().Get("page"); p != "" {
		fmt.Sscanf(p, "%d", &page)
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}

	words, hasMore, err := h.service.GetFavoritesPaginated(r.Context(), userID, page, limit)
	if err != nil {
		log.Println("[ERROR] GetFavoritesPaginated failed for userID:", userID.Hex(), "error:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println("[DEBUG] Favorites paginated fetched userID:", userID.Hex(), "count:", len(words), "hasMore:", hasMore)
	response := map[string]interface{}{
		"favorites": words,
		"hasMore":   hasMore,
		"page":      page,
	}
	json.NewEncoder(w).Encode(response)
}

// ------------------- Get Subscribed Users (Admin) -------------------
func (h *UserHandler) GetSubscribedUsers(w http.ResponseWriter, r *http.Request) {
	log.Println("[DEBUG] GetSubscribedUsers endpoint called")

	users, err := h.service.GetSubscribedUsers(r.Context())
	if err != nil {
		log.Println("[ERROR] GetSubscribedUsers failed:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println("[DEBUG] Subscribed users fetched, count:", len(users))
	json.NewEncoder(w).Encode(users)
}
func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request, userID primitive.ObjectID) {
	user, err := h.service.GetUserByID(r.Context(), userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(user)
}
