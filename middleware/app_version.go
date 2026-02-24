package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"USDT_BackEnd/config"
)

const (
	appPlatformHeader    = "X-App-Platform"
	appVersionHeader     = "X-App-Version"
	appVersionCodeHeader = "X-App-Version-Code"
	apiPrefix            = "/api/"
	authPrefix           = "/api/auth/"
)

type updateRequiredResponse struct {
	Code           string `json:"code"`
	Message        string `json:"message"`
	MinVersionCode int    `json:"minVersionCode"`
	StoreURL       string `json:"storeUrl,omitempty"`
}

// AppVersionMiddleware blocks old Android app versions when ANDROID_MIN_VERSION_CODE is configured.
// It only applies to API routes under /api/.
func AppVersionMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.MinAndroidVersionCode <= 0 || !strings.HasPrefix(r.URL.Path, apiPrefix) {
				next.ServeHTTP(w, r)
				return
			}

			headerPlatform := strings.ToLower(strings.TrimSpace(r.Header.Get(appPlatformHeader)))
			platform := headerPlatform
			if platform == "" {
				platform = inferPlatformFromUserAgent(r.UserAgent())
			}

			// Require explicit app headers only for Android auth calls.
			// This blocks old Android clients that don't send version headers,
			// while letting legacy iOS requests continue.
			if cfg.RequireAppHeadersAuth && strings.HasPrefix(r.URL.Path, authPrefix) &&
				headerPlatform == "" && platform == "android" {
				log.Printf("[AppVersion] blocked android auth without platform header: path=%s ua=%q", r.URL.Path, r.UserAgent())
				writeUpdateRequired(w, cfg)
				return
			}

			if platform != "android" {
				next.ServeHTTP(w, r)
				return
			}

			versionCodeStr := strings.TrimSpace(r.Header.Get(appVersionCodeHeader))
			versionCode, err := strconv.Atoi(versionCodeStr)
			if err == nil && cfg.MaxAndroidVersionCode > 0 && versionCode > cfg.MaxAndroidVersionCode {
				log.Printf(
					"[AppVersion] blocked android request by max version: path=%s x-app-version=%q x-app-version-code=%d max=%d ua=%q",
					r.URL.Path,
					strings.TrimSpace(r.Header.Get(appVersionHeader)),
					versionCode,
					cfg.MaxAndroidVersionCode,
					r.UserAgent(),
				)
				writeUpdateRequired(w, cfg)
				return
			}
			if err == nil && versionCode >= cfg.MinAndroidVersionCode {
				next.ServeHTTP(w, r)
				return
			}

			log.Printf(
				"[AppVersion] blocked android request by min version: path=%s x-app-version=%q x-app-version-code=%q min=%d ua=%q",
				r.URL.Path,
				strings.TrimSpace(r.Header.Get(appVersionHeader)),
				versionCodeStr,
				cfg.MinAndroidVersionCode,
				r.UserAgent(),
			)
			writeUpdateRequired(w, cfg)
		})
	}
}

func inferPlatformFromUserAgent(userAgent string) string {
	ua := strings.ToLower(strings.TrimSpace(userAgent))
	if ua == "" {
		return ""
	}

	if strings.Contains(ua, "android") || strings.Contains(ua, "okhttp") {
		return "android"
	}

	if strings.Contains(ua, "iphone") ||
		strings.Contains(ua, "ipad") ||
		strings.Contains(ua, "ios") ||
		strings.Contains(ua, "cfnetwork") ||
		strings.Contains(ua, "darwin") {
		return "ios"
	}

	return ""
}

func writeUpdateRequired(w http.ResponseWriter, cfg *config.Config) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUpgradeRequired)
	_ = json.NewEncoder(w).Encode(updateRequiredResponse{
		Code:           "APP_UPDATE_REQUIRED",
		Message:        "This app version is no longer supported. Please update to continue.",
		MinVersionCode: cfg.MinAndroidVersionCode,
		StoreURL:       cfg.AndroidUpdateURL,
	})
}
