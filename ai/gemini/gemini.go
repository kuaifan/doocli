package gemini

import (
	"github.com/google/generative-ai-go/genai"
	"net/http"
)

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

type CustomTransport struct {
	Transport http.RoundTripper
	Key       string
}

func (c *CustomTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	query := req.URL.Query()
	query.Add("key", c.Key)
	req.URL.RawQuery = query.Encode()
	return c.Transport.RoundTrip(req)
}
