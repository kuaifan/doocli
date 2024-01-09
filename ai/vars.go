package ai

import (
	"doocli/ai/qianwen/config"

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

var (
	HttpPort  string
	ServerUrl string

	ClaudeToken  string
	ClaudeAgency string

	OpenaiKey       string
	OpenaiAgency    string
	OpenaiWordCount int

	WenxinKey    string
	WenxinSecret string
	WenxinModel  string

	QianwenKey   string
	QianwenModel string

	sources        *sse.Server
	clients        []*clientModel
	openaiContext  []*openaiModel
	wenxinContext  []*wenxinModel
	qianwenContext []*qianwenModel

	clears = []string{":clear", ":reset", ":restart", ":new", ":清空上下文", ":重置上下文", ":重启", ":重启对话"}
)
