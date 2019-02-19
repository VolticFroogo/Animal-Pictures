package db

import (
	"encoding/json"
	"time"

	"github.com/VolticFroogo/Animal-Pictures/models"
	"github.com/zemirco/uid"
)

// GetHotPosts will get the respective hot posts for a given page.
func GetHotPosts(page int) (posts []models.Post, err error) {
	rows, err := db.Query("SELECT P.uuid, P.title, P.description, P.images, P.votes, P.rating, P.creation, U.uuid, U.email, U.password, U.username, U.privilege, U.creation, U.fname, U.lname, U.description, U.imageExtension FROM posts AS P INNER JOIN users AS U ON P.useruuid = U.uuid ORDER BY P.rating DESC, P.creation DESC LIMIT ?, ?", page*models.PostsPerPage, models.PostsPerPage+page*models.PostsPerPage)
	if err != nil {
		return
	}

	defer rows.Close()

	for rows.Next() {
		var post models.Post
		var imagesJSON, votesJSON string

		err = rows.Scan(&post.UUID, &post.Title, &post.Description, &imagesJSON, &votesJSON, &post.Rating, &post.Creation, &post.Owner.UUID, &post.Owner.Email, &post.Owner.Password, &post.Owner.Username, &post.Owner.Privilege, &post.Owner.Creation, &post.Owner.Fname, &post.Owner.Lname, &post.Owner.Description, &post.Owner.ImageExtension) // Scan data from query.
		if err != nil {
			return
		}

		err = json.Unmarshal([]byte(imagesJSON), &post.Images)
		if err != nil {
			return
		}

		err = json.Unmarshal([]byte(votesJSON), &post.Votes)
		if err != nil {
			return
		}

		for _, upvote := range post.Votes {
			if upvote {
				post.Upvotes++
			} else {
				post.Downvotes++
			}
		}

		posts = append(posts, post)
	}

	return
}

// GetPost returns a post given a UUID.
func GetPost(uuid string) (post models.Post, err error) {
	rows, err := db.Query("SELECT P.title, P.description, P.images, P.votes, P.rating, P.creation, U.uuid, U.email, U.password, U.username, U.privilege, U.creation, U.fname, U.lname, U.description, U.imageExtension FROM posts AS P INNER JOIN users AS U ON P.useruuid = U.uuid WHERE P.uuid=?", uuid)
	if err != nil {
		return
	}

	defer rows.Close()

	post.UUID = uuid
	if rows.Next() {
		var imagesJSON, votesJSON string

		err = rows.Scan(&post.Title, &post.Description, &imagesJSON, &votesJSON, &post.Rating, &post.Creation, &post.Owner.UUID, &post.Owner.Email, &post.Owner.Password, &post.Owner.Username, &post.Owner.Privilege, &post.Owner.Creation, &post.Owner.Fname, &post.Owner.Lname, &post.Owner.Description, &post.Owner.ImageExtension) // Scan data from query.
		if err != nil {
			return
		}

		err = json.Unmarshal([]byte(imagesJSON), &post.Images)
		if err != nil {
			return
		}

		err = json.Unmarshal([]byte(votesJSON), &post.Votes)
		if err != nil {
			return
		}

		for _, upvote := range post.Votes {
			if upvote {
				post.Upvotes++
			} else {
				post.Downvotes++
			}
		}
	}

	return
}

// NewPost creates a new post.
func NewPost(title, description, userUUID string, images []string) (post models.Post, err error) {
	imagesJSON, err := json.Marshal(images)
	if err != nil {
		return
	}

	var exists bool
	for {
		post.UUID = uid.New(8)
		exists, err = rowExists("SELECT useruuid FROM posts WHERE uuid=?", post.UUID)
		if err != nil {
			return post, err
		}

		if !exists {
			break
		}
	}

	post = models.Post{
		UUID:        post.UUID,
		Title:       title,
		Description: description,
		Images:      images,
		Votes:       make(map[string]bool),
		Creation:    time.Now().Unix(),
	}

	post.Rating = post.GetRating()

	_, err = db.Exec("INSERT INTO posts (uuid, useruuid, title, description, images, votes, rating, creation) VALUES (?, ?, ?, ?, ?, ?, ?, ?)", post.UUID, userUUID, post.Title, post.Description, imagesJSON, "{}", post.Rating, post.Creation)

	return
}

// SetVote sets a vote on a post.
func SetVote(post models.Post, uuid string, vote bool) (score int, err error) {
	if oldVote, ok := post.Votes[uuid]; ok {
		if vote == oldVote {
			if vote == true {
				post.Upvotes--
			} else {
				post.Downvotes--
			}

			delete(post.Votes, uuid)
		} else {
			if vote == true {
				post.Upvotes++
				post.Downvotes--
			} else {
				post.Upvotes--
				post.Downvotes++
			}

			post.Votes[uuid] = vote
		}
	} else {
		if vote == true {
			post.Upvotes++
		} else {
			post.Downvotes++
		}

		post.Votes[uuid] = vote
	}

	votesJSON, err := json.Marshal(post.Votes)
	if err != nil {
		return
	}

	score = post.Score()
	post.Rating = post.GetRating()

	_, err = db.Exec("UPDATE posts SET votes=?, rating=? WHERE uuid=?", votesJSON, post.Rating, post.UUID)
	return
}
