// Copyright 2015 Sevki <s@sevki.org>.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package jshint is a wrapper around `jshint` and `jsxhint` cmdlets, to
// install them:
//
// 	npm install -g jshint
//
// they should be present in `$PATH`. It is higly advised that you also
// have a `.jshintrc` file.
package jshint

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"sevki.org/joker/analyzers"
)

type JSHintAnalyzer struct {
	scnr      bufio.Scanner
	msgBuffer analyzers.Message
}

func Init() analyzers.Scanner {
	cmd := exec.Command(flag.Args()[0], flag.Args()[1:]...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	return &JSHintAnalyzer{scnr: *bufio.NewScanner(stdout)}
}
func init() {
	analyzers.Register("jshint", Init)
	analyzers.Register("jsxhint", Init)
}
func (j *JSHintAnalyzer) Scan() bool {

	j.scnr.Scan()
	str := j.scnr.Text()

	var msg analyzers.Message
	fmt.Sscanf(str, "%s line %d, col %d, %[^\n]", &msg.Filename, &msg.Line, &msg.Col, &msg.Body)
	msg.Filename = strings.Trim(msg.Filename, ":")
	n := len(fmt.Sprintf("%s line %d, col %d, ", msg.Filename, msg.Line, msg.Col))
	if msg.Line == 0 {
		return false
	}
	msg.Body = str[n:]
	j.msgBuffer = msg
	return true
}
func (j *JSHintAnalyzer) Message() analyzers.Message {
	return j.msgBuffer
}
