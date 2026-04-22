package Nodes

import (
	"context"

	flowcontract "github.com/futurxlab/golanggraph/contract"
	"github.com/futurxlab/golanggraph/state"
)

type ExtractLocationNode struct {
	Tool *ToolImpl
}

func (n *ExtractLocationNode) Name() string {
	return "parse_location"
}

func (n *ExtractLocationNode) Run(
	ctx context.Context,
	s *state.State,
	stream flowcontract.StreamFunc,
) error {
	userInput := s.Metadata["userInput"].(string)
	if userInput == "" {
		s.SetNextNodes([]string{"fallback"})
		return nil
	}

	// Set current node context
	s.SetNode(n.Name())

	location, err := n.Tool.ExtractLocation(ctx, userInput)
	if err != nil || location == nil {
		s.SetNextNodes([]string{"fallback"})
		return err
	}

	s.Metadata["location"] = *location

	s.SetNextNodes([]string{"fetch_weather"})
	return nil
}
