package clang

import "testing"

var includeTestCases = []struct {
	name string
	text string
	kind int
	path string
}{
	{
		"user",
		"#include \"a/b.c\"",
		IncludeUser,
		"a/b.c",
	},
	{
		"user (space indented)",
		"    #include \"a/b.c\"",
		IncludeUser,
		"a/b.c",
	},
	{
		"user (tab indented)",
		"\t#include \"a/b.c\"",
		IncludeUser,
		"a/b.c",
	},
	{
		"sys",
		"#include <a/b.c>",
		IncludeSys,
		"a/b.c",
	},
	{
		"sys (space indented)",
		"    #include <a/b.c>",
		IncludeSys,
		"a/b.c",
	},
	{
		"sys (tab indented)",
		"\t#include <a/b.c>",
		IncludeSys,
		"a/b.c",
	},
	{
		"no match (#include misspelled)",
		"#inculde <a/b.c>",
		IncludeError,
		"",
	},
	{
		"no match (square bracket path)",
		"#include [a/b.c]",
		IncludeError,
		"",
	},
	{
		"no match (additional symbols)",
		"#include <a/b.c> d",
		IncludeError,
		"",
	},
}

func TestIncludeValid(t *testing.T) {
	for _, test := range includeTestCases {
		t.Run(test.name, func(t *testing.T) {
			expected := test.kind != IncludeError
			actual := Include(test.text).Valid()
			if actual != expected {
				t.Fatalf("unexpected result: %v", actual)
			}
		})
	}
}

func TestIncludeIsUser(t *testing.T) {
	for _, test := range includeTestCases {
		t.Run(test.name, func(t *testing.T) {
			e := test.kind == IncludeUser
			a := Include(test.text).IsUser()
			if a != e {
				t.Fatalf("unexpected result: %v", a)
			}
		})
	}
}

func TestIncludeIsSys(t *testing.T) {
	for _, test := range includeTestCases {
		t.Run(test.name, func(t *testing.T) {
			e := test.kind == IncludeSys
			a := Include(test.text).IsSys()
			if a != e {
				t.Fatalf("unexpected result: %v", a)
			}
		})
	}
}

func TestIncludeKind(t *testing.T) {
	for _, test := range includeTestCases {
		t.Run(test.name, func(t *testing.T) {
			kind := Include(test.text).Kind()
			if test.kind != IncludeError {
				if kind != test.kind {
					t.Fatalf("unexpected kind: %v", kind)
				}
			}
		})
	}
}

func TestIncludePath(t *testing.T) {
	for _, test := range includeTestCases {
		t.Run(test.name, func(t *testing.T) {
			path, err := Include(test.text).Path()
			if test.kind != IncludeError {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if path != test.path {
					t.Fatalf("unexpected path: %v", path)
				}
			} else {
				if err == nil {
					t.Fatalf("expected error")
				}
			}
		})
	}
}