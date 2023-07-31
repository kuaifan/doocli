package ai

import (
	"github.com/alexandrevicenzi/go-sse"
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

var (
	HttpPort  string
	ServerUrl string

	ClaudeToken  string
	ClaudeAgency string

	OpenaiKey    string
	OpenaiAgency string

	sources       *sse.Server
	clients       []*clientModel
	openaiContext []*openaiModel
)
