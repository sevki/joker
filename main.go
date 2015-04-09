// Copyright 2015 Sevki <s@sevki.org>.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// joker reads from io.Reader's and comments on diffs inline on github.
// Usage:
//
//	joker -repo=douchejs \
//		-owner=superjsdev2015 \
//		-token={token}
//		-commit=`git describe --always` \
//		-scanner=jshint
//		jsxhint --harmony .
//
// You can add more analyzers, checkout sevki.org/joker/analyzers
// for interface definition for analyzers.
//
// http://sevki.org/joker/analyzers/jshint is a reference
// implementation of what a analyzer should look like.
//
// This app should be run by a CI after you've pushed your changes
// because its sole function is to comment on diffs.
//
// Create a token by going to http://github.com/settings/tokens/new
package main // import "sevki.org/joker"

import (
	"flag"
	"log"

	"code.google.com/p/goauth2/oauth"
	"github.com/google/go-github/github"
	"golang.org/x/codereview/patch"
	"sevki.org/joker/analyzers"
	_ "sevki.org/joker/analyzers/jshint"
)

var (
	client *github.Client
	repo   = flag.String("repo", "", "repo")
	owner  = flag.String("owner", "", "owner")
	token  = flag.String("token", "", "token")
	sha    = flag.String("commit", "deadbeef", "commit")
	scnr   = flag.String("scanner", "", "scanner")
)

func main() {
	flag.Parse()
	t := &oauth.Transport{
		Token: &oauth.Token{AccessToken: *token},
	}

	client = github.NewClient(t.Client())

	commit, _, err := client.Repositories.GetCommit(*owner, *repo, *sha)
	if err != nil {
		log.Fatal(err.Error())
	}
	scanner := analyzers.GetScanner(*scnr)

	for scanner.Scan() {
		comment(scanner.Message(), commit)
	}
}

func comment(msg analyzers.Message, commit *github.RepositoryCommit) {
	// Check IF the file has changed
	for _, k := range commit.Files {
		if *k.Filename == msg.Filename {
			// Parse the file patch
			text, _ := patch.ParseTextDiff([]byte(*k.Patch))
			for _, i := range text {
				//  Check if the line has changed
				if i.Line == msg.Line {
					goto POST
				}
			}
		}
	}
	return
POST:

	_, _, err := client.Repositories.CreateComment(
		*owner,
		*repo,
		*commit.SHA,
		&github.RepositoryComment{
			Body:     &msg.Body,
			Path:     &msg.Filename,
			Position: &msg.Line,
		})
	if err != nil {
		log.Println(err.Error())
	}
}
