package handlers

import (
	"encoding/json"
	"log"
	"net/http"
)

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

type ResetPasswordRequest struct {
	Email       string `json:"email"`
	OTP         string `json:"otp"`
	NewPassword string `json:"newPassword"`
}

func (h *UserHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	log.Println("[DEBUG] ForgotPassword endpoint called")
	var req ForgotPasswordRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("[ERROR] ForgotPassword: invalid request body:", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		log.Println("[ERROR] ForgotPassword: email is empty")
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	log.Println("[DEBUG] ForgotPassword: sending OTP to:", req.Email)
	if err := h.service.SendResetOTP(r.Context(), req.Email); err != nil {
		log.Println("[ERROR] ForgotPassword: failed to send OTP:", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Println("[DEBUG] ForgotPassword: OTP sent successfully to:", req.Email)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "OTP sent to email",
	})
}

func (h *UserHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	log.Println("[DEBUG] ResetPassword endpoint called")
	var req ResetPasswordRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("[ERROR] ResetPassword: invalid request body:", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.OTP == "" || req.NewPassword == "" {
		log.Println("[ERROR] ResetPassword: missing required fields")
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	}

	log.Println("[DEBUG] ResetPassword: attempting reset for:", req.Email)
	if err := h.service.ResetPasswordWithOTP(
		r.Context(), req.Email, req.OTP, req.NewPassword,
	); err != nil {
		log.Println("[ERROR] ResetPassword: failed for:", req.Email, "error:", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Println("[DEBUG] ResetPassword: successful for:", req.Email)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Password reset successful",
	})
}
