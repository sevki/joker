// Copyright 2015 Sevki <s@sevki.org>.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package golint is a go lint analyzer.
package golint

import (
	"log"

	"path"

	"io/ioutil"
	"os"

	"github.com/golang/lint"
	"sevki.org/joker/analyzers"
	"sevki.org/joker/git"
)

type analyzer struct {
	msgs []analyzers.Message
}

var (
	context = "joker-golint"
)

// Init implements the analyzer interface
func Init(changeSet git.ChangeSet) analyzers.Scanner {

	filtered := make(map[string][]byte)
	for _, k := range changeSet {
		switch path.Ext(*k.Filename) {
		case path.Ext("main.go"):
			f, err := os.Open(*k.Filename)
			if err != nil {
				log.Fatal(err)
			}
			bytes, err := ioutil.ReadAll(f)
			if err != nil {
				log.Fatal(err)
			}
			filtered[*k.Filename] = bytes
		}
	}
	var linter lint.Linter
	l := &linter

	probs, err := l.LintFiles(filtered)
	if err != nil {
		log.Fatal(err)
	}
	var msgs []analyzers.Message
	for _, prob := range probs {
		msgs = append(msgs, analyzers.Message{
			Body:     prob.Text,
			Filename: prob.Position.Filename,
			Line:     int32(prob.Position.Line),
			DiffLine: 0,
			Issue:    false,
		})
	}

	return &analyzer{msgs: msgs}
}
func init() {
	analyzers.Register("golint", Init)
}
func (j *analyzer) Scan() bool {
	return len(j.msgs) != 0
}
func (j *analyzer) Message() analyzers.Message {
	tmp := j.msgs[0]
	j.msgs = j.msgs[1:]
	return tmp
}
