package email

import (
	"bytes"
	"html/template"
	"log"

	"github.com/VolticFroogo/Animal-Pictures/models"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
)

// Register sends the account registry email.
func Register(code, username, email string) (err error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("eu-west-1")},
	)
	if err != nil {
		return
	}

	// Create an SES session.
	svc := ses.New(sess)

	t, err := template.ParseFiles("templates/email/register.html") // Parse the HTML page.
	if err != nil {
		log.Printf("Template parsing error: %v", err)
		return
	}

	variables := models.EmailTemplateVariables{
		Code:     code,
		Username: username,
	}

	var tBytes bytes.Buffer
	err = t.Execute(&tBytes, variables)
	if err != nil {
		log.Printf("Template execution error: %v", err)
		return
	}

	// Assemble the email.
	input := &ses.SendEmailInput{
		Source: aws.String("\"Animal Pictures\" <noreply@froogo.co.uk>"),
		Destination: &ses.Destination{
			ToAddresses: []*string{
				aws.String(email),
			},
		},
		Message: &ses.Message{
			Subject: &ses.Content{
				Charset: aws.String("UTF-8"),
				Data:    aws.String("Register Account"),
			},
			Body: &ses.Body{
				Html: &ses.Content{
					Charset: aws.String("UTF-8"),
					Data:    aws.String(tBytes.String()),
				},
				Text: &ses.Content{
					Charset: aws.String("UTF-8"),
					Data:    aws.String("Welcome " + username + ",\nTo finish the registration process of your account please visit: https://ap.froogo.co.uk/verify?code=" + code + "\nIf you haven't registered an account please just ignore this email, sorry for any inconvenience."),
				},
			},
		},
	}

	// Attempt to send the email.
	_, err = svc.SendEmail(input)
	return
}

// Recovery sends the recovery email.
func Recovery(code, username, email string) (err error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("eu-west-1")},
	)
	if err != nil {
		return
	}

	// Create an SES session.
	svc := ses.New(sess)

	t, err := template.ParseFiles("templates/email/recovery.html") // Parse the HTML page.
	if err != nil {
		log.Printf("Template parsing error: %v", err)
		return
	}

	variables := models.EmailTemplateVariables{
		Code:     code,
		Username: username,
	}

	var tBytes bytes.Buffer
	err = t.Execute(&tBytes, variables)
	if err != nil {
		log.Printf("Template execution error: %v", err)
		return
	}

	// Assemble the email.
	input := &ses.SendEmailInput{
		Source: aws.String("\"Animal Pictures\" <noreply@froogo.co.uk>"),
		Destination: &ses.Destination{
			ToAddresses: []*string{
				aws.String(email),
			},
		},
		Message: &ses.Message{
			Subject: &ses.Content{
				Charset: aws.String("UTF-8"),
				Data:    aws.String("Reset Your Password"),
			},
			Body: &ses.Body{
				Html: &ses.Content{
					Charset: aws.String("UTF-8"),
					Data:    aws.String(tBytes.String()),
				},
				Text: &ses.Content{
					Charset: aws.String("UTF-8"),
					Data:    aws.String("Hello " + username + ",\nTo reset your password please click this link: https://ap.froogo.co.uk/password-recovery/?code=" + code + "\nIf it wasn't you trying to reset your password please just ignore this email, sorry for any inconvenience.\nHowever, if you are receiving lots of these emails please contact support for assistance."),
				},
			},
		},
	}

	// Attempt to send the email.
	_, err = svc.SendEmail(input)
	return
}
