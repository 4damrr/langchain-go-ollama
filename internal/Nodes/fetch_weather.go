package Nodes

import (
	"context"

	flowcontract "github.com/futurxlab/golanggraph/contract"
	"github.com/futurxlab/golanggraph/state"
)

type FetchWeatherNode struct {
	Tool *ToolImpl
}

func (n *FetchWeatherNode) Name() string {
	return "fetch_weather"
}

func (n *FetchWeatherNode) Run(
	ctx context.Context,
	s *state.State,
	stream flowcontract.StreamFunc,
) error {
	loc := s.Metadata["location"]
	if loc == "" {
		s.SetNextNodes([]string{"fallback"})
		return nil
	}

	// Set current node context
	s.SetNode(n.Name())

	w, err := n.Tool.FetchWeather(ctx, loc.(string))
	if err != nil || w == nil {
		return err
	}

	s.Metadata["weather"] = *w

	//s.SetNextNodes([]string{"format"})
	return nil
}
