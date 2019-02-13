package main

import (
	"log"
	"math/rand"
	"time"

	"github.com/VolticFroogo/Animal-Pictures/captcha"
	"github.com/VolticFroogo/Animal-Pictures/db"
	"github.com/VolticFroogo/Animal-Pictures/handler"
	"github.com/VolticFroogo/Animal-Pictures/middleware/myJWT"
	"github.com/VolticFroogo/Animal-Pictures/upload"
)

func main() {
	// Seed the randomiser to prevent repeated seeds and values.
	rand.Seed(time.Now().UTC().UnixNano())

	captcha.Init()

	if err := upload.Init(); err != nil {
		log.Printf("Error initialising uploader: %v", err)
		return
	}

	// Initialise the database.
	if err := db.InitDB(); err != nil {
		log.Printf("Error initialising database: %v", err)
		return
	}

	// Load up the RSA keys.
	if err := myJWT.InitKeys(); err != nil {
		log.Printf("Error initialising JWT keys: %v", err)
		return
	}

	// Start the website handler.
	handler.Start()
}
