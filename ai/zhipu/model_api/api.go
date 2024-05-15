package model_api

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"net/textproto"
	"regexp"
	"strings"
	"time"

	"doocli/ai/zhipu/utils"
)

var v4url string = "https://open.bigmodel.cn/api/paas/v4/"

var v3url string = "https://open.bigmodel.cn/api/paas/v4/"

type PostParams struct {
	Model    string      `json:"model"`
	Messages []*Messages `json:"messages"`
	Stream   bool        `json:"stream"`
}

type Messages struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type GLM4VPostParams struct {
	Model    string          `json:"model"`
	Messages []*GLM4VMessage `json:"messages"`
	Stream   bool            `json:"stream"`
}
type GLM4VMessage struct {
	Role    string          `json:"role"`
	Content []*GLM4VContent `json:"content"`
}

type GLM4VContent struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	ImageUrl *struct {
		Url string `json:"url"`
	} `json:"image_url,omitempty"`
}

type StreamResponseData struct {
	ID      string `json:"id"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int    `json:"index"`
		FinishReason string `json:"finish_reason,omitempty"`
		Delta        struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage,omitempty"`
}

func ParseResponse(rawResponse string) (*StreamResponseData, error) {
	// 找到有效数据部分的起始位置
	startIndex := strings.Index(rawResponse, "{")
	if startIndex == -1 {
		return nil, nil
	}

	// 截取有效数据部分
	validData := rawResponse[startIndex:]

	// 手动解析
	var data *StreamResponseData
	if err := json.Unmarshal([]byte(validData), &data); err != nil {
		return nil, fmt.Errorf("%v", err)
	}

	return data, nil
}

// 通用模型
func BeCommonModel(expireAtTime int64, postParams PostParams, apiKey string) (map[string]interface{}, error) {

	token, _ := utils.GenerateToken(apiKey, expireAtTime)

	// 示例用法
	apiURL := v4url + "chat/completions"
	timeout := 60 * time.Second

	postResponse, err := utils.Post(apiURL, token, postParams, timeout)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}
	return postResponse, nil
}

func BeCommonModelStream(expireAtTime int64, postParams *PostParams, glm4vparam *GLM4VPostParams, apiKey string) (*bufio.Scanner, error) {
	token, _ := utils.GenerateToken(apiKey, expireAtTime)

	// 示例用法
	apiURL := v4url + "chat/completions"
	timeout := 60 * time.Second

	var postResponse *http.Response
	if glm4vparam != nil {
		postResponse2, err := utils.Stream(apiURL, token, glm4vparam, timeout)
		if err != nil {
			return nil, err
		}
		postResponse = postResponse2
	} else {
		postResponse3, err := utils.Stream(apiURL, token, postParams, timeout)
		if err != nil {
			return nil, err
		}
		postResponse = postResponse3
	}

	scanner := bufio.NewScanner(postResponse.Body)
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return scanner, nil
}

type PostImageParams struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

// 图像大模型
func ImageLargeModel(expireAtTime int64, prompt string, apiKey string, model string) (map[string]interface{}, error) {

	token, _ := utils.GenerateToken(apiKey, expireAtTime)

	// 示例用法
	apiURL := v4url + "images/generations"
	timeout := 60 * time.Second

	// 示例 POST 请求
	postParams := PostImageParams{
		Model:  model,
		Prompt: prompt,
	}

	postResponse, err := utils.Post(apiURL, token, postParams, timeout)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}
	return postResponse, nil
}

type PostSuperhumanoidParams struct {
	Prompt []Prompt `json:"prompt"`
	Meta   []Meta   `json:"meta"`
}
type Prompt struct {
	Role    string `json:"prompt"`
	Content string `json:"content"`
}
type Meta struct {
	UserInfo string `json:"user_info"`
	BotInfo  string `json:"bot_info"`
	BotName  string `json:"bot_name"`
	UserName string `json:"user_name"`
}

// 超拟人大模型
func SuperhumanoidModel(expireAtTime int64, meta []Meta, prompt []Prompt, apiKey string) (map[string]interface{}, error) {

	token, _ := utils.GenerateToken(apiKey, expireAtTime)

	// 示例用法
	apiURL := v3url + "model-api/charglm-3/sse-invoke"
	timeout := 60 * time.Second

	// 示例 POST 请求
	postParams := PostSuperhumanoidParams{
		Prompt: prompt,
		Meta:   meta,
	}

	postResponse, err := utils.Post(apiURL, token, postParams, timeout)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}
	return postResponse, nil
}

type PostVectorParams struct {
	Input string `json:"input"`
	Model string `json:"model"`
}

// 向量模型
func VectorModel(expireAtTime int64, input string, apiKey string, model string) (map[string]interface{}, error) {

	token, _ := utils.GenerateToken(apiKey, expireAtTime)

	// 示例用法
	apiURL := v4url + "mbeddings"
	timeout := 60 * time.Second

	// 示例 POST 请求
	postParams := PostVectorParams{
		Input: input,
		Model: model,
	}

	postResponse, err := utils.Post(apiURL, token, postParams, timeout)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}
	return postResponse, nil
}

type PostFineTuningParams struct {
	Model        string `json:"model"`
	TrainingFile string `json:"training_file"`
}

// 模型微调
func ModelFineTuning(expireAtTime int64, trainingFile string, apiKey string, model string) (map[string]interface{}, error) {

	token, _ := utils.GenerateToken(apiKey, expireAtTime)

	// 示例用法
	apiURL := v4url + "fine_tuning/jobs"
	timeout := 60 * time.Second

	// 示例 POST 请求
	postParams := PostFineTuningParams{
		Model:        model,
		TrainingFile: trainingFile,
	}

	postResponse, err := utils.Post(apiURL, token, postParams, timeout)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}
	return postResponse, nil
}

type PostFileParams struct {
	File    *FileHeader `json:"file"`
	Purpose string      `json:"purpose"`
}

type FileHeader struct {
	Filename string
	Header   textproto.MIMEHeader
	Size     int64

	content   []byte
	tmpfile   string
	tmpoff    int64
	tmpshared bool
}

// 文件管理
func FileManagement(expireAtTime int64, purpose string, apiKey string, model string, file *FileHeader) (map[string]interface{}, error) {

	token, _ := utils.GenerateToken(apiKey, expireAtTime)

	// 示例用法
	apiURL := v4url + "files"
	timeout := 60 * time.Second

	// 示例 POST 请求
	postParams := PostFileParams{
		File:    file,
		Purpose: purpose,
	}

	postResponse, err := utils.Post(apiURL, token, postParams, timeout)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}
	return postResponse, nil
}

func ParseGLM4VMessage(text string) ([]*GLM4VContent, error) {
	var messages []*GLM4VContent
	//oldtext := text
	// 定义URL的正则表达式
	urlRegex := `https?://[^\s]+`
	re := regexp.MustCompile(urlRegex)

	// 查找所有匹配的URL
	urls := re.FindAllString(text, -1)

	if len(urls) == 0 {
		return nil, nil
	}
	// 提取URL并从文本中删除
	for _, url := range urls {
		imageUrl := &struct {
			Url string `json:"url"`
		}{Url: url}
		fmt.Println("插入url消息")
		messages = append(messages, &GLM4VContent{
			Type:     "image_url",
			ImageUrl: imageUrl,
		})
		text = re.ReplaceAllString(text, "")
	}

	// 修剪文本以删除多余的空格
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	text = regexp.MustCompile(`(^\s+|\s+$)`).ReplaceAllString(text, "")

	// 如果还有剩余的文本，作为文本内容添加
	if text != "" {
		messages = append(messages, &GLM4VContent{
			Type: "text",
			Text: text,
		})
	}
	//
	//resp, err := json.Marshal(messages)
	//if err != nil {
	//	return nil, err
	//}

	return messages, nil

}
