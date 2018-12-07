package cayley

import (
	"github.com/cayleygraph/cayley/graph/path"
	"context"
	"github.com/cayleygraph/cayley/quad"
	"hash/fnv"
	"github.com/cayleygraph/cayley"
	"gonum.org/v1/gonum/graph"
	"github.com/cayleygraph/cayley/graph/iterator"
	"gonum.org/v1/gonum/graph/simple"
)

type Value struct {
	quad.Value
}

func (node Value) ID() int64 {
	s := fnv.New32a()
	s.Write([]byte(node.Value.String()))
	return int64(s.Sum32())
}

type Iterator struct {
	it  cayley.Iterator
	qs  cayley.QuadStore
	ctx context.Context
}

func NewIterator(ctx context.Context, it cayley.Iterator, qs cayley.QuadStore) *Iterator {
	return &Iterator{
		it:  it.Clone(),
		qs:  qs,
		ctx: ctx,
	}
}

func (it *Iterator) Next() bool {
	return it.it.Next(it.ctx)
}

func (it *Iterator) Len() int {
	count := iterator.NewCount(it.it.Clone(), it.qs)
	if !count.Next(it.ctx) {
		return 0
	}
	return int(it.qs.NameOf(count.Result()).(quad.Int))
}

func (it *Iterator) Reset() {
	it.it.Reset()
}

func (it *Iterator) Node() graph.Node {
	value := it.qs.NameOf(it.it.Result())
	return Value{value}
}

type Directed struct {
	qs    cayley.QuadStore
	adj   *path.Path
	ctx   context.Context
	nodes map[int64]graph.Node
}

func NewDirected(ctx context.Context, qs cayley.QuadStore, adj *path.Path) *Directed {
	return &Directed{
		ctx:   ctx,
		qs:    qs,
		adj:   adj,
		nodes: make(map[int64]graph.Node),
	}
}

func (g *Directed) Node(id int64) graph.Node {
	node, ok := g.nodes[id]
	if ok {
		return node
	}
	nodes := g.Nodes()
	for nodes.Next() {
		n := nodes.Node()
		nid := n.ID()
		g.nodes[nid] = n
		if nid == id {
			return n
		}
	}
	return nil
}

func (g *Directed) Nodes() graph.Nodes {
	it := path.StartPath(g.qs).BuildIterator()
	return NewIterator(g.ctx, it, g.qs)
}

func (g *Directed) From(id int64) graph.Nodes {
	u := g.Node(id)
	if u == nil {
		return nil
	}
	it := path.StartPath(g.qs, u.(Value)).Follow(g.adj).BuildIterator()
	return NewIterator(g.ctx, it, g.qs)
}

func (g *Directed) To(id int64) graph.Nodes {
	v := g.Node(id)
	if v == nil {
		return nil
	}
	it := path.StartPath(g.qs, v.(Value)).Follow(g.adj.Reverse()).BuildIterator()
	return NewIterator(g.ctx, it, g.qs)
}

func (g *Directed) HasEdgeFromTo(uid, vid int64) bool {
	u := g.Node(uid)
	if u == nil {
		return false
	}
	v := g.Node(vid)
	if v == nil {
		return false
	}
	nodes, err := path.StartPath(g.qs, u.(Value)).Follow(g.adj).Is(v.(Value)).Iterate(g.ctx).AllValues(nil)
	if err != nil {
		// TODO: Warning?
		return false
	}
	return len(nodes) > 0
}

func (g *Directed) HasEdgeBetween(xid, yid int64) bool {
	if g.HasEdgeFromTo(xid, yid) {
		return true
	}
	if g.HasEdgeFromTo(yid, xid) {
		return true
	}
	return false
}

func (g *Directed) Edge(uid, vid int64) graph.Edge {
	if !g.HasEdgeFromTo(uid, vid) {
		return nil
	}
	u := g.Node(uid)
	v := g.Node(vid)
	return simple.Edge{F: u, T: v}
}