package models

import (
	"math"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

const (
	// AuthTokenValidTime is the lifetime of an auth token.
	AuthTokenValidTime = time.Hour * 2 // 2 hours.
	// RefreshTokenValidTime is the lifetime of a refresh token.
	RefreshTokenValidTime = time.Hour * 24 * 14 // 2 weeks.
	// EmailAntiSpamTime is the time we have to wait before sending another email.
	EmailAntiSpamTime = time.Hour * 24 // 1 day.
	// CautiousIPTime is how long an IP will stay on the cautious map.
	CautiousIPTime = time.Hour * 24 // 1 day.
	// CaptchaScore is the minimum score required to be accepted with a v3 reCAPTCHA.
	CaptchaScore = 0.5
	// HotPostsTickRate is how often the hot posts page should be updated.
	HotPostsTickRate = time.Minute // 1 minute.
	// PostsPerPage is how many posts there are on a page.
	PostsPerPage = 20
)

// Privileges
const (
	PrivUnverified = iota
	PrivUser
	PrivModerator
	PrivAdmin
)

// User is a user retrieved from a Database.
type User struct {
	Creation                                  int64
	Privilege                                 int `json:"-"`
	UUID, Username                            string
	Description, Fname, Lname, ImageExtension string `json:",omitempty"`
	Email, Password                           string `json:"-"`
}

// GetCreation is a template function used to return a human readable date from the creation unix timestamp.
func (user User) GetCreation() string {
	return time.Unix(user.Creation, 0).Format("Monday, 2 January 2006")
}

// HasProfilePicture returns if a user has a profile picture.
func (user User) HasProfilePicture() bool {
	return user.ImageExtension != ""
}

// ProfilePicture returns the URL of a user's profile picture.
func (user User) ProfilePicture() string {
	return "https://s3.eu-west-2.amazonaws.com/froogo-ap/user/" + user.UUID + user.ImageExtension
}

// TokenClaims are the claims in a token.
type TokenClaims struct {
	jwt.StandardClaims
	CSRF string `json:"csrf"`
}

// Post is the struct for posts.
type Post struct {
	UUID, UserUUID, Title, Description string
	Images                             []string
	Creation                           int64
	Votes                              map[string]bool `json:"-"`
	Upvotes, Downvotes                 int
	Rating                             float64
	Vote                               int `json:"-"`
}

// GetCreation is a template function used to return a human readable date from the creation unix timestamp.
func (post Post) GetCreation() string {
	return time.Unix(post.Creation, 0).Format("Monday, 2 January 2006")
}

// Score returns the overall score from votes of a post.
func (post Post) Score() int {
	return post.Upvotes - post.Downvotes
}

// GetRating gets the rating of a post.
func (post Post) GetRating() float64 {
	order := math.Log10(math.Max(math.Abs(float64(post.Score())), float64(1)))

	var sign float64
	if post.Score() > 0 {
		sign = 1
	} else if post.Score() < 0 {
		sign = -1
	}

	seconds := post.Creation - 1550144333
	return sign*order + float64(seconds/45000)
}

// TemplateVariables is the struct used when executing a template.
type TemplateVariables struct {
	CsrfSecret string
	Self, User User
	UnixTime   int64
	LoggedIn   bool
	Post       Post
	Posts      []Post
}

// AJAXData is the struct used with the AJAX middleware.
type AJAXData struct {
	CsrfSecret string
}

// JTI is the struct used for JTIs in the DB.
type JTI struct {
	ID            int
	Expiry        int64
	JTI, UserUUID string
}

// ResponseWithID is a simple struct for responding to an AJAX request.
type ResponseWithID struct {
	Success bool   `json:"success"`
	ID      string `json:"id"`
}

// ResponseWithIDInt is a simple struct for responding to an AJAX request.
type ResponseWithIDInt struct {
	Success bool `json:"success"`
	ID      int  `json:"id"`
}

// EmailTemplateVariables is the struct for template variables used when sending emails.
type EmailTemplateVariables struct {
	Code, Username string
}
