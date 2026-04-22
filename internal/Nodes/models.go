package Nodes

type AgentState struct {
	Input    string `json:"input"`
	Location string `json:"location"`
	Weather  string `json:"Nodes"`
	Output   string `json:"output"`
}

type Weather struct {
	TempC       string `json:"temp_C"`
	FeelsLikeC  string `json:"feels_like_C"`
	UvIndex     string `json:"uv_index"`
	Humidity    string `json:"humidity"`
	Description string `json:"description"`
}
