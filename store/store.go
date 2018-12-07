package store

import (
	"github.com/cayleygraph/cayley/schema"
	"github.com/cayleygraph/cayley/graph"
	"context"
	"github.com/cayleygraph/cayley/quad"
	"github.com/segmentio/ksuid"
)

type Store struct {
	Schema *schema.Config
	Graph  *graph.Handle
}

func New(g *graph.Handle) *Store {
	s := Store{
		Schema: schema.NewConfig(),
		Graph:  g,
	}
	s.Schema.GenerateID = func(_ interface{}) quad.Value {
		return quad.BNode(ksuid.New().String())
	}
	return &s
}

func (s *Store) GenerateIRI(_ interface{}) quad.IRI {
	return quad.IRI(string(s.Schema.GenerateID(nil).Native().(quad.BNode)))
}

func (s *Store) Select(ctx context.Context, dst interface{}, ids ...quad.Value) error {
	return s.Schema.LoadTo(ctx, s.Graph, dst, ids...)
}

func (s *Store) Insert(o interface{}) (quad.Value, error) {
	qw := graph.NewWriter(s.Graph)
	node, err := s.Schema.WriteAsQuads(qw, o)
	if err == nil {
		err = qw.Flush()
	}
	return node, err
}

type Valuer interface {
	Value() quad.Value
}