// Copyright (C) 2019 rameshvk. All rights reserved.
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

// Package gsum summarizes gmails
package gsum

import (
	"encoding/base64"
	"errors"
	"io/ioutil"
	"jaytaylor.com/html2text"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/mail"
	"regexp"
	"strings"
)

// Options to configure behavior
type Options struct {
}

// Message takes a raw smtp message string and returns a simplified
// version of it.
//
// It only includes From and Subject among the headers and understands
// transfer encodings like quoted-printable
//
// Options can be used to further control behavior
func Message(rawMessage string, opt Options) (string, error) {
	r := strings.NewReader(rawMessage)
	m, err := mail.ReadMessage(r)
	if err != nil {
		return "", err
	}
	body, err := ioutil.ReadAll(m.Body)
	if err != nil {
		return "", err
	}

	from := m.Header.Get("From")
	if from != "" {
		from = "From: " + from + "\n"
	}
	subject := m.Header.Get("Subject")
	if subject != "" {
		subject = "Subject: " + subject + "\n"
	}

	ct, cte := m.Header.Get("Content-Type"), m.Header.Get("Content-Transfer-Encoding")
	s, err := Body(string(body), ct, cte, opt)
	return from + subject + s, err
}

// Body takes a raw SMTP body and simplifies it.
//
// The body itself should be  the raw body string or optionally a
// base64 encoded version  of it.
//
// Content types can be text/plain or a RFC1341 style multipart.
// The first text part is used for multipart messages.
//
// Options can be used to further control behavior
func Body(body, contentType, transferEncoding string, opt Options) (string, error) {
	// attempt to do some base64-decoding anyway
	if decoded, err := base64.URLEncoding.DecodeString(body); err == nil {
		body = string(decoded)
	}
	if decoded, err := base64.StdEncoding.DecodeString(body); err == nil {
		body = string(decoded)
	}

	if strings.ToLower(transferEncoding) == "quoted-printable" {
		b, err := ioutil.ReadAll(quotedprintable.NewReader(strings.NewReader(body)))
		if err != nil {
			return "", err
		}
		body = string(b)
	}

	ct := strings.ToLower(contentType)
	if strings.Contains(ct, "multipart/") {
		return parseMultipart(body, contentType, opt)
	}

	if strings.Contains(ct, "text/html") {
		body = stripHTML(body)
	}

	return stripEmbedded(body), nil
}

func parseMultipart(body, ct string, opt Options) (string, error) {
	mtype, params, err := mime.ParseMediaType(ct)
	if err != nil {
		return "", err
	}

	boundary := params["boundary"]
	if boundary == "" || !strings.HasPrefix(mtype, "multipart/") {
		return "", errors.New("bad content type: " + ct)
	}

	mr := multipart.NewReader(strings.NewReader(body), boundary)
	results := []string{}
	p, err := mr.NextPart()
	for ; err == nil; p, err = mr.NextPart() {
		if inner, err := ioutil.ReadAll(p); err == nil {
			s := string(inner)
			ct := p.Header.Get("content-type")
			if !strings.HasPrefix(ct, "text/") {
				continue
			}
			if strings.Contains(ct, "text/html") {
				s = stripHTML(s)
			} else {
				s = stripEmbedded(s)
			}
			results = append(results, s)
			continue
		}
		break
	}

	if mtype == "multipart/alternative" && len(results) > 0 {
		return results[0], nil
	}
	return strings.Join(results, "\n"), nil
}

func stripEmbedded(body string) string {
	if loc := stripEmbeddedRE.FindStringIndex(strings.ToLower(body)); loc != nil {
		return strings.TrimSpace(body[:loc[0]])
	}
	return strings.TrimSpace(body)
}

func stripHTML(body string) string {
	htmlOpt := html2text.Options{}
	text, err := html2text.FromString(body, htmlOpt)
	if err != nil {
		return err.Error()
	}
	return text
}

var stripVariants = []string{
	// From: <email>
	"\\s*from:.*[a-z0-9]+@[a-z0-9]+.*",

	// >From: <email>
	">\\s*from:.*[a-z0-9]+@[a-z0-9]+.*",

	// gmail style reply: date <email>
	"^[0-9]{4}/[0-9]{1,2}/[0-9]{1,2} .* <\\s*\\S+@\\S+\\s*>",

	// some email clients just say email@domain.com wrote:
	"^.*[a-z0-9]@[a-z0-9].*wrote:\\s+",

	// on ... wrote:
	"on[\\s\\S]*wrote.*:",

	// forwarded stuff
	"^____+",
	"^.*forwarded\\s+message:",
	"^.*original\\s+message:",
	"^-+\\s+forwarded\\s+message\\s+-+",
	"^-+\\s+original\\s+message\\s+-+",
}

var stripEmbeddedRE = regexp.MustCompile("(^|\n)(" + strings.Join(stripVariants, "|") + ")($|\r|\n)")
