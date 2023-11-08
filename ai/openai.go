package ai

import (
	"context"
	"doocli/utils"
	"net/http"
	"net/url"

	"github.com/sashabaranov/go-openai"
	"github.com/tidwall/gjson"
)

func OpenaiSend(w http.ResponseWriter, req *http.Request) {
	send := callSend(w, req)
	if send == nil {
		return
	}
	tmpKey := OpenaiKey
	tmpAgency := OpenaiAgency
	tmpModel := openai.GPT3Dot5Turbo
	tmpValue := gjson.Get(send.extras, "openai_key")
	if tmpValue.Exists() {
		tmpKey = tmpValue.String()
	}
	tmpValue = gjson.Get(send.extras, "openai_agency")
	if tmpValue.Exists() {
		tmpAgency = tmpValue.String()
	}
	tmpValue = gjson.Get(send.extras, "openai_model")
	if tmpValue.Exists() {
		tmpModel = tmpValue.String()
	}
	if tmpKey == "" {
		writeJson(w, map[string]string{
			"code":    "400",
			"message": "OpenaiKey is empty",
		})
		send.callRequest("sendtext", map[string]string{
			"update_id":   send.id,
			"update_mark": "no",
			"dialog_id":   send.dialogId,
			"text":        "openai key is empty",
			"text_type":   "md",
			"silence":     "yes",
		}, map[string]string{
			"version": send.version,
			"token":   send.token,
		})
		return
	}

	if utils.InArray(send.text, clears) {
		send.openaiContextClear()
		send.callRequest("sendtext", map[string]string{
			"update_id":   send.id,
			"update_mark": "no",
			"dialog_id":   send.dialogId,
			"text":        "Operation Successful",
			"text_type":   "md",
			"silence":     "yes",
		}, map[string]string{
			"version": send.version,
			"token":   send.token,
		})
		return
	}

	go func() {
		var oa *openai.Client
		if tmpAgency != "" {
			config := openai.DefaultConfig(tmpKey)
			proxyUrl, err := url.Parse(tmpAgency)
			if err != nil {
				panic(err)
			}
			transport := &http.Transport{
				Proxy: http.ProxyURL(proxyUrl),
			}
			config.HTTPClient = &http.Client{
				Transport: transport,
			}
			oa = openai.NewClientWithConfig(config)
		} else {
			oa = openai.NewClient(tmpKey)
		}
		oc := send.openaiContext()
		stream, err := oa.CreateChatCompletionStream(
			context.Background(),
			openai.ChatCompletionRequest{
				Model:    tmpModel,
				Messages: oc.messages,
				Stream:   true,
			},
		)
		if err != nil {
			writeJson(w, map[string]string{
				"code":    "400",
				"message": err.Error(),
			})
			message := err.Error()
			if message == "" {
				message = "Claude Create Chat Error, Please Try Again Later"
			}
			send.callRequest("sendtext", map[string]string{
				"update_id":   send.id,
				"update_mark": "no",
				"dialog_id":   send.dialogId,
				"text":        message,
				"text_type":   "md",
				"silence":     "yes",
			}, map[string]string{
				"version": send.version,
				"token":   send.token,
			})
			return
		}
		defer stream.Close()

		client := getClient(send.id, true)
		client.openaiStream(stream)
		if client.message == "" {
			client.message = "empty"
		}
		message := client.message
		oc.messages = append(oc.messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: message,
		})
		client.sendMessage("done")
		client.remove()
		send.callRequest("sendtext", map[string]string{
			"update_id":   send.id,
			"update_mark": "no",
			"dialog_id":   send.dialogId,
			"text":        message,
			"text_type":   "md",
			"silence":     "yes",
		}, map[string]string{
			"version": send.version,
			"token":   send.token,
		})
	}()
	//
	writeJson(w, map[string]string{
		"code":   "200",
		"msg_id": send.id,
	})
}
