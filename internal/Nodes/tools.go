package Nodes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"langchain-go-ollama/internal/llm"
	"log"
	"net/http"
)

type ToolImpl struct {
	llm *llm.OllamaLLM
}

type Tool interface {
	FetchWeather(ctx context.Context, loc string) (*Weather, error)
	ExtractLocation(ctx context.Context, userInput string) (*string, error)
	OutdoorRecommendation(ctx context.Context, weather string) (*string, error)
}

func NewTool(llm *llm.OllamaLLM) (*ToolImpl, error) {
	if llm == nil {
		return nil, errors.New("llm is nil")
	}
	return &ToolImpl{
		llm: llm,
	}, nil
}

func (n *ToolImpl) FetchWeather(ctx context.Context, loc string) (*Weather, error) {
	url := fmt.Sprintf("https://wttr.in/%s?format=j1", loc)

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	defer response.Body.Close()
	var data map[string]interface{}
	if err := json.NewDecoder(response.Body).Decode(&data); err != nil {
		return nil, err
	}

	current := data["current_condition"].([]interface{})[0].(map[string]interface{})

	return &Weather{
		TempC:       current["temp_C"].(string),
		FeelsLikeC:  current["FeelsLikeC"].(string),
		UvIndex:     current["uvIndex"].(string),
		Humidity:    current["humidity"].(string),
		Description: current["weatherDesc"].([]interface{})[0].(map[string]interface{})["value"].(string),
	}, nil
}

func (n *ToolImpl) ExtractLocation(ctx context.Context, userInput string) (*string, error) {
	prompt := fmt.Sprintf(`
Extract the city name from this sentence.

Sentence: %s

Only return the city name. No explanation.
If there is no city name, return Bandung as default.
`, userInput)

	answer, err := n.llm.Ask(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return &answer, nil
}

func (n *ToolImpl) OutdoorRecommendation(ctx context.Context, weather string) (*string, error) {
	prompt := fmt.Sprintf(`
Weather data:
%s

Give a decision whether it is suitable for running outdoor activities or not based on the weather data.

You must return a JSON object with keys 'input', 'location', 'decision', and 'output'.
Rules: decision should be either 'suitable' or 'not suitable' based on the weather data. If the temperature is above 30°C, or if there is rain, then it is 'not suitable'. Otherwise, it is 'suitable'.
no extra text outside the JSON object.

The input must be the same as the user question, location must be the city name, and output must be a brief explanation of the decision.`, weather)
	answer, err := n.llm.Ask(ctx, prompt)
	if err != nil {
		return nil, err
	}
	return &answer, nil
}
