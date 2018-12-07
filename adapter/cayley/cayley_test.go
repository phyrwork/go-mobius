package cayley

import (
	"testing"
	"github.com/cayleygraph/cayley"
	"github.com/cayleygraph/cayley/quad"
	"context"
	"github.com/cayleygraph/cayley/graph/path"
	"gonum.org/v1/gonum/graph"
	"log"
	"gonum.org/v1/gonum/graph/topo"
)

func newValue(k interface{}) (v Value) {
	switch t := k.(type) {
	case string:
		v = Value{quad.String(t)}
	case quad.Value:
		v = Value{t}
	default:
		log.Panicf("not supported for test value: %v", k)
	}
	return v
}

func listEqual(a, b []graph.Node) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i].ID() != b[i].ID() {
			return false
		}
	}
	return true
}

func cycleEqual(a, b []graph.Node) bool {
	if len(a) != len(b) {
		return false
	}
	if len(a) < 3 {
		log.Panicf("not a cycle (too short): %v", a)
	}
	if len(b) < 3 {
		log.Panicf("not a cycle (too short): %v", b)
	}
	a = a[1:]
	b = b[1:]
	for i := 0; true; {
		if listEqual(a, b) {
			return true
		}
		i++
		if i >= len(a) {
			break
		}
		b = append(b[1:], b[0])
	}
	return false
}

func cycleSetEqual(a, b [][]graph.Node) bool {
	if len(a) != len(b) {
		return false
	}
	rem := func (a [][]graph.Node) map[int][]graph.Node {
		m := make(map[int][]graph.Node)
		for i, c := range a {
			m[i] = c
		}
		return m
	}
	ra, rb := rem(a), rem(b)
	for len(ra) != 0 && len(rb) != 0 {
		found := false
		search:
			for ia, ca := range ra {
				for ib, cb := range rb {
					if cycleEqual(ca, cb) {
						found = true
						delete(ra, ia)
						delete(rb, ib)
						break search
					}
				}
			}
		if !found {
			return false
		}
	}
	return true
}

func TestIterator(t *testing.T) {
	ctx := context.TODO()
	qs, err := cayley.NewMemoryGraph()
	if err != nil {
		t.Skipf("error creating graph: %v", err)
	}
	for _, e := range []struct {
		from string
		to   string
	}{
		{"1", "2"},
		{"2", "3"},
	} {
		q := quad.Make(quad.String(e.from), nil, quad.String(e.to), nil)
		if err := qs.AddQuad(q); err != nil {
			t.Skipf("error creating edge: %v", err)
		}
	}
	a := make(map[int64]graph.Node)
	it := NewIterator(ctx, path.StartPath(qs).BuildIterator(), qs)
	for it.Next() {
		n := it.Node()
		a[n.ID()] = n
	}
	if len(a) != 3 {
		t.Fatalf("unexpected node count: expected %v, got %v", 3, len(a))
	}
}

func TestDirected_Node(t *testing.T) {
	ctx := context.TODO()
	qs, err := cayley.NewMemoryGraph()
	if err != nil {
		t.Skipf("error creating graph: %v", err)
	}
	q := quad.Make(quad.String("1"), nil, quad.String("2"), nil)
	if err := qs.AddQuad(q); err != nil {
		t.Skipf("error creating edge: %v", err)
	}
	e := Value{quad.String("1")}
	eid := e.ID()
	g := NewDirected(ctx, qs, path.StartMorphism().Out())
	a := g.Node(eid)
	aid := a.ID()
	if aid != eid {
		t.Fatalf("unexpected node: expected %v, got %v", eid, aid)
	}
}

func TestDirected_From(t *testing.T) {
	ctx := context.TODO()
	qs, err := cayley.NewMemoryGraph()
	if err != nil {
		t.Skipf("error creating graph: %v", err)
	}
	q := quad.Make(quad.String("1"), nil, quad.String("2"), nil)
	if err := qs.AddQuad(q); err != nil {
		t.Skipf("error creating edge: %v", err)
	}
	g := NewDirected(ctx, qs, path.StartMorphism().Out())
	u := Value{quad.String("1")}
	uid := u.ID()
	v := Value{quad.String("2")}
	vid := v.ID()
	a := g.From(uid)
	c := a.Len()
	if c != 1 {
		t.Fatalf("unexpected node count: expected %v, got %v", 1, c)
	}
	a.Next()
	aid := a.Node().ID()
	if aid != vid {
		t.Fatalf("unexpected node: expected %v, got %v", vid, aid)
	}
}

func TestDirected_To(t *testing.T) {
	ctx := context.TODO()
	qs, err := cayley.NewMemoryGraph()
	if err != nil {
		t.Skipf("error creating graph: %v", err)
	}
	q := quad.Make(quad.String("1"), nil, quad.String("2"), nil)
	if err := qs.AddQuad(q); err != nil {
		t.Skipf("error creating edge: %v", err)
	}
	g := NewDirected(ctx, qs, path.StartMorphism().Out())
	u := Value{quad.String("1")}
	uid := u.ID()
	v := Value{quad.String("2")}
	vid := v.ID()
	a := g.To(vid)
	c := a.Len()
	if c != 1 {
		t.Fatalf("unexpected node count: expected %v, got %v", 1, c)
	}
	a.Next()
	aid := a.Node().ID()
	if aid != uid {
		t.Fatalf("unexpected node: expected %v, got %v", uid, aid)
	}
}

func TestDirected_HasEdgeFromTo(t *testing.T) {
	tests := []struct {
		name string
		edges []struct{
			from string
			to   string
		}
		from string
		to   string
		has  bool
	}{
		{
			"true",
			[]struct{
				from string
				to   string
			}{
				{"1", "2"},
			},
			"1",
			"2",
			true,
		},
		{
			"false",
			[]struct{
				from string
				to   string
			}{
				{"2", "1"},
			},
			"1",
			"2",
			false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.TODO()
			qs, err := cayley.NewMemoryGraph()
			if err != nil {
				t.Skipf("error creating graph: %v", err)
			}
			for _, e := range test.edges {
				q := quad.Make(newValue(e.from).Value, nil, newValue(e.to).Value, nil)
				if err := qs.AddQuad(q); err != nil {
					t.Skipf("error creating edge: %v", err)
				}
			}
			g := NewDirected(ctx, qs, path.StartMorphism().Out())
			h := g.HasEdgeFromTo(newValue(test.from).ID(), newValue(test.to).ID())
			if h != test.has {
				t.Fatalf("unexpected result: expected %v, got %v", test.has, h)
			}
		})
	}
}

func TestDirected_HasEdgeBetween(t *testing.T) {
	tests := []struct {
		name string
		edges []struct{
			from string
			to   string
		}
		x   string
		y   string
		has bool
	}{
		{
			"out",
			[]struct{
				from string
				to   string
			}{
				{"1", "2"},
			},
			"1",
			"2",
			true,
		},
		{
			"in",
			[]struct{
				from string
				to   string
			}{
				{"2", "1"},
			},
			"1",
			"2",
			true,
		},
		{
			"false",
			[]struct{
				from string
				to   string
			}{
				{"1", "3"},
			},
			"1",
			"2",
			false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.TODO()
			qs, err := cayley.NewMemoryGraph()
			if err != nil {
				t.Skipf("error creating graph: %v", err)
			}
			for _, e := range test.edges {
				q := quad.Make(newValue(e.from).Value, nil, newValue(e.to).Value, nil)
				if err := qs.AddQuad(q); err != nil {
					t.Skipf("error creating edge: %v", err)
				}
			}
			g := NewDirected(ctx, qs, path.StartMorphism().Out())
			h := g.HasEdgeBetween(newValue(test.x).ID(), newValue(test.y).ID())
			if h != test.has {
				t.Fatalf("unexpected result: expected %v, got %v", test.has, h)
			}
		})
	}
}

// Gonum integration test
func TestDirectedCyclesIn(t *testing.T) {
	tests := []struct {
		name string
		edges []struct {
			from string
			to string
		}
		cycles [][]string
	}{
		{
			"simple cycle",
			[]struct {
				from string
				to string
			}{
				{"8", "9"},
				{"9", "8"},
			},
			[][]string{
				{"8", "9", "8"},
			},
		},
		{
			"strongly connected component (isolated)",
			[]struct {
				from string
				to string
			}{
				{"1", "2"},
				{"1", "5"},
				{"2", "3"},
				{"3", "1"},
				{"3", "2"},
				{"3", "4"},
				{"3", "6"},
				{"4", "5"},
				{"5", "2"},
				{"6", "4"},
			},
			[][]string{
				{"1", "2", "3", "1"},
				{"1", "5", "2", "3", "1"},
				{"2", "3", "2"},
				{"2", "3", "4", "5", "2"},
				{"2", "3", "6", "4", "5", "2"},
			},
		},
		{
			"strongly connected components (connected)",
			[]struct {
				from string
				to string
			}{
				{"1", "2"},
				{"1", "5"},
				{"1", "8"},
				{"2", "3"},
				{"2", "7"},
				{"2", "9"},
				{"3", "1"},
				{"3", "2"},
				{"3", "4"},
				{"3", "6"},
				{"4", "5"},
				{"5", "2"},
				{"6", "4"},
				{"8", "9"},
				{"9", "8"},
			},
			[][]string{
				{"8", "9", "8"},
				{"1", "2", "3", "1"},
				{"1", "5", "2", "3", "1"},
				{"2", "3", "2"},
				{"2", "3", "4", "5", "2"},
				{"2", "3", "6", "4", "5", "2"},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Setup
			ctx := context.TODO()
			qs, err := cayley.NewMemoryGraph()
			if err != nil {
				t.Skipf("error creating graph: %v", err)
			}
			for _, e := range test.edges {
				q := quad.Make(newValue(e.from).Value, nil, newValue(e.to).Value, nil)
				if err := qs.AddQuad(q); err != nil {
					t.Skipf("error creating edge: %v", err)
				}
			}
			e := make([][]graph.Node, len(test.cycles))
			for i, c := range test.cycles {
				e[i] = make([]graph.Node, len(c))
				for j, v := range c {
					e[i][j] = newValue(v)
				}
			}
			// Test
			g := NewDirected(ctx, qs, path.StartMorphism().Out())
			a := topo.DirectedCyclesIn(g)
			if !cycleSetEqual(e, a) {
				t.Fatalf("unexpected cycles: expected %v, got %v", e, a)
			}
		})
	}
}

