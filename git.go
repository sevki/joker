package main

import (
	"bufio"
	"bytes"
	"fmt"
)

func diffLine(bz []byte, s, l int) int {
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
			return t
		}
	}
	fmt.Println(t, n)
	return -2
}
