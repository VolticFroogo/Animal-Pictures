package post

import (
	"html/template"
	"net/http"

	"github.com/VolticFroogo/Animal-Pictures/captcha"
	"github.com/VolticFroogo/Animal-Pictures/db"
	"github.com/VolticFroogo/Animal-Pictures/helpers"
	"github.com/VolticFroogo/Animal-Pictures/models"
	"github.com/VolticFroogo/Animal-Pictures/upload"
	"github.com/gorilla/context"
)

type response struct {
	UUID string
}

// New is the handler for the new post request.
func New(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 5*1024*1024) // 50MB max request size otherwise decline.
	err := r.ParseMultipartForm(5 * 1024 * 1024)         // Parse multipart form, use total 5MB of RAM.
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		helpers.ThrowErr(w, r, "Parsing multipart form error", err)
		return
	}

	form := r.MultipartForm // Define the form from the request's multipart form.

	var v2, v3 string

	if val, ok := form.Value["captchaV2"]; ok {
		v2 = val[0]
	}

	if val, ok := form.Value["captcha"]; ok {
		v3 = val[0]
	}

	// Secure our request with reCAPTCHA v2 and v3.
	if !captcha.V3(v2, v3, r.Header.Get("CF-Connecting-IP"), "post_new") {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Upload the image.
	location, err := upload.Image(form.File["image"][0])
	if err != nil {
		if err == upload.ErrNotImage {
			// They are trying to upload a file that we think isn't an image.
			w.WriteHeader(http.StatusUnsupportedMediaType)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		helpers.ThrowErr(w, r, "Uploading image error", err)
		return
	}

	images := []string{location}
	post, err := db.NewPost(form.Value["title"][0], form.Value["description"][0], context.Get(r, "uuid").(string), images)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		helpers.ThrowErr(w, r, "Adding Post to DB error", err)
		return
	}

	helpers.JSONResponse(response{
		UUID: post.UUID,
	}, w)
}

// PageNew is the handler for the new post page.
func PageNew(w http.ResponseWriter, r *http.Request) {
	uuid, loggedIn := context.GetOk(r, "uuid")

	variables := models.TemplateVariables{
		LoggedIn: loggedIn,
	}

	if loggedIn {
		self, err := db.GetUserFromUUID(uuid.(string))
		if err != nil {
			helpers.ThrowErr(w, r, "Getting user from DB error", err)
			return
		}

		csrfSecret, err := r.Cookie("csrfSecret")
		if err != nil {
			helpers.ThrowErr(w, r, "Getting CSRF Secret cookie error", err)
			return
		}

		variables.Self = self
		variables.CsrfSecret = csrfSecret.Value
	}

	t, err := template.ParseFiles("templates/post/new.html", "templates/nested.html") // Parse the HTML pages.
	if err != nil {
		helpers.ThrowErr(w, r, "Template parsing error", err)
		return
	}

	err = t.Execute(w, variables)
	if err != nil {
		helpers.ThrowErr(w, r, "Template execution error", err)
	}
}
