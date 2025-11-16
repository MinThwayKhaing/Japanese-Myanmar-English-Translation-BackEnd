package handlers

import (
	"encoding/json"
	"net/http"

	"USDT_BackEnd/services"
)

type SubscriptionHandler struct {
	service *services.SubscriptionService
}

func NewSubscriptionHandler(service *services.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{service: service}
}

// GET /api/subscriptions/prices
func (h *SubscriptionHandler) GetPrices(w http.ResponseWriter, r *http.Request) {
	plans, err := h.service.GetPrices(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(plans)
}

// PUT /api/subscriptions/prices
func (h *SubscriptionHandler) UpdatePrices(w http.ResponseWriter, r *http.Request) {
	type reqBody struct {
		MonthlyPrice float64 `json:"monthlyPrice"`
		YearlyPrice  float64 `json:"yearlyPrice"`
		FreeMonths   int     `json:"freeMonths"`
	}
	var req reqBody
	json.NewDecoder(r.Body).Decode(&req)

	if err := h.service.UpdatePrices(r.Context(), req.MonthlyPrice, req.YearlyPrice, req.FreeMonths); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"message": "Subscription prices updated"})
}
