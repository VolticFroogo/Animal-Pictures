package user

import (
	"html/template"
	"net/http"

	"github.com/VolticFroogo/Animal-Pictures/db"
	"github.com/VolticFroogo/Animal-Pictures/helpers"
	"github.com/VolticFroogo/Animal-Pictures/models"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
)

// Page is the response for a GET request to a user's page.
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

	user, err := db.GetUserFromUUID(vars["uuid"])
	if err != nil {
		helpers.ThrowErr(w, r, "Getting user from DB error", err)
		return
	}

	variables.User = user

	var t *template.Template

	if user.Creation == 0 {
		t, err = template.ParseFiles("templates/user/not-found.html", "templates/nested.html") // Parse the HTML pages.
		if err != nil {
			helpers.ThrowErr(w, r, "Template parsing error", err)
			return
		}
	} else {
		t, err = template.ParseFiles("templates/user/page.html", "templates/nested.html") // Parse the HTML pages.
		if err != nil {
			helpers.ThrowErr(w, r, "Template parsing error", err)
			return
		}
	}

	err = t.Execute(w, variables)
	if err != nil {
		helpers.ThrowErr(w, r, "Template execution error", err)
	}
}

// Verify is the response for when a user clicks the verify button after registering their account.
func Verify(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	uuid, _, err := db.GetEmailVerification(vars["code"])
	if err != nil {
		helpers.ThrowErr(w, r, "Getting email verification error", err)
		return
	}

	if uuid == "" {
		http.Redirect(w, r, "https://ap.froogo.co.uk/login/?code=2", http.StatusTemporaryRedirect)
		return
	}

	err = db.EditPrivilege(uuid, models.PrivUser)
	if err != nil {
		helpers.ThrowErr(w, r, "Editing privilege error", err)
		return
	}

	http.Redirect(w, r, "https://ap.froogo.co.uk/login/?code=1", http.StatusTemporaryRedirect)
}
