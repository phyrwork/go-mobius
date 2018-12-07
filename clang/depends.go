package clang

import (
	"github.com/phyrwork/mobius/fs"
	"context"
	"github.com/cayleygraph/cayley/quad"
	"fmt"
	"github.com/cayleygraph/cayley/graph/path"
)

func IncludePath(s *fs.Fs, dirs ...string) (p *path.Path, errs []error) {
	nodes := make([]quad.Value, 0)
	for _, p := range dirs {
		node, err := s.Lookup(nil, nil, p)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if node == nil {
			errs = append(errs, fmt.Errorf("include directory %v not found", p))
			continue
		}
		nodes = append(nodes, node)
	}
	if len(nodes) == 0 {
		// No path to resolve against
		// Return empty here, otherwise starting a p with no nodes will return all nodes
		return
	}
	p = path.StartPath(s.Store.Graph, nodes...)
	return
}

func ResolveInclude(ctx context.Context, p *path.Path, includes ...Include) (depends map[quad.Value]struct{}, errs []error) {
	if p == nil {
		// No path to resolve against
		errs = []error{fmt.Errorf("nil include path")}
		return
	}
	depends = make(map[quad.Value]struct{})
	for _, i := range includes {
		ip, err := i.Path()
		if err != nil {
			errs = append(errs, fmt.Errorf("error in include %v: %v", i, err))
			continue
		}
		nodes, err := p.Follow(fs.Path(ip).StartMorphism()).Iterate(ctx).AllValues(nil)
		if err != nil {
			errs = append(errs, fmt.Errorf("error following include path: %v", err))
			continue
		}
		switch len(nodes) {
		case 0:
			errs = append(errs, fmt.Errorf("%v not resolved", i))
		case 1:
			depends[nodes[0]] = struct{}{}
		default:
			ids := make([]string, len(nodes))
			for n, node := range nodes {
				ids[n] = node.String()
			}
			errs = append(errs, fmt.Errorf("%v resolved ambiguously: %v", i, ids))
			continue
		}
	}
	return
}