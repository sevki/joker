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
	"path"
	"strings"

	"sevki.org/joker/analyzers"
	"sevki.org/joker/git"
)

type analyzer struct {
	scnr      bufio.Scanner
	msgBuffer analyzers.Message
}

// Init implements the analyzer interface
func Init(changeSet git.ChangeSet) analyzers.Scanner {
	var filtered []string
	for _, k := range changeSet {
		switch path.Ext(*k.Filename) {
		case path.Ext("main.js"):
			filtered = append(filtered, *k.Filename)
		case path.Ext("app.jsx"):
			filtered = append(filtered, *k.Filename)
		}
	}
	args := append(flag.Args()[1:], filtered...)
	cmd := exec.Command(flag.Args()[0], args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	return &analyzer{scnr: *bufio.NewScanner(stdout)}
}
func init() {
	analyzers.Register("jshint", Init)
	analyzers.Register("jsxhint", Init)
}
func (j *analyzer) Scan() bool {

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
func (j *analyzer) Message() analyzers.Message {
	return j.msgBuffer
}
