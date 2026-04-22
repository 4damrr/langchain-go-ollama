package graph

import (
	"log"

	"github.com/futurxlab/golanggraph/logger"
	"github.com/futurxlab/golanggraph/state"
)

func NewGraphBuilder() {
	_, err := logger.NewLogger()
	if err != nil {
		log.Fatal(err)
	}
}

type Pipeline struct {
	nodes []Node
}

func NewPipeline() *Pipeline {
	return &Pipeline{}
}

type Node interface {
	Run(state *state.State) error
}

func (p *Pipeline) Run(s *state.State) error {
	for _, node := range p.nodes {
		if err := node.Run(s); err != nil {
			return err
		}
	}
	return nil
}
