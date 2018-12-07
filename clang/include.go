package clang

import (
	"regexp"
	"fmt"
)

var (
	RegExpIncludeAny = regexp.MustCompile("^\\s*(#include\\s+(<(.*)>|\"(.*)\"))\\s*$")
	RegExpIncludeUser = regexp.MustCompile("^\\s*(#include\\s+(\"(.*)\"))\\s*$")
	RegExpIncludeSys = regexp.MustCompile("^\\s*(#include\\s+(<(.*)>))\\s*$")
)

const (
	IncludeError = iota
	IncludeUser
	IncludeSys
)

type Include string

func (i Include) Valid() bool {
	return RegExpIncludeAny.MatchString(string(i))
}

func (i Include) IsUser() bool {
	return RegExpIncludeUser.MatchString(string(i))
}

func (i Include) IsSys() bool {
	return RegExpIncludeSys.MatchString(string(i))
}

func (i Include) Kind() int {
	if i.IsUser() {
		return IncludeUser
	} else if i.IsSys() {
		return IncludeSys
	} else {
		return IncludeError
	}
}

func (i Include) Path() (string, error) {

	if m := RegExpIncludeUser.FindStringSubmatch(string(i)); m != nil {
		return m[3], nil
	} else if m := RegExpIncludeSys.FindStringSubmatch(string(i)); m != nil {
		return m[3], nil
	} else {
		return "", fmt.Errorf("path not found")
	}
}