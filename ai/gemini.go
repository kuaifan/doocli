package ai

import (
	"context"
	"doocli/ai/gemini"
	"doocli/utils"
	"github.com/google/generative-ai-go/genai"
	"github.com/tidwall/gjson"
	"google.golang.org/api/option"
	"net/http"
)

func GeminiSend(w http.ResponseWriter, req *http.Request) {
	send := callSend(w, req)
	if send == nil {
		return
	}
	tmpKey := GeminiKey
	tmpModel := GeminiModel
	tmpValue := gjson.Get(send.extras, "gemini_key")
	if tmpValue.Exists() {
		tmpKey = tmpValue.String()
	}
	tmpValue = gjson.Get(send.extras, "gemini_model")
	if tmpValue.Exists() {
		tmpModel = tmpValue.String()
	}

	sendtext := map[string]string{
		"update_id":   send.id,
		"update_mark": "no",
		"dialog_id":   send.dialogId,
		"text":        "Operation Successful",
		"text_type":   "md",
		"silence":     "yes",
	}
	tokens := map[string]string{
		"version": send.version,
		"token":   send.token,
	}

	if tmpKey == "" {
		writeJson(w, map[string]string{
			"code":    "400",
			"message": "OpenaiKey is empty",
		})
		send.callRequest("sendtext", sendtext, tokens, true)
		return
	}

	if utils.InArray(send.text, clears) {
		send.geminiContextClear()
		sendtext["text"] = "Operation Successful"
		send.callRequest("sendtext", sendtext, tokens, true)
		return
	}
	go func() {
		client2, err := genai.NewClient(context.Background(), option.WithAPIKey(tmpKey))
		if err != nil {
			sendtext["text"] = "err：" + err.Error()
			send.callRequest("sendtext", sendtext, tokens, true)
		}

		gemiCLient := gemini.NewGemniClient(client2, tmpModel)
		client := getClient(send.id, true)
		client.message = send.text
		model := send.geminiContext()
		//设置上下文
		model.messages = append(client.geminiStream(gemiCLient, model.messages))
		if client.message == "" {
			client.message = "empty"
		}
		sendtext["text"] = client.message
		client.sendMessage("done")
		client.remove()
		send.callRequest("sendtext", sendtext, tokens, false)
	}()

	writeJson(w, map[string]string{
		"code":   "200",
		"msg_id": send.id,
	})
	return
}
