package db

import (
	"time"

	"github.com/VolticFroogo/Animal-Pictures/models"
	"github.com/zemirco/uid"
)

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

// EditSelfEmail updates a user's email after verification.
func EditSelfEmail(uuid string, email string) (err error) {
	_, err = db.Exec("UPDATE users SET email=? WHERE uuid=?", email, uuid)
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
