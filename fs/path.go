package fs

import (
	"path/filepath"
	"github.com/cayleygraph/cayley/quad"
	"github.com/cayleygraph/cayley/graph/path"
	"strings"
	"github.com/cayleygraph/cayley/graph"
)

type Path string

func (s Path) String() string {
	return string(s)
}

func (s Path) Up() Path {
	return Path(filepath.Dir(string(s)))
}

func (s Path) Down(name string) Path {
	return Path(filepath.Join(string(s), name))
}

func (s Path) Base() string {
	return filepath.Base(string(s))
}

func (s Path) StartPath(g graph.QuadStore, nodes ...quad.Value) *path.Path {
	p := path.StartPath(g, nodes...)
	// TODO: Consider filepath.Split instead
	for _, base := range strings.Split(filepath.ToSlash(string(s)), "/") {
		switch base {
		case "":
			// No-op
		case ".":
			// No-op
		case "..":
			p = p.Follow(UpMorphism)
		default:
			p = p.Follow(DownMorphism(base))
		}
	}
	return p
}

func (s Path) StartMorphism(nodes ...quad.Value) *path.Path {
	return s.StartPath(nil, nodes...)
}