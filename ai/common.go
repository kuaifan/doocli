package ai

import (
	"context"
	"doocli/ai/claude/types"
	"doocli/ai/gemini"
	"doocli/ai/qianwen"
	qianwenconfig "doocli/ai/qianwen/config"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/alexandrevicenzi/go-sse"
	"github.com/google/generative-ai-go/genai"
	aicustomv1 "github.com/hitosea/go-wenxin/gen/go/baidubce/ai_custom/v1"
	"github.com/nahid/gohttp"
	"github.com/sashabaranov/go-openai"
	"github.com/tidwall/gjson"
	"google.golang.org/api/iterator"
	"io"
	"net/http"
	"time"
	"unicode/utf8"
)

func callSend(w http.ResponseWriter, req *http.Request) *sendModel {
	send := &sendModel{
		id:         "",
		text:       req.PostFormValue("text"),
		token:      req.PostFormValue("token"),
		dialogId:   req.PostFormValue("dialog_id"),
		dialogType: req.PostFormValue("dialog_type"),
		msgId:      req.PostFormValue("msg_id"),
		msgUid:     req.PostFormValue("msg_uid"),
		mention:    req.PostFormValue("mention"),
		botUid:     req.PostFormValue("bot_uid"),
		version:    req.PostFormValue("version"),
		extras:     req.PostFormValue("extras"),
	}

	if send.text == "" || send.token == "" || send.dialogId == "" || send.msgUid == "" || send.botUid == "" || send.version == "" {
		writeJson(w, map[string]string{
			"code":    "400",
			"message": "Parameter error",
		})
		return nil
	}

	replyId := ""
	if send.dialogType == "group" {
		replyId = send.msgId
	}

	send.id = send.callRequest("sendtext", map[string]string{
		"dialog_id": send.dialogId,
		"reply_id":  replyId,
		"text":      "...",
		"text_type": "md",
		"silence":   "yes",
	}, map[string]string{
		"version": send.version,
		"token":   send.token,
	}, false)
	if send.id == "" {
		writeJson(w, map[string]string{
			"code":    "400",
			"message": "Response failed",
		})
		return nil
	}

	go send.callRequest("stream", map[string]string{
		"dialog_id":  send.dialogId,
		"userid":     send.msgUid,
		"stream_url": "/ai/stream/" + send.id,
	}, map[string]string{
		"version": send.version,
		"token":   send.token,
	}, false)

	return send
}

func (send *sendModel) callRequest(action string, data map[string]string, header map[string]string, done bool) string {
	if done {
		doneClient(send.id)
	}
	//
	tmpUrl := ServerUrl
	tmpValue := gjson.Get(send.extras, "server_url")
	if tmpValue.Exists() {
		tmpUrl = tmpValue.String()
	}
	var callUrl string
	if action == "stream" {
		callUrl = tmpUrl + "/api/dialog/msg/stream"
	} else {
		callUrl = tmpUrl + "/api/dialog/msg/sendtext"
	}
	r := gohttp.NewRequest()
	if data != nil {
		r.FormData(data)
	}
	if header != nil {
		r.Headers(header)
	}
	res, err := r.Post(callUrl)
	if err != nil || res == nil {
		return ""
	}

	body, err := res.GetBodyAsString()
	if err != nil {
		return ""
	}
	value := gjson.Get(body, "data.id")

	return value.String()
}

func writeJson(w http.ResponseWriter, m map[string]string) {
	mjson, err := json.Marshal(m)
	if err != nil {
		_, _ = w.Write([]byte("Error"))
	}
	_, _ = w.Write(mjson)
}

func getClient(id string, createAuto bool) *clientModel {
	for _, client := range clients {
		if client.id == id {
			return client
		}
	}
	if createAuto {
		client := &clientModel{
			id: id,
		}
		clients = append(clients, client)
		return client
	}
	return nil
}

func doneClient(id string) {
	go func() {
		client := getClient(id, true)
		for i := 0; i < 30; i++ {
			time.Sleep(1 * time.Second)
			client.message = "done"
			client.sendMessage("done")
		}
		client.remove()
	}()
}

func (send *sendModel) openaiContext() *openaiModel {
	key := "openai_" + send.dialogId + "_" + send.msgUid
	var value *openaiModel
	for _, oc := range openaiContext {
		if oc.key == key {
			value = oc
			break
		}
	}
	if value == nil {
		value = &openaiModel{
			key:      key,
			messages: make([]openai.ChatCompletionMessage, 0),
		}
		openaiContext = append(openaiContext, value)
	} else if len(value.messages) > 10 {
		value.messages = value.messages[len(value.messages)-10:]
	}
	value.messages = append(value.messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: send.text,
	})

	length := 0
	index := 0
	for i := len(value.messages) - 1; i >= 0; i-- {
		length += len(value.messages[i].Content)
		if length > 4000 {
			value.messages = value.messages[len(value.messages)-index:]
			break
		}
		index++
	}
	return value
}

func (send *sendModel) openaiContextClear() {
	key := "openai_" + send.dialogId + "_" + send.msgUid
	for i, oc := range openaiContext {
		if oc.key == key {
			openaiContext = append(openaiContext[:i], openaiContext[i+1:]...)
			break
		}
	}
}

func (client *clientModel) claudeResponse(response chan types.PartialResponse) {
	client.append = ""
	client.message = ""
	number := 0
	for {
		message, ok := <-response
		if !ok {
			return
		}
		if message.Error != nil {
			return
		}
		client.append = message.Text
		client.message = fmt.Sprintf("%s%s", client.message, client.append)
		//
		if number == 0 || len(client.message) < 100 {
			client.sendMessage("replace")
		} else {
			client.sendMessage("append")
		}
		if number > 20 {
			number = 0
		} else {
			number++
		}
	}
}

func (client *clientModel) openaiStream(stream *openai.ChatCompletionStream, chunkSize int) {
	client.append = ""
	client.message = ""
	number := 0
	if chunkSize < 1 {
		chunkSize = 7
	}
	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return
		}
		if err != nil {
			return
		}
		message := response.Choices[0].Delta.Content
		client.append = fmt.Sprintf("%s%s", client.append, message)
		client.message = fmt.Sprintf("%s%s", client.message, message)
		if number == 0 || len(client.message) < 10 {
			client.sendMessage("replace")
			client.append = ""
		} else if utf8.RuneCountInString(client.append) >= chunkSize {
			client.sendMessage("append")
			client.append = ""
		}
		if number > 20 {
			number = 0
		} else {
			number++
		}
	}
}

func (client *clientModel) wenxinStream(stream aicustomv1.WenxinworkshopService_ChatCompletionsStreamClient) {
	client.append = ""
	client.message = ""
	number := 0
	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return
		}
		if err != nil {
			return
		}
		client.append = response.Result
		client.message = fmt.Sprintf("%s%s", client.message, client.append)
		//
		if number == 0 || len(client.message) < 100 {
			client.sendMessage("replace")
		} else {
			client.sendMessage("append")
		}
		if number > 20 {
			number = 0
		} else {
			number++
		}
	}
}

func (send *sendModel) wenxinContext() *wenxinModel {
	user := "wenxin_" + send.dialogId + "_" + send.msgUid
	var value *wenxinModel
	for _, oc := range wenxinContext {
		if oc.user == user {
			value = oc
			break
		}
	}
	if value == nil {
		value = &wenxinModel{
			user:     user,
			messages: make([]*aicustomv1.Message, 0),
		}
		wenxinContext = append(wenxinContext, value)
	} else if len(value.messages) > 10 {
		value.messages = value.messages[len(value.messages)-10:]
	}
	value.messages = append(value.messages, &aicustomv1.Message{
		Role:    "user",
		Content: send.text,
	})
	length := 0
	index := 0
	for i := len(value.messages) - 1; i >= 0; i-- {
		length += len(value.messages[i].Content)
		if length > 4000 {
			value.messages = value.messages[len(value.messages)-index:]
			break
		}
		index++
	}
	return value
}

func (send *sendModel) wenxinContextClear() {
	user := "wenxin_" + send.dialogId + "_" + send.msgUid
	for i, oc := range wenxinContext {
		if oc.user == user {
			wenxinContext = append(wenxinContext[:i], wenxinContext[i+1:]...)
			break
		}
	}
}
func (send *sendModel) qianwenContext() *qianwenModel {
	user := "qianwen_" + send.dialogId + "_" + send.msgUid
	var value *qianwenModel
	for _, oc := range qianwenContext {
		if oc.user == user {
			value = oc
			break
		}
	}
	if value == nil {
		value = &qianwenModel{
			user:     user,
			messages: make([]*qianwenconfig.HistoryResquest, 0),
		}
		qianwenContext = append(qianwenContext, value)
	} else if len(value.messages) > 10 {
		value.messages = value.messages[len(value.messages)-10:]
	}
	value.messages = append(value.messages, &qianwenconfig.HistoryResquest{
		User: "user",
		Bot:  send.text,
	})
	length := 0
	index := 0
	for i := len(value.messages) - 1; i >= 0; i-- {
		length += len(value.messages[i].Bot)
		if length > 4000 {
			value.messages = value.messages[len(value.messages)-index:]
			break
		}
		index++
	}
	return value
}
func (send *sendModel) qianwenContextClear() {
	user := "qianwen_" + send.dialogId + "_" + send.msgUid
	for i, oc := range qianwenContext {
		if oc.user == user {
			qianwenContext = append(qianwenContext[:i], qianwenContext[i+1:]...)
			break
		}
	}
}
func (client *clientModel) qianwenStream(cli *qianwen.Client) {
	client.append = ""
	client.message = ""
	number := 0
	for {
		resp, ok := <-cli.Sender
		if !ok {
			return
		}
		client.append = resp.Message[len(client.message):]
		client.message = resp.Message
		//
		if number == 0 || len(client.message) < 100 {
			client.sendMessage("replace")
		} else {
			client.sendMessage("append")
		}
		if number > 20 {
			number = 0
		} else {
			number++
		}
	}
}

func (send *sendModel) geminiContext() *geminiModel {
	user := "gemini_" + send.dialogId + "_" + send.msgUid
	var value *geminiModel
	for _, oc := range geminiContext {
		if oc.user == user {
			value = oc
			break
		}
	}
	if value == nil {
		value = &geminiModel{
			user:     user,
			messages: make([]*genai.Content, 0),
		}
		geminiContext = append(geminiContext, value)
		return value
	} else if len(value.messages) > 10 {
		value.messages = value.messages[len(value.messages)-10:]
	}
	length := 0
	index := 0
	for i := len(value.messages) - 1; i >= 0; i-- {
		length += len(value.messages[i].Role)
		if length > 4000 {
			value.messages = value.messages[len(value.messages)-index:]
			break
		}
		index++
	}
	return value
}

func (client *clientModel) geminiStream(cli *gemini.GeminiClient, history []*genai.Content) ([]*genai.Content, error) {
	prompt := genai.Text(client.message)
	cs := cli.Model.StartChat()
	cs.History = history
	iter := cs.SendMessageStream(context.Background(), prompt)
	number := 0
	client.message = ""
	client.append = ""
	for {
		resp, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return cs.History, err
		}
		s, ok := resp.Candidates[0].Content.Parts[0].(genai.Text)
		if ok {
			msg := string(s)
			client.append = fmt.Sprintf("%s%s", client.append, msg)
			client.message = fmt.Sprintf("%s%s", client.message, msg)
			if number == 0 || len(client.message) < 10 {
				client.sendMessage("replace")
				client.append = ""
			} else {
				client.sendMessage("append")
				client.append = ""
			}
			if number > 20 {
				number = 0
			} else {
				number++
			}
		}
	}
	return cs.History, nil
}
func (send *sendModel) geminiContextClear() {
	user := "gemini_" + send.dialogId + "_" + send.msgUid
	for i, oc := range geminiContext {
		if oc.user == user {
			geminiContext = append(geminiContext[:i], geminiContext[i+1:]...)
			break
		}
	}
}

func (client *clientModel) sendMessage(event string) {
	if event == "append" {
		sources.SendMessage("/stream/"+client.id, sse.NewMessage(client.id, client.append, event))
	} else {
		sources.SendMessage("/stream/"+client.id, sse.NewMessage(client.id, client.message, event))
	}
}

func (client *clientModel) remove() {
	for i, c := range clients {
		if c == client {
			clients = append(clients[:i], clients[i+1:]...)
			return
		}
	}
}
