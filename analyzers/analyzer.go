// Copyright 2015 Sevki <s@sevki.org>.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package analyzers define how analyzers should act. Analyzers
// implement `scanner` interface. They must also register their
// initfunctions so they can be initialized. see
// https://sevki.org/joker/analyzers/jshint for reference.
package analyzers

import (
	"sevki.org/joker/git"
)

// Message represents a Github comment structure.
type Message struct {
	Body     string
	Filename string
	Line     int32
	DiffLine int
	Asignee  string
	Issue    bool
	// Github doesn't care about this in commits.
	Col int
}

// Scanner interface for the analyzerss
type Scanner interface {
	Scan() bool
	Message() Message
}

var analysers map[string]InitFunc

func init() {
	analysers = make(map[string]InitFunc)
}

// InitFunc type that scanners must implement
type InitFunc func(changeset git.ChangeSet) Scanner

//GetScanner initializes scanner for a fileset.
func GetScanner(scnr string, changeSet git.ChangeSet) Scanner {
	a := analysers[scnr](changeSet)
	return a
}

//Register registers a scanner to be used.
func Register(scnr string, scnrFunc InitFunc) {
	analysers[scnr] = scnrFunc
}
