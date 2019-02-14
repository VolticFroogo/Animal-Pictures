package db

import (
	"log"
	"time"

	"github.com/VolticFroogo/Animal-Pictures/helpers"
	"github.com/VolticFroogo/Animal-Pictures/models"
)

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
