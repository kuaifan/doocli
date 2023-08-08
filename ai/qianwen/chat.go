package qianwen

import (
	"bufio"
	"bytes"
	"context"
	qianwenconfig "doocli/ai/qianwen/config"
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strconv"
)
type Client struct {
	client      *http.Client
	req 		*http.Request
	Sender   	chan qianwenconfig.BaseResponse
}

func (cli *Client) ChatStream() error {
	resp, err := cli.client.Do(cli.req)
	if err != nil {
		return err
	}
	statusParts := resp.Status
	statusCode, err := strconv.Atoi(statusParts[:3])
	if err != nil {
		return err
	}
	if  statusCode != http.StatusOK{
		defer resp.Body.Close()
		var baseResult  qianwenconfig.BaseResponse
		var errResp qianwenconfig.ErrorResponse
		baseResult.Status = http.StatusOK

		data,err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		err = json.Unmarshal(data, &errResp)
		if err != nil {
			return err
		}
		baseResult.FinishReason = "stop"
		baseResult.Message = errResp.Message
		cli.Sender <- baseResult
		return nil
	}

	go func() {
		defer resp.Body.Close()
		var status int64
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()

			statusRegex := regexp.MustCompile(`:HTTP_STATUS/(.*?)(\r?\n|$)`)
			statusLine := statusRegex.FindStringSubmatch(line)
			if len(statusLine) > 2 && status == 0 {
				status,_ = strconv.ParseInt(statusLine[1],0,64)
			}

			dataRegex := regexp.MustCompile(`data:(.*?)(\r?\n|$)`)
			match := dataRegex.FindStringSubmatch(line)
			if len(match) < 2 {
				continue
			}
			jsonData := match[1]

			var baseResult  qianwenconfig.BaseResponse
			var result qianwenconfig.ChatResponse
			var errResp qianwenconfig.ErrorResponse
			baseResult.Status = status
			if status != http.StatusOK {
				err := json.Unmarshal([]byte(jsonData), &errResp)
				if err != nil {
					return
				}
				baseResult.FinishReason = "stop"
				baseResult.Message = errResp.Message
				cli.Sender <- baseResult
			}else{
				err := json.Unmarshal([]byte(jsonData), &result)
				if err != nil {
					continue
				}
				baseResult.FinishReason = result.Output.FinishReason
				baseResult.Message = result.Output.Text
				cli.Sender <- baseResult
				if baseResult.Recv() != nil {
					close(cli.Sender)
				}
			}
		}
	}()
	return nil

}

func New(ctx context.Context,apiKey string,data map[string]interface{}) (*Client,error) {
	cli := &Client{
		client: &http.Client{},
		Sender: make(chan qianwenconfig.BaseResponse,9999),
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
