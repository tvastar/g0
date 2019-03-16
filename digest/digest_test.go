// Copyright (C) 2019 rameshvk. All rights reserved.
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package digest_test

import (
	"flag"
	"github.com/tvastar/g0/digest"
	"io/ioutil"
	"log"
	"path"
	"runtime"
	"strings"
	"testing"
)

func TestMessage(t *testing.T) {
	files := map[string]string{
		"smtp1.input.txt":         "smtp1.output.txt",
		"forward1.input.txt":      "forward1.output.txt",
		"forward2.input.txt":      "forward2.output.txt",
		"forward3.input.txt":      "forward3.output.txt",
		"forward4.input.txt":      "forward4.output.txt",
		"from.input.txt":          "from.output.txt",
		"from2.input.txt":         "from2.output.txt",
		"from3.input.txt":         "from3.output.txt",
		"negative.input.txt":      "negative.output.txt",
		"someoneWrote1.input.txt": "someoneWrote1.output.txt",
		"someoneWrote2.input.txt": "someoneWrote2.output.txt",
	}

	opts := digest.Options{
		LineLimit: 20,
		ColLimit:  80,
		OmitLinks: true,
	}

	for before, after := range files {
		testFile(t, before, after, func(input string) (string, error) {
			if !strings.HasPrefix(before, "smtp1") {
				return digest.Body(input, "text/plain", "", opts)
			}
			return digest.Message(input, opts)
		})
	}
}

func testFile(t *testing.T, input, golden string, fn func(string) (string, error)) {
	_, caller, _, _ := runtime.Caller(1)

	t.Run(input+"=>"+golden, func(t *testing.T) {
		read := func(s string) string {
			s = path.Join(path.Dir(caller), "testdata/"+s)
			bytes, err := ioutil.ReadFile(s)
			if err != nil {
				t.Fatal("Could not read", s, err)
			}
			return string(bytes)
		}

		got, err := fn(read(input))
		if err != nil {
			t.Fatal("Error", err)
		}

		if *goldenFlag {
			s := path.Join(path.Dir(caller), "testdata/"+golden)
			if err := ioutil.WriteFile(s, []byte(got), 07); err != nil {
				t.Error("Could not save golden output", s, err)
			}
			log.Println("Saved output to", s)
		} else if expected := read(golden); expected != got {
			t.Error("Unexpected", golden)
		}
	})
}

var goldenFlag = flag.Bool("golden", false, "build golden files instead of verifying")
