package routes

import (
	"errors"
	"net/http"

	"USDT_BackEnd/config"
	"USDT_BackEnd/handlers"
	"USDT_BackEnd/middleware"
	"USDT_BackEnd/services"

	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func RegisterRoutes(mux *http.ServeMux, cfg *config.Config) {
	// ====== Services ======
	userService := services.NewUserService(cfg)
	subService := services.NewSubscriptionService() // ✅ no args

	// ====== Handlers ======
	wordHandler := handlers.NewWordHandler() // internally creates its own service
	userHandler := handlers.NewUserHandler(userService)
	subHandler := handlers.NewSubscriptionHandler(subService)

	// ====== Middlewares ======
	auth := middleware.AuthMiddleware(cfg)

	// ========== PUBLIC ROUTES ==========
	mux.HandleFunc("GET /api/words/search", wordHandler.SearchWords)
	mux.HandleFunc("GET /api/words/{id}", wordHandler.GetWordByID)

	// User authentication
	mux.HandleFunc("POST /api/auth/register", userHandler.Register)
	mux.HandleFunc("POST /api/auth/login", userHandler.Login)

	// Subscription plans (publicly visible)
	mux.HandleFunc("GET /api/subscriptions/prices", subHandler.GetPrices)

	// ========== AUTHENTICATED ROUTES ==========
	mux.Handle("PUT /api/users/password", auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := r.Context().Value(middleware.UserKey)
		if claims == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		userID, _ := extractUserIDFromClaims(claims)
		userHandler.ChangePassword(w, r, userID)
	})))

	mux.Handle("GET /api/users/favorites", auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := r.Context().Value(middleware.UserKey)
		if claims == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		userID, _ := extractUserIDFromClaims(claims)
		userHandler.GetFavorites(w, r, userID)
	})))

	// Delete own account
mux.Handle("DELETE /api/users/me", auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(middleware.UserKey)
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := extractUserIDFromClaims(claims)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	userHandler.DeleteMe(w, r, userID)
})))

	// Favorite management
	mux.Handle("POST /api/users/favorites/add", auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := r.Context().Value(middleware.UserKey)
		if claims == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		userID, _ := extractUserIDFromClaims(claims)
		userHandler.SaveFavorite(w, r, userID)
	})))
	mux.Handle("GET /api/users/profile", auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := r.Context().Value(middleware.UserKey)
		if claims == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		userID, _ := extractUserIDFromClaims(claims)
		userHandler.GetProfile(w, r, userID)
	})))
	mux.Handle("DELETE /api/users/favorites/remove", auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := r.Context().Value(middleware.UserKey)
		if claims == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		userID, _ := extractUserIDFromClaims(claims)
		userHandler.RemoveFavorite(w, r, userID)
	})))

	mux.Handle("GET /api/users/favorites/paginated", auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := r.Context().Value(middleware.UserKey)
		if claims == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		userID, _ := extractUserIDFromClaims(claims)
		userHandler.GetFavoritesPaginated(w, r, userID)
	})))

	// ========== ADMIN ROUTES ==========
	mux.Handle("GET /api/words", auth(http.HandlerFunc(wordHandler.GetAllWords)))
	mux.Handle("POST /api/words", auth(http.HandlerFunc(wordHandler.CreateWord)))
	mux.Handle("PUT /api/words/{id}", auth(http.HandlerFunc(wordHandler.UpdateWord)))
	mux.Handle("DELETE /api/words/{id}", auth(http.HandlerFunc(wordHandler.DeleteWord)))
	// Select single word (shared for user/admin)
	mux.HandleFunc("GET /api/words/selectone/{id}", wordHandler.SelectOneWord)

	// Excel upload (authenticated route)
	mux.Handle("POST /api/words/excel-upload", auth(http.HandlerFunc(wordHandler.ExcelCreateWords)))

	mux.Handle("GET /api/users/subscribed", auth(http.HandlerFunc(userHandler.GetSubscribedUsers)))
	mux.Handle("PUT /api/subscriptions/prices", auth(http.HandlerFunc(subHandler.UpdatePrices)))

	// ===== Optional: Health Check =====
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status": "ok"}`))
	})
}

// ========== Helpers ==========

func extractUserIDFromClaims(claims interface{}) (primitive.ObjectID, error) {
	if m, ok := claims.(jwt.MapClaims); ok {
		if idVal, exists := m["user_id"]; exists {
			if idStr, ok := idVal.(string); ok && idStr != "" {
				return primitive.ObjectIDFromHex(idStr)
			}
		}
	}
	return primitive.NilObjectID, errors.New("user_id not found in claims")
}
