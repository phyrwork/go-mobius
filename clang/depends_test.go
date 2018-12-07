package clang

import (
	"testing"
	"github.com/phyrwork/mobius/store"
	"github.com/cayleygraph/cayley"
	"github.com/phyrwork/mobius/fs"
	"context"
	"github.com/cayleygraph/cayley/quad"
	"reflect"
)

func NewStore(t *testing.T) *store.Store {
	g, err := cayley.NewMemoryGraph()
	if err != nil {
		t.Fatal(err)
	}
	return store.New(g)
}

func NewFs(t *testing.T) *fs.Fs {
	s := NewStore(t)
	r, err := fs.NewRoot(s)
	if err != nil {
		t.Fatalf("error initializing fs: %v", err)
	}
	return &fs.Fs{
		Store: s,
		Root:  r,
	}
}

func TestResolveInclude(t *testing.T) {
	tests := []struct {
		name     string
		src      string
		includes []string
		files    []string
		path     []string
		depends  []string
		err      bool
	}{
		{
			"on path",
			"a",
			[]string{"#include <b>"},
			[]string{"b"},
			[]string{"."},
			[]string{"b"},
			false,
		},
		{
			"beside path",
			"a/b",
			[]string{"#include <../b>"},
			[]string{"b"},
			[]string{"a"},
			[]string{"b"},
			false,
		},
		{
			"below path",
			"a",
			[]string{"#include <b/c>"},
			[]string{"b/c"},
			[]string{"."},
			[]string{"b/c"},
			false,
		},
		{
			"not on path",
			"a",
			[]string{"#include <b>"},
			[]string{"b", "a/d"},
			[]string{"a/d"},
			[]string{},
			true,
		},
		{
			"no path",
			"a",
			[]string{"#include <b>"},
			[]string{"b", "a/d"},
			[]string{},
			[]string{},
			true,
		},
		{
			"no file",
			"a",
			[]string{"#include <b>"},
			[]string{},
			[]string{"."},
			[]string{},
			true,
		},
		{
			"ambiguous",
			"a",
			[]string{"#include <d>"},
			[]string{"b/d", "c/d"},
			[]string{"b", "c"},
			[]string{},
			true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.TODO()
			// Setup
			s := NewFs(t)
			includes := make([]Include, len(test.includes))
			for n, i := range test.includes {
				includes[n] = Include(i)
			}
			_, err := s.Create(ctx, test.src)
			if err != nil {
				t.Skipf("error creating test file %v: %v", test.src, err)
			}
			for _, p := range test.files {
				_, err := s.Create(ctx, p)
				if err != nil {
					t.Skipf("error creating test file %v: %v", p, err)
				}
			}
			e := make(map[quad.Value]struct{})
			for _, p := range test.depends {
				node, err := s.Lookup(ctx, nil, p)
				if err != nil {
					t.Skipf("error looking up test file %v: %v", p, err)
				}
				e[node] = struct{}{}
			}
			p, errs := IncludePath(s, test.path...)
			if errs != nil {
				t.Skipf("error initializing include path: %v", errs)
			}
			// Test
			a, errs := ResolveInclude(ctx, p, includes...)
			if (len(e) > 0 || len(a) > 0) && !reflect.DeepEqual(e, a) { // DeepEqual doesn't like two len() = 0
				t.Fatalf("unexpected resolutions: expected %v, got %v", e, a)
			}
			if len(errs) > 0 != test.err {
				t.Fatalf("unexpected error status: %v", errs)
			}
		})
	}
}

