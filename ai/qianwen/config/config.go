package config

import (
	"io"
	"net/http"
)

const (
	QianwenAPIURLv1 = "https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation"
)

type QianWenClientConfig struct {
	authToken 			 string
	BaseURL              string
	ContentType              string
	XDashScopeSSE        string   // 是否开启SSE
	Data 				 map[string]interface{}
	HTTPClient           *http.Client
}
type ChatResponse struct {
	Output 				OutputResponse	`json:"output"`
	Usage 				UsageResponse	`json:"usage"`
	RequestId			string  		`json:"request_id"`
}

type OutputResponse struct {
	FinishReason    string  `json:"finish_reason"`
	Text 			string  `json:"text"`
}

type UsageResponse struct {
	InputTokens 	int 	`json:"input_tokens"`
	OutputTokens    int 	`json:"output_tokens"`
}

type ChatRequest struct {
	Model 			string					`json:"model"`
	Input 			InputResquest			`json:"input"`
	parameters 		ParametersResquest 		`json:"parameters"`
}

type InputResquest struct {
	Message 	string				`json:"prompt"`
	History 	[]*HistoryResquest	`json:"history"`
}

type ParametersResquest struct {
	TopP 		 float64 	`json:"top_p"`
	TopK 		 int		`json:"top_k"`
	Seed 		 int 		`json:"seed"`
	EnableSearch bool 		`json:"enable_search"`
}
type HistoryResquest struct {
	User 	string `json:"user"`
	Bot 	string `json:"bot"`
}

func (res ChatResponse) Recv()  error {
	if res.Output.FinishReason=="stop" {
		return io.EOF
	}
	return nil
}

func DefaultConfig(authToken string) QianWenClientConfig {
	return QianWenClientConfig{
		authToken: 		authToken,
		BaseURL:   		QianwenAPIURLv1,
		ContentType:    "application/json",
		XDashScopeSSE:  "enable",
		Data:			make(map[string]interface{}),
		HTTPClient: 	&http.Client{},
	}
}

