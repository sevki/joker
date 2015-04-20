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
	"fmt"
	"log"

	"code.google.com/p/goauth2/oauth"
	"github.com/google/go-github/github"
	"sevki.org/joker/analyzers"
	_ "sevki.org/joker/analyzers/jshint"
	_ "sevki.org/joker/analyzers/todo"
	"sevki.org/joker/git"
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

	commit, _, err :=
		client.Repositories.GetCommit(*owner, *repo, *sha)

	if err != nil {
		log.Fatal(err.Error())
	}
	scanner := analyzers.GetScanner(*scnr, commit.Files)

	for scanner.Scan() {
		comment(scanner.Message(), commit)
	}
}

func comment(msg analyzers.Message, commit *github.RepositoryCommit) {

	if msg.Issue {
		body := fmt.Sprintf(
			"https://github.com/%s/%s/blob/%s/%s#L%d",
			*owner,
			*repo,
			*sha,
			msg.Filename,
			msg.Line,
		)
		_, _, err := client.Issues.Create(
			*owner,
			*repo,
			&github.IssueRequest{
				Title:    &msg.Body,
				Body:     &body,
				Assignee: &msg.Asignee,
				Labels:   []string{"TODO"},
			},
		)
		if err != nil {
			log.Println(err.Error())
		}

	} else {
		msg.DiffLine = git.LineIsNew(commit, msg.Line, msg.Filename)
		if msg.DiffLine < 0 {
			return
		}
		_, _, err := client.Repositories.CreateComment(
			*owner,
			*repo,
			*commit.SHA,
			&github.RepositoryComment{
				Body:     &msg.Body,
				Path:     &msg.Filename,
				Position: &msg.DiffLine,
			})
		if err != nil {
			log.Println(err.Error())
		}
	}
}
