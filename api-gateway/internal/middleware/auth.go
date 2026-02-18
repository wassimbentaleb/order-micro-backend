package middleware

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/redis/go-redis/v9"
)

func AuthMiddleware(rdb *redis.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")
			if len(token) > 7 && token[:7] == "Bearer " {
				token = token[7:]
			}

			if token == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"error": "missing token"})
				return
			}

			val, err := rdb.Get(context.Background(), "session:"+token).Result()
			if err == redis.Nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"error": "invalid or expired session"})
				return
			}
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": "session validation failed"})
				return
			}

			// Parse user from session and add user_id to header for downstream services
			var user map[string]interface{}
			if err := json.Unmarshal([]byte(val), &user); err == nil {
				if id, ok := user["id"].(string); ok {
					r.Header.Set("X-User-ID", id)
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
