// Copyright (C) 2019 rameshvk. All rights reserved.
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

// Package gsum summarizes gmails
package gsum

import (
	"encoding/base64"
	"google.golang.org/api/gmail/v1"
	"io/ioutil"
	"mime/quotedprintable"
	"regexp"
	"strings"
	"unicode"
)

// Summarize converts a single gmail message into text
func Summarize(m *gmail.Message, padding string) string {
	p := m.Payload
	if p == nil {
		return m.Snippet
	}
	headers := getHeaders(p.Headers)
	return "From: " + headers["from"] +
		"\nSubject: " + headers["subject"] +
		"\n" + pad(body(p), padding) + "\n"
}

func pad(s, padding string) string {
	if s != "" {
		lines := strings.Split(s, "\n")
		result := []string{}
		for _, l := range lines {
			result = append(result, padding+l)
		}
		s = strings.Join(result, "\n")
	}
	return s
}

func body(p *gmail.MessagePart) string {
	t := strings.ToLower(p.MimeType)

	if t != "" && !strings.Contains(t, "text") {
		if strings.Contains(t, "multipart/alternative") {
			for _, part := range p.Parts {
				if strings.Contains(strings.ToLower(part.MimeType), "text") {
					return body(part)
				}
			}
		}
		return "( " + p.MimeType + ")"
	}

	if p.Body == nil || p.Body.Data == "" {
		result := []string{}
		for _, part := range p.Parts {
			if inner := body(part); inner != "" {
				result = append(result, inner)
			}
		}
		return strings.Join(result, "\n---\n")
	}

	// gmail seems to encode the data string in base64 :(
	data := strings.TrimSpace(p.Body.Data)
	if decoded, err := base64.URLEncoding.DecodeString(data); err == nil {
		data = string(decoded)
	}

	h := getHeaders(p.Headers)
	cte := strings.ToLower(h["content-transfer-encoding"])
	if cte == "quoted-printable" {
		b, err := ioutil.ReadAll(quotedprintable.NewReader(strings.NewReader(data)))
		if err != nil {
			return data
		}
		return strip(string(b))
	}

	if cte != "" && cte != "8bit" {
		return "unknown content-transfer-encoding " + cte
	}

	return strip(data)
}

func getHeaders(hh []*gmail.MessagePartHeader) map[string]string {
	headers := map[string]string{}
	for _, h := range hh {
		headers[strings.ToLower(h.Name)] = h.Value
	}
	return headers
}

func trimPrint(s string) string {
	result := []rune{}
	for _, r := range s {
		if unicode.IsPrint(r) {
			result = append(result, r)
		}
	}
	return strings.TrimSpace(string(result))
}

func strip(s string) string {
	// strip any forwards or embedded emails
	inner := stripInner(s)

	// allow only 10 lines and skip empty lines
	max := 10
	line := regexp.MustCompile("(\r|\n)")
	lines := line.Split(inner, max*3+1)
	result := []string{}
	count := 0
	for _, l := range lines {
		l = trimPrint(l)
		if l != "" && count < max {
			count++
			if len(l) > 80 {
				l = l[:80]
			}
			result = append(result, l)
		}
	}
	return strings.Join(result, "\n")
}

func stripInner(s string) string {
	multiRE := regexp.MustCompile("(^|\n)on[\\s\\S]*wrote.*:($|\r|\n)")
	if loc := multiRE.FindStringIndex(strings.ToLower(s)); loc != nil {
		return strings.TrimSpace(s[:loc[0]])
	}

	line := regexp.MustCompile("(\r|\n)")

	p := line.Split(s, 2)
	match := s

	for {
		if loc := stripRE.FindStringIndex(strings.ToLower(p[0])); loc != nil {
			return strings.TrimSpace(s[:len(s)-len(match)])
		}
		if len(p) == 1 {
			return s
		}

		match = p[1]
		p = line.Split(match, 2)
	}
}

var stripRE = compiled()

func compiled() *regexp.Regexp {
	simpleEmailRE := "\\S+@\\S+"
	return regexp.MustCompile(
		"(" +
			// any line that starts with From: email
			"^>?\\s*from:.*" + simpleEmailRE + ".*$|" +

			// gmail style reply: date <email>
			"^[0-9]{4}/[0-9]{1,2}/[0-9]{1,2} .* <\\s*" + simpleEmailRE + "\\s*>$|" +

			// some email clients just say email@domain.com wrote:
			"^" + simpleEmailRE + ".*wrote:\\s+$|" +

			// forwarded stuff
			"^____+$|" +
			"^.*forwarded\\s+message:$|" +
			"^.*original\\s+message:$|" +
			"^-+\\s+forwarded\\s+message\\s+-+$|" +
			"^-+\\s+original\\s+message\\s+-+$)")

}
