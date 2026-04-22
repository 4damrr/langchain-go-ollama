package main

import (
	"context"
	"fmt"
	"langchain-go-ollama/internal/Nodes"
	"langchain-go-ollama/internal/llm"
	"os"

	"github.com/futurxlab/golanggraph/checkpointer"
	"github.com/futurxlab/golanggraph/edge"
	"github.com/futurxlab/golanggraph/flow"
	"github.com/futurxlab/golanggraph/logger"
	"github.com/futurxlab/golanggraph/state"
)

func main() {
	//ctx := context.Background()

	ollamaClient, err := llm.NewOllamaLLM()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	tools, err := Nodes.NewTool(ollamaClient)
	if err != nil {
		return
	}

	fetchWeather := Nodes.FetchWeatherNode{
		Tool: tools,
	}
	parseLocation := Nodes.ExtractLocationNode{Tool: tools}
	summary := Nodes.SummaryNode{
		Tool: tools,
	}

	langGraphLogger, err := logger.NewLogger()
	if err != nil {
		return
	}
	flow, err := flow.NewFlowBuilder(langGraphLogger).
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
		panic(err)
	}
	userInput := make(map[string]interface{})
	userInput["userInput"] = "What is the weather like in here?"
	initialState := state.State{
		Metadata: userInput,
	}
	outputState, err := flow.Exec(context.Background(), initialState, nil)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(outputState.Metadata["summary"])
}
