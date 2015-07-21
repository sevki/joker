// Copyright 2015 Sevki <s@sevki.org>.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// joker reads from io.Reader's and comments on diffs inline on github.
// Usage:
//
//	joker -repo=douchejs \
//		-owner=superjsdev2015 \
//		-token={token} \
//		-commit=`git describe --always` \
//		-scanner=jshint \
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
//
// Available analyzers are golint, todo, jshint
package main // import "sevki.org/joker"

import (
	"flag"
	"fmt"
	"log"

	"code.google.com/p/goauth2/oauth"
	"github.com/google/go-github/github"
	"sevki.org/joker/analyzers"
	_ "sevki.org/joker/analyzers/golint"
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
	pr     = flag.Int("pr", 0, "pullrequest")
	scnr   = flag.String("scanner", "", "scanner")
)

func main() {
	flag.Parse()
	t := &oauth.Transport{
		Token: &oauth.Token{AccessToken: *token},
	}

	client = github.NewClient(t.Client())
	context := fmt.Sprintf("joker-%s", *scnr)
	var commits []github.RepositoryCommit
	if *pr == 0 {
		commit, _, err :=
			client.Repositories.GetCommit(*owner, *repo, *sha)
		if err != nil {
			log.Fatal(err.Error())
		}
		commits = append(commits, *commit)
	} else {

		cmts, _, err := client.PullRequests.ListCommits(*owner, *repo, *pr, nil)
		if err != nil {
			log.Fatal(err.Error())
		}
		for _, c := range cmts {
			statuses, _, err := client.Repositories.ListStatuses(*owner, *repo, *c.SHA, nil)
			skip := false
			for _, status := range statuses {
				if *status.Context == context {
					skip = true
				}
			}
			if skip {
				continue
			}
			commit, _, err :=
				client.Repositories.GetCommit(*owner, *repo, *c.SHA)
			if err != nil {
				log.Fatal(err.Error())
			}

			commits = append(commits, *commit)
		}
	}
	for _, commit := range commits {
		scanner, err := analyzers.GetScanner(*scnr, commit.Files)
		if err != nil {
			log.Fatal(err.Error())
		}
		issuesPosted := 0
		for scanner.Scan() {
			if comment(scanner.Message(), &commit) {
				issuesPosted += 1
			}
		}
		msg := fmt.Sprintf("%s has found %d error(s).", *scnr, issuesPosted)
		url := fmt.Sprintf("https://github.com/%s/%s/commit/%s", *owner, *repo, *commit.SHA)
		state := "success"
		if issuesPosted > 0 {
			state = "error"
		}
		client.Repositories.CreateStatus(*owner,
			*repo,
			*commit.SHA,
			&github.RepoStatus{
				Context:     &context,
				URL:         &url,
				Description: &msg,
				State:       &state,
			})

	}

}

func comment(msg analyzers.Message, commit *github.RepositoryCommit) bool {

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
				Labels:   &[]string{"TODO"},
			},
		)
		if err != nil {
			log.Println(err.Error())
		}

	} else {

		msg.DiffLine = git.LineIsNew(commit, msg.Line, msg.Filename)
		if msg.DiffLine < 0 {
			return false
		}
		log.Println(msg)
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
	return true
}
