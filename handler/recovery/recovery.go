package recovery

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/VolticFroogo/Animal-Pictures/captcha"
	"github.com/VolticFroogo/Animal-Pictures/db"
	"github.com/VolticFroogo/Animal-Pictures/email"
	"github.com/VolticFroogo/Animal-Pictures/helpers"
	"github.com/VolticFroogo/Animal-Pictures/models"
)

// Response codes.
const (
	Success = iota
	InvalidEmail
	Recaptcha
	Internal
	SendingEmail
	InvalidCode
)

type message struct {
	Code, Password, Email, Captcha, CaptchaV2 string
}

type response struct {
	Code int
}

// Begin is the function called after an AJAX request is sent from the forgot password page.
func Begin(w http.ResponseWriter, r *http.Request) {
	var data message                             // Create struct to store data.
	err := json.NewDecoder(r.Body).Decode(&data) // Decode response to struct.
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		helpers.ThrowErr(w, r, "JSON decoding error", err)
		return
	}

	// Secure our request with reCAPTCHA v2 and v3.
	if !captcha.V3(data.CaptchaV2, data.Captcha, r.Header.Get("CF-Connecting-IP"), "forgot_password") {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := db.GetUserFromEmail(data.Email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		helpers.ThrowErr(w, r, "Getting user error", err)
		return
	}
	if user.UUID == "" {
		// Even though we aren't sending an email we can't reveal if a user exists so we say that we MAY have sent an email.
		w.WriteHeader(http.StatusOK)
		return
	}

	// Check if we have sent a recovery email within the last X amount of time.
	// If we have we won't send them an email to prevent spam.
	_, _, creation, err := db.GetRecoveryFromUser(user.UUID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		helpers.ThrowErr(w, r, "Getting previous recovery error", err)
		return
	} else if creation != 0 && time.Unix(creation, 0).Add(models.EmailAntiSpamTime).After(time.Now()) {
		// Even though we aren't sending an email we can't reveal if a user exists so we say that we MAY have sent an email.
		w.WriteHeader(http.StatusOK)
		return
	}

	code, err := db.AddRecovery(user.UUID, data.Email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		helpers.ThrowErr(w, r, "Adding recovery error", err)
		return
	}

	err = email.Recovery(code, user.Username, data.Email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		helpers.ThrowErr(w, r, "Send email error", err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// End is the final function which is called when a user submits their new password.
func End(w http.ResponseWriter, r *http.Request) {
	var data message                             // Create struct to store data.
	err := json.NewDecoder(r.Body).Decode(&data) // Decode response to struct.
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		helpers.ThrowErr(w, r, "JSON decoding error", err)
		return
	}

	// Secure our request with reCAPTCHA v2 and v3.
	if !captcha.V3(data.CaptchaV2, data.Captcha, r.Header.Get("CF-Connecting-IP"), "reset_password") {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userUUID, email, err := db.GetRecovery(data.Code)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		helpers.ThrowErr(w, r, "Getting recovery error", err)
		return
	}

	if userUUID == "" || email == "" {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	hash, err := helpers.HashPassword(data.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		helpers.ThrowErr(w, r, "Hashing password error", err)
		return
	}

	err = db.EditPassword(userUUID, hash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		helpers.ThrowErr(w, r, "Editing password error", err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
