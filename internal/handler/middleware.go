package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/GuilhAndrad/point-api/internal/model"
	"github.com/go-chi/jwtauth"
)

type contextKey string
const userIDKey contextKey = "userID"
const userRoleKey contextKey = "userRole"

// AuthMiddleware valida o JWT e injeta userID e role no contexto
func AuthMiddleware(secret string) func(http.Handler) http.Handler {
    tokenAuth := jwtauth.New("HS256", []byte(secret), nil)

    return func(next http.Handler) http.Handler {
        return jwtauth.Verifier(tokenAuth)(
            http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

                token, claims, err := jwtauth.FromContext(r.Context())
                _ = token

                if err != nil || claims == nil {
                    respondError(w, http.StatusUnauthorized, "token inválido")
                    return
                }

                ctx := context.WithValue(r.Context(), userIDKey, int64(claims["user_id"].(float64)))
                ctx = context.WithValue(ctx, userRoleKey, model.Role(claims["role"].(string)))

                next.ServeHTTP(w, r.WithContext(ctx))
            }),
        )
    }
}

// AdminOnly bloqueia acesso de não-administradores
func AdminOnly(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        role, _ := r.Context().Value(userRoleKey).(model.Role)
        if role != model.RoleAdmin {
            respondError(w, http.StatusForbidden, "acesso restrito a administradores")
            return
        }
        next.ServeHTTP(w, r)
    })
}

func getUserID(ctx context.Context) int64 {
    id, _ := ctx.Value(userIDKey).(int64)
    return id
}

// respondJSON e respondError: helpers usados em todos os handlers
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, msg string) {
    respondJSON(w, status, map[string]string{"error": msg})
}