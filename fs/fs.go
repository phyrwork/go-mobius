package fs

import (
	"github.com/cayleygraph/cayley/quad"
	"github.com/phyrwork/mobius/store"
	"github.com/cayleygraph/cayley/graph/path"
	"context"
	"fmt"
	"path/filepath"
)

const (
	Basename = quad.IRI("fs:name")
	Dir      = quad.IRI("fs:dir")
	Root     = quad.IRI("fs:root")
)

var (
	RootMorphism = path.StartMorphism().Has(quad.IRI("rdf:type"), quad.IRI(Root))
	UpMorphism   = path.StartMorphism().Out(Dir)
)

func DownMorphism(basenames ...string) *path.Path {
	values := make([]quad.Value, len(basenames))
	for i, basename := range basenames {
		values[i] = quad.String(basename)
	}
	return path.StartMorphism().In(Dir).Has(Basename, values...)
}

type File struct {
	IRI  quad.IRI `quad:"@id"`
	Name string   `quad:"fs:name"`
	Dir  *File    `quad:"fs:dir,opt"`
}

type root struct {
	rdfType struct{} `quad:"@type > fs:root"`
	File
}

func NewRoot(store *store.Store) (File, error) {
	r := root{}
	r.IRI = store.GenerateIRI(nil)
	r.Name = "."
	_, err := store.Insert(r)
	if err != nil {
		return File{}, err
	}
	return r.File, nil
}

func FindRoot(ctx context.Context, store *store.Store, dst interface{}) (quad.Value, error) {
	v, err := path.StartPath(store.Graph).Follow(RootMorphism).Iterate(ctx).FirstValue(nil)
	if err != nil || v == nil {
		return v, err
	}
	return v, store.Select(ctx, dst, v)
}

type Fs struct {
	Store *store.Store
	Root  File
}

func NewFs(ctx context.Context, store *store.Store) (*Fs, error) {
	var root File
	v, err := FindRoot(ctx, store, &root)
	if err != nil {
		return nil, fmt.Errorf("error finding fs root: %v", err)
	}
	if v == nil {
		root, err = NewRoot(store)
		if err != nil {
			return nil, fmt.Errorf("error creating fs root: %v", err)
		}
	}
	return &Fs{
		Store: store,
		Root:  root,
	}, nil
}

func (fs *Fs) Lookup(ctx context.Context, dst interface{}, path string) (node quad.Value, err error) {
	node, err = Path(path).StartPath(fs.Store.Graph, fs.Root.IRI).Iterate(ctx).FirstValue(nil)
	if err != nil {
		return
	}
	if node != nil && dst != nil {
		if err = fs.Store.Select(ctx, dst, node); err != nil {
			err = fmt.Errorf("error getting node %v: %v", node, err)
			return
		}
	}
	return
}

func (fs *Fs) Open(ctx context.Context, path string) (f File, err error) {
	node, err := fs.Lookup(ctx, &f, path)
	if err != nil {
		return
	}
	if node == nil {
		err = fmt.Errorf("file %v not found", path)
		return
	}
	return
}

func (fs *Fs) Create(ctx context.Context, path string) (f File, err error) {
	node, err := fs.Lookup(ctx, &f, path)
	if err != nil {
		return
	}
	if node != nil {
		err = fmt.Errorf("file %v exists", path)
		return
	}
	p := Path(path)
	up := p.Up().String()
	var dir File
	node, err = fs.Lookup(ctx, &dir, up)
	if err != nil {
		return
	}
	if node == nil {
		dir, err = fs.Create(ctx, up)
		if err != nil {
			return
		}
	}
	f.IRI = fs.Store.GenerateIRI(nil)
	f.Name = p.Base()
	f.Dir = &dir
	_, err = fs.Store.Insert(f)
	return
}

func (fs *Fs) Path(ctx context.Context, node quad.Value) (s string, err error) {
	p := path.StartPath(fs.Store.Graph).FollowRecursive(UpMorphism, -1, nil)
	var nodes []quad.Value
	nodes, err = p.Iterate(ctx).AllValues(nil)
	if err != nil {
		err = fmt.Errorf("error finding nodes up: %v", err)
		return
	}
	nodes = append([]quad.Value{node}, nodes...)
	for _, node = range nodes {
		var file File
		err = fs.Store.Select(ctx, &file, node)
		if err != nil {
			err = fmt.Errorf("error reading node %v: %v", node, err)
			return
		}
		if file.IRI == fs.Root.IRI {
			return
		}
		s = filepath.Join(file.Name, s)
	}
	err = fmt.Errorf("root not found")
	return
}