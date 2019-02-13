package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/VolticFroogo/Animal-Pictures/helpers"
	"github.com/VolticFroogo/Animal-Pictures/middleware/myJWT"
	"github.com/VolticFroogo/Animal-Pictures/models"
	"github.com/gorilla/context"
)

type apiAuthResponse struct {
	AuthToken, RefreshToken string
}

// View handles authentication for requests which have features accessible by users but being logged in isn't necessary.
func View(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	authTokenString, err := r.Cookie("authToken")
	if err != nil {
		if err == http.ErrNoCookie {
			next(w, r)
			return
		}

		helpers.ThrowErr(w, r, "Reading cookie error", err)
		return
	}

	refreshTokenString, err := r.Cookie("refreshToken")
	if err != nil {
		if err == http.ErrNoCookie {
			next(w, r)
			return
		}

		helpers.ThrowErr(w, r, "Reading cookie error", err)
		return
	}

	if authTokenString.Value != "" {
		authTokenValid, uuid, err := myJWT.CheckToken(authTokenString.Value, "", false, false, true)
		if err != nil {
			helpers.ThrowErr(w, r, "Checking token error", err)
			return
		}

		if authTokenValid {
			context.Set(r, "uuid", uuid)
			next(w, r)
			return
		}
	}

	if refreshTokenString.Value != "" {
		refreshTokenValid, uuid, err := myJWT.CheckToken(refreshTokenString.Value, "", true, false, true)
		if err != nil {
			helpers.ThrowErr(w, r, "Checking token error", err)
			return
		}

		if refreshTokenValid {
			newAuthTokenString, newRefreshTokenString, newCsrfSecret, err := myJWT.RefreshTokens(refreshTokenString.Value)
			if err != nil {
				helpers.ThrowErr(w, r, "Creating new tokens error", err)
				return
			}

			WriteNewAuth(w, r, newAuthTokenString, newRefreshTokenString, newCsrfSecret)

			context.Set(r, "uuid", uuid)
			next(w, r)
			return
		}
	}
}

// User handles authentication for requests that should only be accessible by a user that is logged in, such as creating a new post.
func User(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	authTokenString, err := r.Cookie("authToken")
	if err != nil {
		if err == http.ErrNoCookie {
			return
		}

		helpers.ThrowErr(w, r, "Reading cookie error", err)
		return
	}

	refreshTokenString, err := r.Cookie("refreshToken")
	if err != nil {
		if err == http.ErrNoCookie {
			return
		}

		helpers.ThrowErr(w, r, "Reading cookie error", err)
		return
	}

	if authTokenString.Value != "" {
		authTokenValid, uuid, err := myJWT.CheckToken(authTokenString.Value, "", false, false, true)
		if err != nil {
			helpers.ThrowErr(w, r, "Checking token error", err)
			return
		}

		if authTokenValid {
			context.Set(r, "uuid", uuid)
			next(w, r)
			return
		}
	}

	if refreshTokenString.Value != "" {
		refreshTokenValid, uuid, err := myJWT.CheckToken(refreshTokenString.Value, "", true, false, true)
		if err != nil {
			helpers.ThrowErr(w, r, "Checking token error", err)
			return
		}

		if refreshTokenValid {
			newAuthTokenString, newRefreshTokenString, newCsrfSecret, err := myJWT.RefreshTokens(refreshTokenString.Value)
			if err != nil {
				helpers.ThrowErr(w, r, "Creating new tokens error", err)
				return
			}

			WriteNewAuth(w, r, newAuthTokenString, newRefreshTokenString, newCsrfSecret)

			context.Set(r, "uuid", uuid)
			next(w, r)
			return
		}
	}
}

// API handles authentication for API requests.
// NOTE: This feature is not yet used and may not be used in production build.
func API(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	token := r.Header.Get("token")

	if token != "" {
		tokenValid, uuid, err := myJWT.CheckToken(token, "", true, false, false)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("Checking token error: %v", err)
			return
		}

		if tokenValid {
			context.Set(r, "uuid", uuid)
			next(w, r)
			return
		}
	}

	w.WriteHeader(http.StatusUnauthorized)
}

// WriteNewAuth writes authentication to a user's browser.
func WriteNewAuth(w http.ResponseWriter, r *http.Request, authTokenString, refreshTokenString, csrfSecret string) {
	expiration := time.Now().Add(models.RefreshTokenValidTime)

	cookie := http.Cookie{Name: "authToken", Value: authTokenString, Expires: expiration, Path: "/", HttpOnly: true, Secure: true}
	http.SetCookie(w, &cookie)

	cookie = http.Cookie{Name: "refreshToken", Value: refreshTokenString, Expires: expiration, Path: "/", HttpOnly: true, Secure: true}
	http.SetCookie(w, &cookie)

	cookie = http.Cookie{Name: "csrfSecret", Value: csrfSecret, Expires: expiration, Path: "/", HttpOnly: true, Secure: true}
	http.SetCookie(w, &cookie)

	return
}

// RedirectToLogin redirects the client to the login.
func RedirectToLogin(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
}
