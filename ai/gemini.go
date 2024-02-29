package ai

import (
	"context"
	"doocli/ai/gemini"
	"doocli/utils"
	"github.com/google/generative-ai-go/genai"
	"github.com/tidwall/gjson"
	"google.golang.org/api/option"
	"net/http"
	"net/url"
	"time"
)

func GeminiSend(w http.ResponseWriter, req *http.Request) {
	send := callSend(w, req)
	if send == nil {
		return
	}

	tmpKey := GeminiKey
	tmpModel := GeminiModel
	tmpProxy := GeminiAgency

	tmpValue := gjson.Get(send.extras, "gemini_key")
	if tmpValue.Exists() {
		tmpKey = tmpValue.String()
	}
	tmpValue = gjson.Get(send.extras, "gemini_model")
	if tmpValue.Exists() {
		tmpModel = tmpValue.String()
	}
	tmpValue = gjson.Get(send.extras, "gemini_agency")
	if tmpValue.Exists() {
		tmpProxy = tmpValue.String()
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
			"message": "GeminiKey is empty",
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
		proxyUrl, err := url.Parse(tmpProxy)
		if err != nil {
			panic(err)
		}
		c := &http.Client{
			Timeout: time.Second * 15, // 设置超时时间为10秒
			Transport: &gemini.CustomTransport{
				Transport: &http.Transport{
					Proxy: http.ProxyURL(proxyUrl),
				},
				Key: tmpKey,
			}}

		client2, err := genai.NewClient(context.Background(), option.WithAPIKey(tmpKey), option.WithHTTPClient(c))
		if err != nil {
			sendtext["text"] = "err：" + err.Error()
			send.callRequest("sendtext", sendtext, tokens, true)
			return
		}
		defer client2.Close()
		iter := client2.ListModels(context.Background())
		_, err = iter.Next()
		if err != nil {
			sendtext["text"] = "err：" + err.Error()
			send.callRequest("sendtext", sendtext, tokens, true)
			return
		}
		gemiCLient := gemini.NewGemniClient(client2, tmpModel)
		client := getClient(send.id, true)
		client.message = send.text
		model := send.geminiContext()

		hs, err := client.geminiStream(gemiCLient, model.messages)
		if err != nil {
			sendtext["text"] = "err：" + err.Error()
			send.callRequest("sendtext", sendtext, tokens, true)
			return
		}
		model.messages = append(hs)
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
