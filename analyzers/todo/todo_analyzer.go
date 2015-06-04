// Copyright 2015 Sevki <s@sevki.org>.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package todo analyzes changed files for TODO messages.
package todo

import (
	"regexp"

	"bufio"
	"bytes"

	"sevki.org/joker/git"
	"sourcegraph.com/sourcegraph/go-diff/diff"

	"sevki.org/joker/analyzers"
)

// TodoAnalyzer internals
type analyzer struct {
	msgBuffer analyzers.Message
	msgs      []analyzers.Message
}

// Init implements the analyzer interface
func Init(changeSet git.ChangeSet) analyzers.Scanner {

	var msgs []analyzers.Message
	re := regexp.MustCompile(`TODO\((.*)\): (.*)`)
	// Get changeset
	for _, c := range changeSet {
		// Parse hunks.
		diffs, _ := diff.ParseHunks([]byte(*c.Patch))
		// Iterate trough hunks
		for _, d := range diffs {
			// hunks to a scanner
			scnr := bufio.NewScanner(bytes.NewBuffer(d.Body))
			// LINE number trackers.
			n := d.NewStartLine - 1
			for scnr.Scan() {
				t := scnr.Text()
				// Count line numbers
				switch t[0] {
				case '-':
					break
				default:
					n++
				}
				// End magic

				// If the line is not a new edition, move on.
				if t[0] != '+' {
					continue
				}
				finds := re.FindStringSubmatch(t)
				if len(finds) != 3 {
					continue
				}
				msgs = append(msgs, analyzers.Message{
					Body:     finds[2],
					Filename: *c.Filename,
					Asignee:  finds[1],
					Line:     n,
					DiffLine: 0,
					Issue:    true,
				})
			}
		}
	}
	return &analyzer{msgs: msgs}
}
func init() {
	analyzers.Register("todo", Init)
}
func (j *analyzer) Scan() bool {
	return len(j.msgs) != 0
}
func (j *analyzer) Message() analyzers.Message {
	tmp := j.msgs[0]
	j.msgs = j.msgs[1:]
	return tmp
}
