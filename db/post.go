package db

import (
	"encoding/json"
	"time"

	"github.com/VolticFroogo/Animal-Pictures/models"
	"github.com/zemirco/uid"
)

// GetPost returns a post given a UUID.
func GetPost(uuid string) (post models.Post, err error) {
	rows, err := db.Query("SELECT useruuid, title, description, images, creation, votes FROM posts WHERE uuid=?", uuid)
	if err != nil {
		return
	}

	defer rows.Close()

	post.UUID = uuid
	if rows.Next() {
		var imagesJSON, votesJSON string

		err = rows.Scan(&post.UserUUID, &post.Title, &post.Description, &imagesJSON, &post.Creation, &votesJSON) // Scan data from query.
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

		post.SetRating()
	}

	return
}

// NewPost creates a new post.
func NewPost(title, description, userUUID string, images []string) (uuid string, err error) {
	imagesJSON, err := json.Marshal(images)
	if err != nil {
		return
	}

	var exists bool
	for {
		uuid = uid.New(8)
		exists, err = rowExists("SELECT useruuid FROM posts WHERE uuid=?", uuid)
		if err != nil {
			return uuid, err
		}

		if !exists {
			break
		}
	}

	_, err = db.Exec("INSERT INTO posts (uuid, useruuid, title, description, images, creation, votes) VALUES (?, ?, ?, ?, ?, ?, ?)", uuid, userUUID, title, description, imagesJSON, time.Now().Unix(), "{}")

	return
}

// SetVote sets a vote on a post.
func SetVote(post models.Post, uuid string, vote bool) (score int, err error) {
	score = post.Score()

	if oldVote, ok := post.Votes[uuid]; ok {
		if vote == oldVote {
			if vote == true {
				score--
			} else {
				score++
			}

			delete(post.Votes, uuid)
		} else {
			if vote == true {
				score += 2
			} else {
				score -= 2
			}

			post.Votes[uuid] = vote
		}
	} else {
		if vote == true {
			score++
		} else {
			score--
		}

		post.Votes[uuid] = vote
	}

	votesJSON, err := json.Marshal(post.Votes)
	if err != nil {
		return
	}

	_, err = db.Exec("UPDATE posts SET votes=? WHERE uuid=?", votesJSON, post.UUID)
	return
}
