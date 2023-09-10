package api

import (
	"crypto/sha256"
	"crypto/subtle"
	"net/http"
)

type BasicAuth struct {
	Username string
	Password string
}

func HTTPBasic(auth *BasicAuth, next http.HandlerFunc) http.HandlerFunc {
	if auth == nil {
		return next
	}

	expectedUsernameHash := sha256.Sum256([]byte(auth.Username))
	expectedPasswordHash := sha256.Sum256([]byte(auth.Password))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if ok {

			usernameHash := sha256.Sum256([]byte(username))
			passwordHash := sha256.Sum256([]byte(password))

			usernameMatch := (subtle.ConstantTimeCompare(usernameHash[:], expectedUsernameHash[:]) == 1)
			passwordMatch := (subtle.ConstantTimeCompare(passwordHash[:], expectedPasswordHash[:]) == 1)

			if usernameMatch && passwordMatch {
				next.ServeHTTP(w, r)
				return
			}
		}

		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}
