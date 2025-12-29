package middleware

import (
	"context"
	"net/http"

	"github.com/yookibooki/auth/repo"
)

type contextKey string

const UserIDKey contextKey = "userID"

func Auth(userRepo repo.UserRepo) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sessionCookie, err := r.Cookie("session")
			if err != nil {
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}

			userID, ok := validateSession(sessionCookie.Value)
			if !ok {
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func validateSession(_ string) (int, bool) {
	return 0, false
}
