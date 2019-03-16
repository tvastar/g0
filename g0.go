// Copyright (C) 2019 rameshvk. All rights reserved.
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"encoding/base64"	
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"runtime"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"

	"github.com/pkg/browser"
	"github.com/tvastar/g0/digest"
)

const localPort = ":5555"

func localFile(name string) string {
	_, fname, _, _ := runtime.Caller(0)
	return path.Join(path.Dir(fname), name)
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := localFile(os.Args[1] + ".json")
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	config.RedirectURL = "http://localhost" + localPort
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)

	authCode := getAuthCode(authURL)

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

func main() {
	b, err := ioutil.ReadFile(localFile("credentials.json"))
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := gmail.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Gmail client: %v", err)
	}

	user := "me"
	r, err := srv.Users.Messages.List(user).Q("in:inbox is:unread").Do()
	if err != nil {
		log.Fatalf("Unable to retrieve labels: %v", err)
	}
	fmt.Println(len(r.Messages), "unread messaages")
	opt := digest.Options{
		LineLimit: 10,
		ColLimit: 80,
		OmitLinks: true,
	}
	
	for _, m := range r.Messages {
		mm, err := srv.Users.Messages.Get(user, m.Id).Format("RAW").Do()
		if err != nil {
			log.Fatalf("Could not read message", m.Id, err)
		}
		decoded, err := base64.URLEncoding.DecodeString(mm.Raw)
		if err != nil {
			log.Fatalf("Could not decode message", m.Id, err)
		}
		
		digested, err := digest.Message(string(decoded), opt)
		if err != nil {
			log.Fatalf("Could not digest message", m.Id, err)
		}
		fmt.Println(digested + "\n\n")
	}
}

type statusCodeHandler chan string

func (s statusCodeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code != "" {
		s <- code
	}
}

func getAuthCode(url string) string {
	ch := make(chan string, 1)
	srv := &http.Server{Addr: localPort, Handler: statusCodeHandler(ch)}

	go func() {
		// returns ErrServerClosed on graceful close
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			// NOTE: there is a chance that next line won't have time to run,
			// as main() doesn't wait for this goroutine to stop. don't use
			// code with race conditions like these for production. see post
			// comments below on more discussion on how to handle this.
			log.Fatalf("ListenAndServe(): %s", err)
		}
	}()

	defer srv.Shutdown(context.TODO())

	if err := browser.OpenURL(url); err != nil {
		log.Fatalf("Unable to open browser: %v", err)
	}

	return <-ch
}
