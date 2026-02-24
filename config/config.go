package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	MongoURI              string
	Database              string
	JWTSecret             string
	TokenExpiry           int
	GoogleClientID        string
	GoogleClientIDiOS     string
	GoogleClientIDAndroid string
	GoogleClientIDs       []string
	DefaultSearchesLeft   int
	MinAndroidVersionCode int
	MaxAndroidVersionCode int
	AndroidUpdateURL      string
	RequireAppHeadersAuth bool
}

func LoadConfig() *Config {
	// Load .env file automatically
	_ = godotenv.Load()

	uri := os.Getenv("MONGODB_URI")
	db := os.Getenv("MONGODB_DBNAME")
	google := os.Getenv("GOOGLE_CLIENT_ID")
	googleIOS := os.Getenv("GOOGLE_CLIENT_ID_IOS")
	googleAndroid := os.Getenv("GOOGLE_CLIENT_ID_ANDROID")
	googleAndroidDev := os.Getenv("GOOGLE_CLIENT_ID_ANDROID_DEV")
	googleAndroidProd := os.Getenv("GOOGLE_CLIENT_ID_ANDROID_PROD")
	googleClientIDsCSV := os.Getenv("GOOGLE_CLIENT_IDS")

	if uri == "" || db == "" {
		log.Fatal("Missing MONGODB_URI or MONGODB_DBNAME environment variables")
	}

	defaultSearches := 300
	if v := os.Getenv("DEFAULT_SEARCHES_LEFT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			defaultSearches = n
		}
	}

	minAndroidVersionCode := 0
	if v := os.Getenv("ANDROID_MIN_VERSION_CODE"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			minAndroidVersionCode = n
		}
	}

	maxAndroidVersionCode := 0
	if v := os.Getenv("ANDROID_MAX_VERSION_CODE"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			maxAndroidVersionCode = n
		}
	}

	googleClientIDs := buildGoogleClientIDList(
		google,
		googleIOS,
		googleAndroid,
		googleAndroidDev,
		googleAndroidProd,
		googleClientIDsCSV,
	)
	if len(googleClientIDs) == 0 {
		log.Fatal("Missing Google OAuth client IDs. Set at least one of GOOGLE_CLIENT_ID / GOOGLE_CLIENT_ID_ANDROID / GOOGLE_CLIENT_IDS")
	}

	return &Config{
		MongoURI:              uri,
		Database:              db,
		JWTSecret:             os.Getenv("JWT_SECRET"),
		TokenExpiry:           72,
		GoogleClientID:        google,
		GoogleClientIDiOS:     googleIOS,
		GoogleClientIDAndroid: googleAndroid,
		GoogleClientIDs:       googleClientIDs,
		DefaultSearchesLeft:   defaultSearches,
		MinAndroidVersionCode: minAndroidVersionCode,
		MaxAndroidVersionCode: maxAndroidVersionCode,
		AndroidUpdateURL:      strings.TrimSpace(os.Getenv("ANDROID_UPDATE_URL")),
		RequireAppHeadersAuth: isTruthy(os.Getenv("REQUIRE_APP_HEADERS_FOR_AUTH")),
	}
}

func buildGoogleClientIDList(values ...string) []string {
	seen := make(map[string]struct{})
	var ids []string

	for _, value := range values {
		for _, raw := range strings.Split(value, ",") {
			id := strings.TrimSpace(raw)
			if id == "" {
				continue
			}
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			ids = append(ids, id)
		}
	}

	return ids
}

func isTruthy(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}
