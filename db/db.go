package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/VolticFroogo/Animal-Pictures/helpers"
	"github.com/VolticFroogo/Animal-Pictures/models"
	_ "github.com/go-sql-driver/mysql" // Necessary for connecting to MySQL.
	"github.com/zemirco/uid"
)

/*
	Structs and variables
*/

// Database connection information.
var (
	// Type of database.
	Type = "mysql"
	// Username to access database.
	Username = "ap"
	// Password to access database.
	Password = os.Getenv("DB_PASSWORD")
	// Protocol of database.
	Protocol = "unix"
	// FileLocation of database socket.
	FileLocation = "/var/run/mysqld/mysqld.sock"
	// Database location.
	Database = "ap"
	// ConnString to connect to database.
	ConnString = Username + ":" + Password + "@" + Protocol + "(" + FileLocation + ")/" + Database
)

var (
	db *sql.DB
)

// InitDB initializes the Database.
func InitDB() (err error) {
	db, err = sql.Open(Type, ConnString)
	if err != nil {
		return
	}

	go jtiGarbageCollector()
	return
}

/*
	Helper functions
*/

func rowExists(query string, args ...interface{}) (exists bool, err error) {
	query = fmt.Sprintf("SELECT exists (%s)", query)
	err = db.QueryRow(query, args...).Scan(&exists)
	return
}

/*
	MySQL DataBase related functions
*/

// StoreRefreshToken generates, stores and then returns a JTI.
func StoreRefreshToken(uuid string) (jti models.JTI, err error) {
	// No need to duplication check as the JTI's don't need to be completely unique.
	jti.JTI, err = helpers.GenerateRandomString(32)
	if err != nil {
		return
	}

	jti.Expiry = time.Now().Add(models.RefreshTokenValidTime).Unix()

	_, err = db.Exec("INSERT INTO jti (jti, useruuid, expiry) VALUES (?, ?, ?)", jti.JTI, uuid, jti.Expiry)
	if err != nil {
		return
	}

	rows, err := db.Query("SELECT id FROM jti WHERE jti=? AND useruuid=? AND expiry=?", jti.JTI, uuid, jti.Expiry)
	if err != nil {
		return
	}

	defer rows.Close()

	rows.Next()
	err = rows.Scan(&jti.ID) // Scan data from query.
	return
}

// GetJTI takes a JTI string and returns the JTI struct.
func GetJTI(jti string) (jtiStruct models.JTI, err error) {
	rows, err := db.Query("SELECT id, useruuid, expiry FROM jti WHERE jti=?", jti)
	if err != nil {
		return
	}

	defer rows.Close()

	jtiStruct.JTI = jti
	rows.Next()
	err = rows.Scan(&jtiStruct.ID, &jtiStruct.UserUUID, &jtiStruct.Expiry) // Scan data from query.
	return
}

// CheckJTI returns the validity of a JTI.
func CheckJTI(jti models.JTI) (valid bool, err error) {
	if jti.Expiry > time.Now().Unix() { // Check if token has expired.
		return true, nil // Token is valid.
	}

	_, err = db.Exec("DELETE FROM jti WHERE id=?", jti.ID)
	if err != nil {
		return false, err
	}

	return false, nil // Token is invalid.
}

// DeleteJTI deletes a JTI based on a jti key.
func DeleteJTI(jti string) (err error) {
	_, err = db.Exec("DELETE FROM jti WHERE jti=?", jti)
	return
}

// DeAuthUser completely removes all of a user's JTI tokens therefore deauthorising them.
func DeAuthUser(uuid string) (err error) {
	_, err = db.Exec("DELETE FROM jti WHERE useruuid=?", uuid)
	return
}

func jtiGarbageCollector() {
	ticker := time.NewTicker(time.Hour) // Tick every hour.
	for {
		<-ticker.C
		rows, err := db.Query("SELECT id, expiry FROM jti")
		if err != nil {
			log.Printf("Error querying JTI DB in JTI garbage collector: %v", err)
			return
		}

		defer rows.Close()

		jti := models.JTI{} // Create struct to store a JTI in.
		for rows.Next() {
			err = rows.Scan(&jti.ID, &jti.Expiry) // Scan data from query.
			if err != nil {
				log.Printf("Error scanning rows in JTI garbage collector: %v", err)
				return
			}

			_, err := CheckJTI(jti)
			if err != nil {
				log.Printf("Error checking in JTI garbage collector: %v", err)
				return
			}
		}
	}
}

// GetUserFromUUID retrieves a user from the MySQL database.
func GetUserFromUUID(uuid string) (user models.User, err error) {
	rows, err := db.Query("SELECT email, password, username, privilege, creation, fname, lname, description, imageExtension FROM users WHERE uuid=?", uuid)
	if err != nil {
		return
	}

	defer rows.Close()

	user.UUID = uuid
	for rows.Next() {
		err = rows.Scan(&user.Email, &user.Password, &user.Username, &user.Privilege, &user.Creation, &user.Fname, &user.Lname, &user.Description, &user.ImageExtension) // Scan data from query.
		if err != nil {
			return
		}
	}

	return
}

// GetUserFromEmail retrieves a user from the MySQL database.
func GetUserFromEmail(email string) (user models.User, err error) {
	rows, err := db.Query("SELECT uuid, password, username, privilege, creation, fname, lname, description, imageExtension FROM users WHERE email=?", email)
	if err != nil {
		return
	}

	defer rows.Close()

	user.Email = email
	for rows.Next() {
		err = rows.Scan(&user.UUID, &user.Password, &user.Username, &user.Privilege, &user.Creation, &user.Fname, &user.Lname, &user.Description, &user.ImageExtension) // Scan data from query.
		if err != nil {
			return
		}
	}

	return
}

// UserExistsFromEmail checks if a user exists from an email.
func UserExistsFromEmail(email string) (bool, error) {
	return rowExists("SELECT uuid FROM users WHERE email=?", email)
}

// EditUser updates a user.
func EditUser(uuid, email, password, username string, privilege int) (err error) {
	_, err = db.Exec("UPDATE users SET email=?, password=?, username=?, privilege=? WHERE uuid=?", email, password, username, privilege, uuid)
	return
}

// EditUserNoPassword updates a user without changing the password.
func EditUserNoPassword(ID int, Email, Fname, Lname string, Privileges int) (err error) {
	_, err = db.Exec("UPDATE users SET email=?, fname=?, lname=?, privilege=? WHERE uuid=?", Email, Fname, Lname, Privileges, ID)
	return
}

// EditSelf updates a user from settings.
func EditSelf(uuid, password, username string) (err error) {
	_, err = db.Exec("UPDATE users SET username=?, password=? WHERE uuid=?", username, password, uuid)
	return
}

// NewUser creates a new user.
func NewUser(email, password, username string, privilege int) (uuid string, err error) {
	for {
		uuid = uid.New(8)

		exists, lErr := rowExists("SELECT email FROM users WHERE uuid=?", uuid)
		if lErr != nil {
			return "", lErr
		}

		if !exists {
			break
		}
	}

	_, err = db.Exec("INSERT INTO users (uuid, email, password, username, privilege, creation, fname, lname, description, imageExtension) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", uuid, email, password, username, privilege, time.Now().Unix(), "", "", "", "")
	return
}

// DeleteUser deletes a user.
func DeleteUser(uuid string) (err error) {
	_, err = db.Exec("DELETE FROM users WHERE uuid=?", uuid)
	return
}

// AddEmailVerification adds an email verification code to the DB.
func AddEmailVerification(userUUID, email string) (uuid string, err error) {
	exists, err := rowExists("SELECT uuid FROM email WHERE useruuid=?", userUUID)
	if err != nil {
		return
	}
	if exists {
		_, err = db.Exec("DELETE FROM email WHERE useruuid=?", userUUID)
		if err != nil {
			return
		}
	}

	for {
		uuid = uid.New(8)
		exists, err = rowExists("SELECT useruuid FROM email WHERE uuid=?", uuid)
		if err != nil {
			return uuid, err
		}

		if !exists {
			break
		}
	}

	_, err = db.Exec("INSERT INTO email (uuid, useruuid, email) VALUES (?, ?, ?)", uuid, userUUID, email)
	return
}

// GetEmailVerification retrieves an email verification information.
func GetEmailVerification(uuid string) (userUUID, email string, err error) {
	rows, err := db.Query("SELECT useruuid, email FROM email WHERE uuid=?", uuid)
	if err != nil {
		return
	}

	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&userUUID, &email)
		if err != nil {
			return
		}

		if userUUID != "" && email != "" {
			_, err = db.Exec("DELETE FROM email WHERE uuid=?", uuid)
		}
	}

	return
}

// EditSelfEmail updates a user's email after verification.
func EditSelfEmail(uuid string, email string) (err error) {
	_, err = db.Exec("UPDATE users SET email=? WHERE uuid=?", email, uuid)
	return
}

// AddRecovery adds a password recovery code to the DB.
func AddRecovery(userUUID, email string) (uuid string, err error) {
	exists, err := rowExists("SELECT uuid FROM recovery WHERE useruuid=?", userUUID)
	if err != nil {
		return
	}
	if exists {
		_, err = db.Exec("DELETE FROM recovery WHERE useruuid=?", userUUID)
		if err != nil {
			return
		}
	}

	for {
		uuid = uid.New(8)
		exists, err = rowExists("SELECT useruuid FROM recovery WHERE uuid=?", uuid)
		if err != nil {
			return uuid, err
		}

		if !exists {
			break
		}
	}

	_, err = db.Exec("INSERT INTO recovery (uuid, useruuid, email, creation) VALUES (?, ?, ?, ?)", uuid, userUUID, email, time.Now().Unix())
	return
}

// GetRecovery retrieves a password recovery code from the DB.
func GetRecovery(uuid string) (userUUID, email string, err error) {
	rows, err := db.Query("SELECT useruuid, email FROM recovery WHERE uuid=?", uuid)
	if err != nil {
		return
	}

	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&userUUID, &email)
		if err != nil {
			return
		}

		_, err = db.Exec("DELETE FROM recovery WHERE uuid=?", uuid)
	}

	return
}

// GetRecoveryFromUser gets the recovery of a given user (if one exists).
func GetRecoveryFromUser(userUUID string) (uuid, email string, creation int64, err error) {
	rows, err := db.Query("SELECT uuid, email, creation FROM recovery WHERE useruuid=?", userUUID)
	if err != nil {
		return
	}

	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&uuid, &email, &creation)
	}

	return
}

// EditPassword updates a user's password.
func EditPassword(uuid, password string) (err error) {
	_, err = db.Exec("UPDATE users SET password=? WHERE uuid=?", password, uuid)
	return
}

// EditPrivilege updates a user's privilege.
func EditPrivilege(uuid string, privilege int) (err error) {
	_, err = db.Exec("UPDATE users SET privilege=? WHERE uuid=?", privilege, uuid)
	return
}

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
