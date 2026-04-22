package Nodes

import (
	"context"
	"encoding/json"

	flowcontract "github.com/futurxlab/golanggraph/contract"
	"github.com/futurxlab/golanggraph/state"
)

type SummaryNode struct {
	Tool *ToolImpl
}

func (n *SummaryNode) Name() string {
	return "summary"
}

func (n *SummaryNode) Run(
	ctx context.Context,
	s *state.State,
	stream flowcontract.StreamFunc,
) error {
	weather := s.Metadata["weather"].(Weather)

	// Set current node context
	s.SetNode(n.Name())

	bytes, _ := json.Marshal(weather)
	str := string(bytes)

	summary, err := n.Tool.OutdoorRecommendation(ctx, str)
	if err != nil || summary == nil {
		return err
	}

	s.Metadata["summary"] = *summary

	//s.SetNextNodes([]string{"format"})
	return nil
}
