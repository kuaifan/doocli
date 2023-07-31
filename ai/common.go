package ai

import (
	"doocli/ai/claude/types"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/alexandrevicenzi/go-sse"
	"github.com/nahid/gohttp"
	"github.com/sashabaranov/go-openai"
	"github.com/tidwall/gjson"
	"io"
	"net/http"
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
	})
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
	})

	return send
}

func (send *sendModel) callRequest(action string, data map[string]string, header map[string]string) string {
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
		client.append = message.Text[len(client.message):]
		client.message = message.Text
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

func (client *clientModel) openaiStream(stream *openai.ChatCompletionStream) {
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
		client.append = response.Choices[0].Delta.Content
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
