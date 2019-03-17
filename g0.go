// Copyright (C) 2019 rameshvk. All rights reserved.
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

// Package g0 implements a simple gmail library for zero-inbox people.
package g0

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"

	"github.com/pkg/browser"
	"github.com/tvastar/g0/digest"
)

// Digests returns the digests of all unread inbox messages
func Digests(configFile, tokenFile, localPort string) ([]string, error) {
	tok, toksrc, err := getClientToken(configFile, tokenFile, localPort)
	if err != nil {
		return nil, err
	}

	defer saveTokenFile(tokenFile, tok)

	opt1 := option.WithScopes(gmail.GmailReadonlyScope)
	opt2 := option.WithTokenSource(toksrc)
	srv, err := gmail.NewService(context.Background(), opt1, opt2)
	if err != nil {
		return nil, err
	}

	user := "me"
	r, err := srv.Users.Messages.List(user).Q("in:inbox is:unread").Do()
	if err != nil {
		return nil, err
	}

	results := []string{}

	opt := digest.Options{
		LineLimit: 10,
		ColLimit:  80,
		OmitLinks: true,
	}

	for _, m := range r.Messages {
		mm, err := srv.Users.Messages.Get(user, m.Id).Format("RAW").Do()
		if err != nil {
			return nil, err
		}
		decoded, err := base64.URLEncoding.DecodeString(mm.Raw)
		if err != nil {
			return nil, err
		}

		digested, err := digest.Message(string(decoded), opt)
		if err != nil {
			return nil, err
		}
		results = append(results, digested)
	}
	return results, nil
}

func getClientToken(configFile, tokenFile, localPort string) (*oauth2.Token, oauth2.TokenSource, error) {
	var config *oauth2.Config
	bytes, err := ioutil.ReadFile(configFile) // nolint
	if err == nil {
		config, err = google.ConfigFromJSON(bytes, gmail.GmailReadonlyScope)
	}
	if err != nil {
		return nil, nil, err
	}

	ctx := context.Background()
	tok := &oauth2.Token{}
	if bytes, err = ioutil.ReadFile(tokenFile); err == nil { // nolint
		err = json.NewDecoder(strings.NewReader(string(bytes))).Decode(tok)
	} else {
		config.RedirectURL = "http://localhost" + localPort
		authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
		tok, err = config.Exchange(ctx, getAuthCode(authURL, localPort))
	}

	return tok, config.TokenSource(ctx, tok), err
}

// MarkRead marks all unread inbox messages as read
func MarkRead(configFile, tokenFile, localPort string) error {
	// NYI
	return nil
}

func saveTokenFile(path string, token *oauth2.Token) {
	log.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	must(err)
	if err == nil {
		must(json.NewEncoder(f).Encode(token))
		must(f.Close())
	}
}

type statusCodeHandler chan string

func (s statusCodeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code != "" {
		s <- code
	}
}

func getAuthCode(url, localPort string) string {
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

	defer func() { must(srv.Shutdown(context.Background())) }()

	if err := browser.OpenURL(url); err != nil {
		log.Fatalf("Unable to open browser: %v", err)
	}

	return <-ch
}

func must(err error) {
	if err != nil {
		log.Println(err)
	}
}
