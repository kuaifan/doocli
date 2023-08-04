package qianwen

import (
	"bufio"
	"bytes"
	"context"
	qianwenconfig "doocli/ai/qianwen/config"
	"encoding/json"
	"net/http"
	"regexp"
)
type Client struct {
	client      *http.Client
	req 		*http.Request
	Sender   	chan qianwenconfig.ChatResponse
}

func (cli *Client) ChatStream() error {
	resp, err := cli.client.Do(cli.req)
	if err != nil {
		return err
	}

	go func() {
		defer resp.Body.Close()
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			dataRegex := regexp.MustCompile(`data:(.*?)(\r?\n|$)`)
			match := dataRegex.FindStringSubmatch(line)
			if len(match) < 2 {
				continue
			}
			jsonData := match[1]
			var result qianwenconfig.ChatResponse
			err := json.Unmarshal([]byte(jsonData), &result)
			if err != nil {
				continue
			}
			cli.Sender <- result
		}
	}()
	return nil

}

func New(ctx context.Context,apiKey string,data map[string]interface{}) (*Client,error) {
	cli := &Client{
		client: &http.Client{},
		Sender: make(chan qianwenconfig.ChatResponse,9999),
	}
	config := qianwenconfig.DefaultConfig(apiKey)
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil,err
	}
	cli.req, err = http.NewRequestWithContext(ctx, "POST", config.BaseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil,err
	}
	cli.req.Header.Add("Authorization", "Bearer "+apiKey)
	cli.req.Header.Add("Content-Type", config.ContentType)
	cli.req.Header.Add("X-DashScope-SSE", config.XDashScopeSSE)

	return cli,nil
}