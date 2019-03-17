// Copyright (C) 2019 rameshvk. All rights reserved.
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/tvastar/g0"
)

func main() {
	_, self, _, _ := runtime.Caller(0)
	tokenFile := path.Join(path.Dir(path.Dir(self)), os.Args[1]+".json")
	configFile := path.Join(path.Dir(path.Dir(self)), "credentials.json")
	digests, err := g0.Digests(configFile, tokenFile, ":5555")
	if err != nil {
		panic(err)
	}
	fmt.Println(len(digests), "unread messages")
	fmt.Println(strings.Join(digests, "\n\n"))
	if err := g0.MarkRead(configFile, tokenFile, ":5555"); err != nil {
		panic(err)
	}
}
