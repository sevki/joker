package git // import "sevki.org/joker/git"

import (
	"bufio"
	"bytes"

	"github.com/google/go-github/github"
	"sourcegraph.com/sourcegraph/go-diff/diff"
)

// ChangeSet represents a set of files that have changed.
type ChangeSet []github.CommitFile

// DiffLine returns the corresponding DiffLine
func DiffLine(bz []byte, s, l int) int {
	buf := bytes.NewBuffer(bz)
	n := s - 1
	t := 0
	scanner := bufio.NewScanner(buf)
	for scanner.Scan() {
		q := string(scanner.Text())
		switch q[0] {
		case '-':
			break
		default:
			n++
		}
		t++
		// new file and diff line same
		if n == l {
			if q[0] == ' ' {
				return -1
			}
			return t
		}
	}
	return -2
}

// LineNumFromDiff returns the line number corresponding to the  Diff line
func LineNumFromDiff(bz []byte, s, l int) int {
	buf := bytes.NewBuffer(bz)
	n := s
	t := 0
	scanner := bufio.NewScanner(buf)
	for scanner.Scan() {
		q := string(scanner.Text())
		switch q[0] {
		case '-':
			break
		default:
			n++
		}
		t++
		// new file and diff line same
		if n == l {
			if q[0] == ' ' {
				return -1
			}
			return n
		}
	}
	return -2
}

// ChangedFiles returned as a string array
func ChangedFiles(fs []github.CommitFile) (changeset []string) {
	for _, f := range fs {
		changeset = append(changeset, *f.Filename)
	}
	return
}

// LineIsNew returns true if the line
func LineIsNew(commit *github.RepositoryCommit, l int, f string) int {

	// Check IF the file has changed
	for _, k := range commit.Files {

		if *k.Filename == f {
			// Parse the file patch
			pdiff, _ := diff.ParseHunks([]byte(*k.Patch))

			for _, i := range pdiff {
				return DiffLine(i.Body, i.NewStartLine, l)
			}
		}
	}
	return -1
}
