// Copyright (C) 2019 rameshvk. All rights reserved.
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package gsum_test

import (
	"github.com/tvastar/g0/gsum"
	"strings"
	"testing"
)

func TestMessage(t *testing.T) {
	files := []string{
		"smtp1",
	}

	for _, f := range files {
		t.Run(f, func(t *testing.T) {
			inf := file("testdata/"+f+".input.txt", t)
			outf := file("testdata/"+f+".output.txt", t)
			result, err := gsum.Message(inf, gsum.Options{})
			if err != nil {
				t.Fatal("Error", err)
			}
			got := strings.TrimSpace(result)
			want := strings.TrimSpace(outf)
			if got != want {
				t.Error("Expected", want, "got", got)
			}
		})
	}
}
