package db

import (
	"time"

	"github.com/zemirco/uid"
)

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
