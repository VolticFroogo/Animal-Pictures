package post

import (
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/VolticFroogo/Animal-Pictures/captcha"
	"github.com/VolticFroogo/Animal-Pictures/db"
	"github.com/VolticFroogo/Animal-Pictures/helpers"
	"github.com/VolticFroogo/Animal-Pictures/models"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
)

type voteRequest struct {
	Upvote             bool
	Captcha, CaptchaV2 string
}

type voteResponse struct {
	Score int
}

// Page is for the for posts.
func Page(w http.ResponseWriter, r *http.Request) {
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

	vars := mux.Vars(r)

	post, err := db.GetPost(vars["uuid"])
	if err != nil {
		helpers.ThrowErr(w, r, "Getting post from DB error", err)
		return
	}

	variables.Post = post

	var t *template.Template

	if post.Creation == 0 {
		t, err = template.ParseFiles("templates/post/not-found.html", "templates/nested.html") // Parse the HTML pages.
		if err != nil {
			helpers.ThrowErr(w, r, "Template parsing error", err)
			return
		}
	} else {
		t, err = template.ParseFiles("templates/post/page.html", "templates/nested.html") // Parse the HTML pages.
		if err != nil {
			helpers.ThrowErr(w, r, "Template parsing error", err)
			return
		}
	}

	if vote, ok := variables.Post.Votes[variables.Self.UUID]; ok {
		if vote {
			variables.Post.Vote = 1
		} else {
			variables.Post.Vote = 2
		}
	}

	err = t.Execute(w, variables)
	if err != nil {
		helpers.ThrowErr(w, r, "Template execution error", err)
	}
}

// Vote is for the for vote requests.
func Vote(w http.ResponseWriter, r *http.Request) {
	var data voteRequest                         // Create struct to store data.
	err := json.NewDecoder(r.Body).Decode(&data) // Decode response to struct.
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		helpers.ThrowErr(w, r, "JSON decoding error", err)
		return
	}

	// Secure our request with reCAPTCHA v2 and v3.
	if !captcha.V3(data.CaptchaV2, data.Captcha, r.Header.Get("CF-Connecting-IP"), "vote") {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)

	post, err := db.GetPost(vars["uuid"])
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		helpers.ThrowErr(w, r, "Getting post from DB error", err)
		return
	}

	if post.UserUUID == "" {
		// Post has been deleted.
		w.WriteHeader(http.StatusGone)
		return
	}

	score, err := db.SetVote(post, context.Get(r, "uuid").(string), data.Upvote)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		helpers.ThrowErr(w, r, "Setting vote error", err)
		return
	}

	helpers.JSONResponse(voteResponse{
		Score: score,
	}, w)
	return
}
