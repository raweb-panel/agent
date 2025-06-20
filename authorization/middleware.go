package authorization

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

var apiToken string

func InitAuthWithPath(projectPath string) {
	envPath := filepath.Join(projectPath, ".env")
	err := godotenv.Load(envPath)
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	apiToken = os.Getenv("APP_KEY")
	if apiToken == "" {
		log.Fatal("APP_KEY not set in .env")
	}
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token != "Bearer "+apiToken {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
