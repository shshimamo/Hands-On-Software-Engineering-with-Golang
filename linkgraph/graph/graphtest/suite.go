package graphtest

import (
	"github.com/google/uuid"
	"github.com/shshimamo/Hands-On-Software-Engineering-with-Golang/linkgraph/graph"
	gc "gopkg.in/check.v1"
	"time"
)

type SuiteBase struct {
	g graph.Graph
}

func (s *SuiteBase) SetGraph(g graph.Graph) {
	s.g = g
}

func (s *SuiteBase) TestUpsertLink(c *gc.C) {
	original := &graph.Link{
		URL:         "https://example.com",
		RetrievedAt: time.Now().Add(-10 * time.Hour),
	}

	err := s.g.UpsertLink(original)
	c.Assert(err, gc.IsNil)
	c.Assert(original.ID, gc.Not(gc.Equals), uuid.Nil, gc.Commentf("expected a linkID to be assigned to the new link"))
}
