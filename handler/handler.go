package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"text/template"

	"github.com/VolticFroogo/Animal-Pictures/captcha"
	"github.com/VolticFroogo/Animal-Pictures/db"
	"github.com/VolticFroogo/Animal-Pictures/email"
	"github.com/VolticFroogo/Animal-Pictures/handler/post"
	"github.com/VolticFroogo/Animal-Pictures/handler/recovery"
	"github.com/VolticFroogo/Animal-Pictures/handler/user"
	"github.com/VolticFroogo/Animal-Pictures/helpers"
	"github.com/VolticFroogo/Animal-Pictures/middleware"
	"github.com/VolticFroogo/Animal-Pictures/middleware/myJWT"
	"github.com/VolticFroogo/Animal-Pictures/models"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
)

type formData struct {
	Email, Username, Password, Captcha, CaptchaV2 string
}

// Start the server by handling the web server.
func Start() {
	r := mux.NewRouter()
	r.StrictSlash(true)
	r.NotFoundHandler = http.HandlerFunc(notFound)

	r.Handle("/", negroni.New(
		negroni.HandlerFunc(middleware.View),
		negroni.Wrap(http.HandlerFunc(index)),
	)).Methods(http.MethodGet)

	r.Handle("/login", http.HandlerFunc(login)).Methods(http.MethodPost)
	r.Handle("/register", http.HandlerFunc(register)).Methods(http.MethodPost)

	r.Handle("/logout", negroni.New(
		negroni.HandlerFunc(middleware.User),
		negroni.Wrap(http.HandlerFunc(logout)),
	))

	r.Handle("/forgot-password", http.HandlerFunc(recovery.Begin)).Methods(http.MethodPost)
	r.Handle("/password-recovery", http.HandlerFunc(recovery.End)).Methods(http.MethodPost)

	r.Handle("/verify/{code}", http.HandlerFunc(user.Verify)).Methods(http.MethodGet)

	r.Handle("/user/{uuid}", negroni.New(
		negroni.HandlerFunc(middleware.View),
		negroni.Wrap(http.HandlerFunc(user.Page)),
	)).Methods(http.MethodGet)

	r.Handle("/post/new", negroni.New(
		negroni.HandlerFunc(middleware.View),
		negroni.Wrap(http.HandlerFunc(post.PageNew)),
	)).Methods(http.MethodGet)

	r.Handle("/post/new", negroni.New(
		negroni.HandlerFunc(middleware.User),
		negroni.Wrap(http.HandlerFunc(post.New)),
	)).Methods(http.MethodPost)

	r.Handle("/post/{uuid}", negroni.New(
		negroni.HandlerFunc(middleware.View),
		negroni.Wrap(http.HandlerFunc(post.Page)),
	)).Methods(http.MethodGet)

	r.Handle("/post/{uuid}/vote", negroni.New(
		negroni.HandlerFunc(middleware.User),
		negroni.Wrap(http.HandlerFunc(post.Vote)),
	)).Methods(http.MethodPost)

	r.PathPrefix("/login").Handler(http.FileServer(http.Dir("./static/")))
	r.PathPrefix("/register").Handler(http.FileServer(http.Dir("./static/")))
	r.PathPrefix("/forgot-password").Handler(http.FileServer(http.Dir("./static/")))
	r.PathPrefix("/password-recovery").Handler(http.FileServer(http.Dir("./static/")))
	r.PathPrefix("/robots.txt").Handler(http.FileServer(http.Dir("./static/")))
	r.PathPrefix("/css/").Handler(http.FileServer(http.Dir("./static/")))
	r.PathPrefix("/js/").Handler(http.FileServer(http.Dir("./static/")))

	log.Printf("Server started...")
	http.ListenAndServe(":87", r)
}

func notFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)

	t, err := template.ParseFiles("templates/not-found.html", "templates/nested.html") // Parse the HTML pages.
	if err != nil {
		helpers.ThrowErr(w, r, "Template parsing error", err)
		return
	}

	err = t.Execute(w, models.TemplateVariables{})
	if err != nil {
		helpers.ThrowErr(w, r, "Template execution error", err)
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	uuid, loggedIn := context.GetOk(r, "uuid")

	variables := models.TemplateVariables{
		LoggedIn: loggedIn,
	}

	if loggedIn {
		user, err := db.GetUserFromUUID(uuid.(string))
		if err != nil {
			helpers.ThrowErr(w, r, "Getting user from DB error", err)
			return
		}

		csrfSecret, err := r.Cookie("csrfSecret")
		if err != nil {
			helpers.ThrowErr(w, r, "Getting CSRF Secret cookie error", err)
			return
		}

		variables.Self = user
		variables.CsrfSecret = csrfSecret.Value
	}

	var err error
	variables.Posts, err = db.GetHotPosts(0)
	if err != nil {
		helpers.ThrowErr(w, r, "Getting hot posts error", err)
		return
	}

	t, err := template.ParseFiles("templates/index.html", "templates/nested.html") // Parse the HTML pages.
	if err != nil {
		helpers.ThrowErr(w, r, "Template parsing error", err)
		return
	}

	err = t.Execute(w, variables)
	if err != nil {
		helpers.ThrowErr(w, r, "Template execution error", err)
	}
}

func login(w http.ResponseWriter, r *http.Request) {
	var credentials formData                            // Create struct to store data.
	err := json.NewDecoder(r.Body).Decode(&credentials) // Decode response to struct.
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		helpers.ThrowErr(w, r, "JSON decoding error", err)
		return
	}

	// Secure our request with reCAPTCHA v2 and v3.
	if !captcha.V3(credentials.CaptchaV2, credentials.Captcha, r.Header.Get("CF-Connecting-IP"), "login") {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := db.GetUserFromEmail(credentials.Email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		helpers.ThrowErr(w, r, "Getting user from DB error", err)
		return
	}

	if !helpers.CheckPassword(credentials.Password, user.Password) {
		// If the user has got the password wrong send a status unauthorized header.
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if user.Privilege == models.PrivUnverified {
		// User has not yet verified their email; send them a forbidden header.
		w.WriteHeader(http.StatusForbidden)
		return
	}

	authTokenString, refreshTokenString, csrfSecret, err := myJWT.CreateNewTokens(user.UUID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		helpers.ThrowErr(w, r, "Creating tokens error", err)
		return
	}

	middleware.WriteNewAuth(w, r, authTokenString, refreshTokenString, csrfSecret)

	w.WriteHeader(http.StatusOK)
}

func register(w http.ResponseWriter, r *http.Request) {
	var data formData                            // Create struct to store data.
	err := json.NewDecoder(r.Body).Decode(&data) // Decode response to struct.
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		helpers.ThrowErr(w, r, "JSON decoding error", err)
		return
	}

	// Secure our request with reCAPTCHA v2 and v3.
	if !captcha.V3(data.CaptchaV2, data.Captcha, r.Header.Get("CF-Connecting-IP"), "register") {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check if the email is valid.
	if !helpers.CheckEmail(data.Email) {
		// Email is invalid, return not acceptable.
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	// Check if a user exists with the requested email.
	exists, err := db.UserExistsFromEmail(data.Email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		helpers.ThrowErr(w, r, "Checking if user exists error", err)
		return
	} else if exists {
		w.WriteHeader(http.StatusConflict)
		return // A user already exists with email.
	}

	// Hash the password.
	hash, err := helpers.HashPassword(data.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		helpers.ThrowErr(w, r, "Hashing password error", err)
		return
	}

	// Create the new user.
	uuid, err := db.NewUser(data.Email, hash, data.Username, models.PrivUnverified)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		helpers.ThrowErr(w, r, "Creating user error", err)
		return
	}

	// Add the email verification to the database.
	code, err := db.AddEmailVerification(uuid, data.Email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		helpers.ThrowErr(w, r, "Creating email verification error", err)
		return
	}

	// Send the registration email.
	err = email.Register(code, data.Username, data.Email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		helpers.ThrowErr(w, r, "Sending registration email error", err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func logout(w http.ResponseWriter, r *http.Request) {
	refreshTokenString, err := r.Cookie("refreshToken")
	if err != nil {
		helpers.ThrowErr(w, r, "Reading cookie error", err)
		return
	}

	myJWT.DeleteJTI(refreshTokenString.Value) // Remove their old Refresh Token.

	middleware.WriteNewAuth(w, r, "", "", "")

	middleware.RedirectToLogin(w, r)
}
