// handlers/subscription_handler.go
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"USDT_BackEnd/models"
	"USDT_BackEnd/services"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SubscriptionHandler struct {
	service *services.SubscriptionService
}

func NewSubscriptionHandler(s *services.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{service: s}
}

// POST /api/subscriptions
func (h *SubscriptionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var sub models.Subscription
	if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	if err := h.service.Create(r.Context(), sub); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "created"})
}

// DELETE /api/subscriptions/{id}
func (h *SubscriptionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "deleted"})
}

// GET /api/subscriptions/{id}
func (h *SubscriptionHandler) GetOne(w http.ResponseWriter, r *http.Request) {
	id, _ := primitive.ObjectIDFromHex(r.PathValue("id"))

	sub, err := h.service.GetOne(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(sub)
}

// GET /api/subscriptions
func (h *SubscriptionHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	subs, err := h.service.GetAll(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(subs)
}

// GET /api/subscriptions/paginated?search=&page=1&limit=10
func (h *SubscriptionHandler) GetPaginated(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	data, total, err := h.service.GetPaginated(
		r.Context(),
		search,
		page,
		limit,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":  data,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}
