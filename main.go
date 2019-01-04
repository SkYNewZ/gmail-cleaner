package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"

	flags "github.com/jessevdk/go-flags"
)

// AppConfig application configuration
type AppConfig struct {
	Search             []string `short:"s" long:"search" description:"Search criteria" required:"true"`
	Delete             bool     `short:"d" long:"delete" description:"Delete messages ?" required:"false"`
	CredentialFilePath string   `long:"credentials-file" description:"Credentials file path as json for using GmailAPI" required:"false" default:"credentials.json"`
}

// MessageElement represents our custom Message
type MessageElement struct {
	subject string
	date    string
	id      string
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func searchMail(queries []string, service *gmail.Service) []MessageElement {
	// getting all messages corresponding to this criteria
	msgs := []MessageElement{}
	pageToken := ""

	// for each query
	for _, search := range queries {
		log.Printf("Searching messages with \"%s\"", search)

		// list all messages corresponding to this query
		for {
			req := service.Users.Messages.List("me").Q(search)
			if pageToken != "" {
				req.PageToken(pageToken)
			}
			r, err := req.Do()
			if err != nil {
				log.Fatalf("Unable to retrieve messages: %v", err)
			}

			for _, m := range r.Messages {
				// get details for each messages corresponding to query
				msg, err := service.Users.Messages.Get("me", m.Id).Do()
				if err != nil {
					log.Fatalf("Unable to retrieve message %v: %v", m.Id, err)
				}

				// get only required values
				var date = ""
				var subject = ""
				for _, h := range msg.Payload.Headers {
					if h.Name == "Date" {
						date = h.Value
					} else if h.Name == "Subject" {
						subject = h.Value
						break
					}
				}
				fmt.Printf("==> \"%s\" - %s\n", subject, date)
				// append to final list
				msgs = append(msgs, MessageElement{
					date:    date,
					id:      m.Id,
					subject: subject,
				})
			}

			// if end of list, stop execution
			if r.NextPageToken == "" {
				break
			}

			// else, take next page
			pageToken = r.NextPageToken
		}
	}

	log.Printf("%v messages found with these criteria...\n", len(msgs))
	return msgs
}

func deleteMessages(messages []MessageElement, deleteMessage bool, service *gmail.Service) {
	for _, message := range messages {
		if deleteMessage == false {
			log.Printf("Trashing \"%s\"", message.subject)
			req := service.Users.Messages.Trash("me", message.id)
			_, err := req.Do()
			if err != nil {
				log.Fatalf("Unable to trash message: %v", err)
			}
		} else {
			log.Printf("Deleting \"%s\"", message.subject)
			req := service.Users.Messages.Delete("me", message.id)
			err := req.Do()
			if err != nil {
				log.Fatalf("Unable to delete message: %v", err)
			}
		}
	}
}

func main() {

	// config
	var opts AppConfig
	_, err := flags.Parse(&opts)
	if flags.WroteHelp(err) {
		os.Exit(0)
	}
	if err != nil {
		log.Fatalf("Unable to read configuration: %v", err)
	}

	b, err := ioutil.ReadFile(opts.CredentialFilePath)
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, gmail.MailGoogleComScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := gmail.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Gmail client: %v", err)
	}

	messages := searchMail(opts.Search, srv)
	if len(messages) > 0 {
		fmt.Println("Are you sure you want to delete/trash these messages ? (yes/No)")
		var confirmation string
		if _, err := fmt.Scanln(&confirmation); err != nil {
			log.Fatalf("Unable to read response: %v", err)
		}

		if confirmation == "Y" || confirmation == "y" || confirmation == "yes" {
			deleteMessages(messages, opts.Delete, srv)
		} else {
			log.Println("Aborted")
		}
	}
}
