package services

import (
	"context"
	"errors"
	"log"
	"time"

	"USDT_BackEnd/models"

	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/api/idtoken"
)

func (s *UserService) validateGoogleToken(ctx context.Context, token string) (*idtoken.Payload, error) {
	clientIDs := s.config.GoogleClientIDs
	if len(clientIDs) == 0 {
		clientIDs = []string{s.config.GoogleClientID}
		if s.config.GoogleClientIDiOS != "" {
			clientIDs = append(clientIDs, s.config.GoogleClientIDiOS)
		}
		if s.config.GoogleClientIDAndroid != "" {
			clientIDs = append(clientIDs, s.config.GoogleClientIDAndroid)
		}
	}
	log.Println("[DEBUG] validateGoogleToken: trying", len(clientIDs), "client IDs")

	for _, cid := range clientIDs {
		payload, err := idtoken.Validate(ctx, token, cid)
		if err == nil {
			log.Println("[DEBUG] validateGoogleToken: matched client ID:", maskedClientID(cid))
			return payload, nil
		}
		log.Println("[DEBUG] validateGoogleToken: failed for client ID:", maskedClientID(cid), "error:", err)
	}
	log.Println("[ERROR] validateGoogleToken: no matching client ID found")
	return nil, errors.New("invalid google token")
}

func maskedClientID(cid string) string {
	if len(cid) <= 20 {
		return cid
	}
	return cid[:14] + "..." + cid[len(cid)-6:]
}

func (s *UserService) GoogleLogin(ctx context.Context, idToken string) (string, *models.User, error) {
	log.Println("[DEBUG] GoogleLogin service called")
	payload, err := s.validateGoogleToken(ctx, idToken)
	if err != nil {
		log.Println("[ERROR] GoogleLogin: token validation failed:", err)
		return "", nil, err
	}
	if payload == nil {
		log.Println("[ERROR] GoogleLogin: payload is nil")
		return "", nil, errors.New("invalid google token payload")
	}

	email, ok := payload.Claims["email"].(string)
	if !ok || email == "" {
		log.Println("[ERROR] GoogleLogin: email not found in token")
		return "", nil, errors.New("email not found in token")
	}
	log.Println("[DEBUG] GoogleLogin: email from token:", email)

	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil || user == nil {
		log.Println("[ERROR] GoogleLogin: account not found for:", email)
		return "", nil, errors.New("account does not exist, please register")
	}

	log.Println("[DEBUG] GoogleLogin: user found, generating JWT for:", email)
	return s.generateJWT(user)
}

func (s *UserService) GoogleRegister(ctx context.Context, idToken string) (string, *models.User, error) {
	log.Println("[DEBUG] GoogleRegister service called")
	payload, err := s.validateGoogleToken(ctx, idToken)
	if err != nil {
		log.Println("[ERROR] GoogleRegister: token validation failed:", err)
		return "", nil, err
	}
	if payload == nil {
		log.Println("[ERROR] GoogleRegister: payload is nil")
		return "", nil, errors.New("invalid google token payload")
	}

	email, ok := payload.Claims["email"].(string)
	if !ok || email == "" {
		log.Println("[ERROR] GoogleRegister: email not found in token")
		return "", nil, errors.New("email not found in token")
	}
	log.Println("[DEBUG] GoogleRegister: email from token:", email)

	emailVerified, _ := payload.Claims["email_verified"].(bool)
	if !emailVerified {
		log.Println("[ERROR] GoogleRegister: email not verified for:", email)
		return "", nil, errors.New("google email not verified")
	}

	existing, _ := s.repo.GetUserByEmail(ctx, email)
	if existing != nil {
		log.Println("[ERROR] GoogleRegister: account already exists for:", email)
		return "", nil, errors.New("account already exists, please login")
	}

	user := &models.User{
		ID:           primitive.NewObjectID(),
		Email:        email,
		GoogleID:     payload.Subject,
		AuthProvider: "GOOGLE",
		Role:         models.RoleUser,
		Subscription: models.UserSubscription{
			SearchesLeft: s.config.DefaultSearchesLeft,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		log.Println("[ERROR] GoogleRegister: failed to create user:", err)
		return "", nil, err
	}

	log.Println("[DEBUG] GoogleRegister: user created, generating JWT for:", email)
	return s.generateJWT(user)
}

// LinkGoogle links a LOCAL user account with Google.
// Validates the Google token, checks that the Google email isn't taken by another user,
// then updates the current user's email, authProvider, and googleId.
func (s *UserService) LinkGoogle(ctx context.Context, userID primitive.ObjectID, idToken string) (*models.User, error) {
	payload, err := s.validateGoogleToken(ctx, idToken)
	if err != nil {
		return nil, err
	}

	email, ok := payload.Claims["email"].(string)
	if !ok || email == "" {
		return nil, errors.New("email not found in Google token")
	}

	// Check current user
	currentUser, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}
	if currentUser.AuthProvider == "GOOGLE" {
		return nil, errors.New("account is already linked with Google")
	}

	// Check if Google email is already used by another account
	existing, _ := s.repo.GetUserByEmail(ctx, email)
	if existing != nil && existing.ID != userID {
		return nil, errors.New("this Google email is already registered with another account")
	}

	// Link
	if err := s.repo.LinkGoogle(ctx, userID, email, payload.Subject); err != nil {
		return nil, err
	}

	// Return updated user
	updatedUser, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return updatedUser, nil
}

func (s *UserService) generateJWT(user *models.User) (string, *models.User, error) {
	if user == nil {
		return "", nil, errors.New("user is nil")
	}
	tokenClaims := jwt.MapClaims{
		"user_id": user.ID.Hex(),
		"role":    user.Role,
		"exp":     time.Now().Add(time.Hour * time.Duration(s.config.TokenExpiry)).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaims)
	tokenString, err := token.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		log.Println("[ERROR] Failed to sign JWT:", err)
		return "", nil, err
	}

	return tokenString, user, nil
}
