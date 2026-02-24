package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"

	"USDT_BackEnd/db"

	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

func (s *UserService) SendResetOTP(ctx context.Context, email string) error {
	log.Println("[DEBUG] SendResetOTP called for:", email)
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil || user == nil {
		log.Println("[ERROR] SendResetOTP: user not found:", email)
		return errors.New("No account found with this email")
	}
	if user.AuthProvider == "GOOGLE" {
		log.Println("[ERROR] SendResetOTP: google account:", email)
		return errors.New("This account uses Google Sign-In. Please reset your password through Google")
	}

	otp := fmt.Sprintf("%06d", rand.Intn(1000000))
	expiry := time.Now().Add(10 * time.Minute)
	log.Println("[DEBUG] SendResetOTP: generated OTP for:", email)

	db.Database.Collection("password_otps").DeleteMany(ctx, bson.M{"email": email})
	db.Database.Collection("password_otps").InsertOne(ctx, bson.M{
		"email":     email,
		"otp":       otp,
		"expiresAt": expiry,
	})

	SendEmail(email, "Password Reset OTP", "Your OTP: "+otp)
	log.Println("[DEBUG] SendResetOTP: OTP email sent to:", email)
	return nil
}

func (s *UserService) ResetPasswordWithOTP(ctx context.Context, email, otp, newPassword string) error {
	log.Println("[DEBUG] ResetPasswordWithOTP called for:", email)
	var record struct {
		OTP       string    `bson:"otp"`
		ExpiresAt time.Time `bson:"expiresAt"`
	}

	err := db.Database.Collection("password_otps").
		FindOne(ctx, bson.M{"email": email, "otp": otp}).Decode(&record)

	if err != nil {
		log.Println("[ERROR] ResetPasswordWithOTP: OTP not found for:", email)
		return errors.New("invalid or expired OTP")
	}
	if time.Now().After(record.ExpiresAt) {
		log.Println("[ERROR] ResetPasswordWithOTP: expired OTP for:", email)
		return errors.New("invalid or expired OTP")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Println("[ERROR] ResetPasswordWithOTP: failed to hash password:", err)
		return errors.New("failed to reset password")
	}
	_, err = db.Database.Collection("users").UpdateOne(
		ctx,
		bson.M{"email": email},
		bson.M{"$set": bson.M{"password": string(hash)}},
	)
	if err != nil {
		log.Println("[ERROR] ResetPasswordWithOTP: failed to update password:", err)
		return errors.New("failed to reset password")
	}

	db.Database.Collection("password_otps").DeleteMany(ctx, bson.M{"email": email})
	log.Println("[DEBUG] ResetPasswordWithOTP: password reset successful for:", email)
	return nil
}
