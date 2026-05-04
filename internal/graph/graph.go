package workflow

import (
	"context"
	"langchain-go-ollama/internal/llm"
	"langchain-go-ollama/internal/nodes"

	"github.com/futurxlab/golanggraph/checkpointer"
	"github.com/futurxlab/golanggraph/edge"
	"github.com/futurxlab/golanggraph/flow"
	"github.com/futurxlab/golanggraph/logger"
	"github.com/futurxlab/golanggraph/state"
)

type Service struct {
	flow *flow.Flow
}

func NewWorkflowService(ollamaClient *llm.OllamaLLM) (*Service, error) {
	tools, err := nodes.NewTool(ollamaClient)
	if err != nil {
		return nil, err
	}

	parseLocation := nodes.ExtractLocationNode{Tool: tools}
	fetchWeather := nodes.FetchWeatherNode{Tool: tools}
	summary := nodes.SummaryNode{Tool: tools}

	langGraphLogger, err := logger.NewLogger()
	if err != nil {
		return nil, err
	}

	f, err := flow.NewFlowBuilder(langGraphLogger).
		SetCheckpointer(checkpointer.NewInMemoryCheckpointer()).
		SetName("my_workflow").
		AddNode(&parseLocation).
		AddNode(&fetchWeather).
		AddNode(&summary).
		AddEdge(edge.Edge{From: flow.StartNode, To: parseLocation.Name()}).
		AddEdge(edge.Edge{From: parseLocation.Name(), To: fetchWeather.Name()}).
		AddEdge(edge.Edge{From: fetchWeather.Name(), To: summary.Name()}).
		AddEdge(edge.Edge{From: summary.Name(), To: flow.EndNode}).
		Compile()

	if err != nil {
		return nil, err
	}

	return &Service{flow: f}, nil
}

func (w *Service) Run(ctx context.Context, input string) (string, error) {
	initialState := state.State{
		Metadata: map[string]interface{}{
			"userInput": input,
		},
	}

	outputState, err := w.flow.Exec(ctx, initialState, nil)
	if err != nil {
		return "", err
	}

	return outputState.Metadata["summary"].(string), nil
}

type UserInput struct {
	Query string `json:"query"`
}
