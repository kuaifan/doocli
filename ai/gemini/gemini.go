package gemini

import (
	"crypto/tls"
	"fmt"
	"github.com/google/generative-ai-go/genai"
	"net/http"
	"net/url"
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

type APIKeyProxyTransport struct {
	APIKey    string
	Transport http.RoundTripper
	ProxyURL  string
}

func (t *APIKeyProxyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	rt := t.Transport
	if rt == nil {
		rt = http.DefaultTransport
	}

	if t.ProxyURL != "" {
		proxyURL, err := url.Parse(t.ProxyURL)
		if err != nil {
			return nil, err
		}
		if transport, ok := rt.(*http.Transport); ok {
			transport.Proxy = http.ProxyURL(proxyURL)
			transport.TLSClientConfig = &tls.Config{
				InsecureSkipVerify: true,
			}
		} else {
			rt = &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			}
		}
	}

	newReq := *req
	args := newReq.URL.Query()
	args.Set("key", t.APIKey)
	newReq.URL.RawQuery = args.Encode()

	resp, err := rt.RoundTrip(&newReq)
	if err != nil {
		return nil, fmt.Errorf("error during round trip: %v", err)
	}

	return resp, nil
}
