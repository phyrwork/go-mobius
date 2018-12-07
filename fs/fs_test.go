package fs

import (
	"testing"
	"github.com/phyrwork/mobius/store"
	"github.com/cayleygraph/cayley"
	"context"
)

func newStore(t *testing.T) *store.Store {
	g, err := cayley.NewMemoryGraph()
	if err != nil {
		t.Fatal(err)
	}
	return store.New(g)
}

func newFs(t *testing.T) *Fs {
	s := newStore(t)
	r, err := NewRoot(s)
	if err != nil {
		t.Fatalf("error initializing fs: %v", err)
	}
	return &Fs{
		Store: s,
		Root:  r,
	}
}

func TestFindRoot(t *testing.T) {
	s := newStore(t)
	ctx := context.TODO()
	e, err := NewRoot(s)
	if err != nil {
		t.Skipf("error creating root: %v", err)
	}
	var a File
	_, err = FindRoot(ctx, s, &a)
	if err != nil {
		t.Fatalf("error finding root: %v", err)
	}
	if a.IRI != e.IRI {
		t.Fatalf("unexpected IRI: expected %v, got %v", e.IRI, a.IRI)
	}
}

func TestNewFs_NewRoot(t *testing.T) {
	s := newStore(t)
	ctx := context.TODO()
	f, err := NewFs(ctx, s)
	if err != nil {
		t.Fatalf("error creating fs: %v", err)
	}
	if len(f.Root.IRI) == 0 {
		t.Fatalf("root not created")
	}
}

func TestNewFs_FindRoot(t *testing.T) {
	s := newStore(t)
	ctx := context.TODO()
	e, err := NewRoot(s)
	if err != nil {
		t.Skipf("error creating root: %v", err)
	}
	f, err := NewFs(ctx, s)
	if err != nil {
		t.Fatalf("error creating fs: %v", err)
	}
	if f.Root.IRI != e.IRI {
		t.Fatalf("unexpected IRI: expected %v, got %v", e.IRI, f.Root.IRI)
	}
}

func TestFs_Create_Empty(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{
			"one level",
			"a",
		},
		{
			"two levels",
			"a/b",
		},
		{
			"hidden file",
			".a",
		},
		{
			"file extension",
			"a.b",
		},
		{
			"complex path",
			".a/b/c.d",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.TODO()
			fs := newFs(t)
			e, err := fs.Create(ctx, test.path)
			if err != nil {
				t.Fatalf("error creating file: %v", err)
			}
			a, err := fs.Open(ctx, test.path)
			if err != nil {
				t.Fatalf("error opening file: %v", err)
			}
			if a.IRI != e.IRI {
				t.Fatalf("unexpected file: expected %v, got %v", e.IRI, a.IRI)
			}
		})
	}
}
