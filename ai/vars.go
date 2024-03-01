package ai

import (
	"doocli/ai/qianwen/config"
	"github.com/google/generative-ai-go/genai"

	"github.com/alexandrevicenzi/go-sse"
	aicustomv1 "github.com/hitosea/go-wenxin/gen/go/baidubce/ai_custom/v1"
	"github.com/sashabaranov/go-openai"
)

type sendModel struct {
	id         string
	text       string
	token      string
	dialogId   string
	dialogType string
	msgId      string
	msgUid     string
	mention    string
	botUid     string
	version    string
	extras     string
}

type clientModel struct {
	id      string
	append  string
	message string
}

type openaiModel struct {
	key      string
	messages []openai.ChatCompletionMessage
}

type wenxinModel struct {
	user     string
	messages []*aicustomv1.Message
}

type qianwenModel struct {
	user     string
	messages []*config.HistoryResquest
}

type geminiModel struct {
	user     string
	messages []*genai.Content
}

var (
	HttpPort  string
	ServerUrl string
	ChunkSize int

	ClaudeToken  string
	ClaudeAgency string

	OpenaiKey    string
	OpenaiAgency string

	WenxinKey    string
	WenxinSecret string
	WenxinModel  string

	QianwenKey   string
	QianwenModel string

	GeminiKey     string
	GeminiModel   string
	GeminiAgency  string
	GeminiTimeout int64 = 20

	sources        *sse.Server
	clients        []*clientModel
	openaiContext  []*openaiModel
	wenxinContext  []*wenxinModel
	qianwenContext []*qianwenModel
	geminiContext  []*geminiModel

	clears = []string{":clear", ":reset", ":restart", ":new", ":清空上下文", ":重置上下文", ":重启", ":重启对话"}
)
