package gemini

import "github.com/google/generative-ai-go/genai"

type GeminiClient struct {
	Client *genai.Client
	Model  *genai.GenerativeModel
}

func NewGemniClient(cli *genai.Client, model string) *GeminiClient {
	return &GeminiClient{
		Client: cli,
		Model:  cli.GenerativeModel(model),
	}
}
