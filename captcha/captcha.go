package captcha

import (
	"os"
	"time"

	"github.com/VolticFroogo/Animal-Pictures/models"
	recaptcha "github.com/dpapathanasiou/go-recaptcha"
)

var (
	// Get our reCAPTCHA secret keys.
	v2Secret   = os.Getenv("CAPTCHA_V2_SECRET")
	v3Secret   = os.Getenv("CAPTCHA_V3_SECRET")
	cautiousIP map[string]int64
)

// Init is called to setup the reCAPTCHA script.
func Init() {
	// Initialise the cautiousIP map.
	cautiousIP = make(map[string]int64)
	go garbageCollector()
}

func garbageCollector() {
	ticker := time.NewTicker(time.Hour) // Tick every hour.
	for {
		<-ticker.C // Tick.

		// Iterate through every cautious IP.
		for ip, creation := range cautiousIP {
			if time.Unix(creation, 0).Add(models.CautiousIPTime).After(time.Now()) {
				// They should be removed from the cautious IP list as it has expired.
				delete(cautiousIP, ip)
			}
		}
	}
}

// V3 returns whether a user should be allowed to continue by checking their v2 or v3 captcha results.
func V3(v2, v3, ip, action string) bool {
	if v3 != "" {
		// User is completing login with a v3 reCAPTCHA.

		if creation, ok := cautiousIP[ip]; ok {
			if time.Unix(creation, 0).Add(models.CautiousIPTime).Before(time.Now()) {
				// They should be removed from the cautious IP list as it has expired.
				delete(cautiousIP, ip)
			} else {
				// User is on the cautiousIP list and hasn't completed the v2 reCAPTCHA.
				// Inform them that they must complete it or we won't let them log in.
				return false
			}
		}

		recaptcha.Init(v3Secret)
		captcha, nErr := recaptcha.Check(ip, v3)
		if nErr != nil {
			return false
		} else if captcha.Score < models.CaptchaScore {
			// The user's score is too low to log in, tell them they need to complete the v2 reCAPTCHA.
			// Add them to the cautious IP list.
			cautiousIP[ip] = time.Now().Unix()
			return false
		}

		if captcha.Action != action {
			return false
		}
	} else if v2 != "" {
		// User just failed the v3 captcha and has to use v2.

		recaptcha.Init(v2Secret)
		captcha, nErr := recaptcha.Confirm(ip, v2)
		if nErr != nil || !captcha {
			return false
		}

		// They have successfully completed the reCAPTCHA v2, remove them from the cautious map.
		if _, ok := cautiousIP[ip]; ok {
			delete(cautiousIP, ip)
		}
	} else {
		// User didn't submit the reCAPTCHA v3 or v2.
		return false
	}

	return true
}
