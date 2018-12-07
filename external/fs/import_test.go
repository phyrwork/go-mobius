package fs

import (
	"testing"
	"regexp"
	"github.com/phyrwork/mobius/filter"
	"github.com/spf13/afero"
	"github.com/phyrwork/mobius/store"
	"github.com/cayleygraph/cayley"
	"github.com/phyrwork/mobius/fs"
	"context"
)

func newStore(t *testing.T) *store.Store {
	g, err := cayley.NewMemoryGraph()
	if err != nil {
		t.Fatal(err)
	}
	return store.New(g)
}

func newFs(ctx context.Context, t *testing.T) *fs.Fs {
	s := newStore(t)
	f, err := fs.NewFs(ctx, s)
	if err != nil {
		t.Skipf("error creating fs: %v", err)
	}
	return f
}

func TestImporter_Import(t *testing.T) {
	tests := []struct {
		name string
		filter filter.Filter
		files map[string]bool
	}{
		{
			"default",
			nil,
			map[string]bool{
				".git/a/b.c": true, // Because the Git index is definitely structured like this... 😏
				".git/a/b.h": true,
				"a/b.c":      true,
				"a/b.h":      true,
			},
		},
		{
			"blacklist",
			filter.NorFilter{List: []filter.Filter{
				filter.RegexpFilter{Regexp: regexp.MustCompile("\\.git.*")},
			}},
			map[string]bool{
				".git/a/b.c": false,
				".git/a/b.h": false,
				"a/b.c":      true,
				"a/b.h":      true,
			},
		},
		{
			"whitelist",
			filter.OrFilter{List: []filter.Filter{
				filter.RegexpFilter{Regexp: regexp.MustCompile(".*\\.c")},
			}},
			map[string]bool{
				".git/a/b.c": true,
				".git/a/b.h": false,
				"a/b.c":      true,
				"a/b.h":      false,
			},
		},
		{
			// Test a typical Importer use case - exclude Git index, whitelist a subset of working set files
			"complex",
			filter.AndFilter{List: []filter.Filter{
				filter.NotFilter{Filt: filter.RegexpFilter{Regexp: regexp.MustCompile("\\.git.*")}},
				filter.RegexpFilter{Regexp: regexp.MustCompile(".*\\.c")},
			}},
			map[string]bool{
				".git/a/b.c": false,
				".git/a/b.h": false,
				"a/b.c":      true,
				"a/b.h":      false,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Setup
			ctx := context.TODO()
			io := afero.NewMemMapFs()
			for path := range test.files {
				afero.WriteFile(io, path, []byte{}, 0644)
			}
			s := newFs(ctx, t)
			// Test
			im := NewImporter(io, "")
			if test.filter != nil {
				im.Filter = test.filter
			}
			im.Import(nil, s)
			for path, e := range test.files {
				node, err := s.Lookup(ctx, nil, path)
				if err != nil {
					t.Skipf("file lookup error: %v", err)
				}
				a := node != nil
				if a != e {
					t.Fatalf("unexpected file lookup %v: expected %v, got %v", path, e, a)
				}
			}
		})
	}
}

