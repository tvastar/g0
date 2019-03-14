// Copyright (C) 2019 rameshvk. All rights reserved.
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package gsum_test

import (
	"github.com/tvastar/g0/gsum"
	"io/ioutil"
	"path"
	"runtime"
	"strings"
	"testing"

	"google.golang.org/api/gmail/v1"
)

func TestNoPayload(t *testing.T) {
	m := &gmail.Message{Snippet: "snip snip"}
	got := strings.TrimSpace(gsum.Summarize(m, ""))
	want := "snip snip"
	if got != want {
		t.Error("Expected", want, "got", got)
	}
}

func TestStrangeMimeType(t *testing.T) {
	h := []*gmail.MessagePartHeader{{"From", "yo", nil, nil}}
	m := &gmail.Message{Payload: &gmail.MessagePart{
		Headers: h,
		Parts: []*gmail.MessagePart{
			{
				MimeType: "strange",
				Body:     &gmail.MessagePartBody{Data: "snip snip"},
			},
		},
	}}
	got := strings.TrimSpace(gsum.Summarize(m, ""))
	want := "From: yo\nSubject: \n( strange)"
	if got != want {
		t.Error("Expected", want, "got", got)
	}
}

func TestStripFiles(t *testing.T) {
	files := []string{
		"forward1",
		"forward2",
		"forward3",
		"forward4",
		"from",
		"from2",
		"from3",
		"negative",
		"someoneWrote1",
		"someoneWrote2",
	}

	for _, f := range files {
		t.Run(f, func(t *testing.T) {
			inf := file("testdata/"+f+".input.txt", t)
			outf := file("testdata/"+f+".output.txt", t)
			got := strings.TrimSpace(gsum.Summarize(getMessage(inf), ""))
			want := "From: \nSubject: \n" + stripEmpty(outf)
			if got != want {
				t.Error("Expected", want, "got", got)
			}
		})
	}
}

func file(s string, t *testing.T) string {
	_, base, _, _ := runtime.Caller(0)
	f, err := ioutil.ReadFile(path.Join(path.Dir(base), s))
	if err != nil {
		t.Fatal("Could not read", s, err)
	}
	return string(f)
}

func getMessage(s string) *gmail.Message {
	return &gmail.Message{Payload: &gmail.MessagePart{
		Parts: []*gmail.MessagePart{
			{
				Body: &gmail.MessagePartBody{Data: s},
			},
		},
	}}
}

func stripEmpty(s string) string {
	lines := strings.Split(s, "\n")
	result := []string{}
	for _, l := range lines {
		if lx := strings.TrimSpace(l); lx != "" {
			result = append(result, lx)
		}
	}
	return strings.Join(result, "\n")
}
