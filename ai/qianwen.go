package ai

import (
	"context"
	"doocli/ai/qianwen"
	"doocli/ai/qianwen/config"
	"doocli/utils"
	"github.com/tidwall/gjson"
	"net/http"
)

func QianWenSend(w http.ResponseWriter, req *http.Request) {
	send := callSend(w, req)
	if send == nil {
		return
	}

	tmpKey := QianwenKey
	tmpModel := QianwenModel
	tmpValue := gjson.Get(send.extras, "qianwen_key")
	if tmpValue.Exists() {
		tmpKey = tmpValue.String()
	}
	tmpValue = gjson.Get(send.extras, "qianwen_model")
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
		send.wenxinContextClear()
		sendtext["text"] = "Operation Successful"
		send.callRequest("sendtext", sendtext, tokens, true)
		return
	}

	go func() {
		oc := send.qianwenContext()
		qianwenClient, err := qianwen.New(context.Background(), tmpKey, map[string]interface{}{
			"model": tmpModel,
			"input": config.InputResquest{
				Message: send.text,
				History: oc.messages,
			},
		})
		if err != nil {
			writeJson(w, map[string]string{"code": "400", "message": err.Error()})
			sendtext["text"] = err.Error()
			send.callRequest("sendtext", sendtext, tokens, true)
			return
		}
		err = qianwenClient.ChatStream()
		if err != nil {
			writeJson(w, map[string]string{"code": "400", "message": err.Error()})
			sendtext["text"] = err.Error()
			send.callRequest("sendtext", sendtext, tokens, true)
			return
		}
		client := getClient(send.id, true)
		client.qianwenStream(qianwenClient)
		if client.message == "" {
			client.message = "empty"
		}
		sendtext["text"] = client.message
		oc.messages = append(oc.messages, &config.HistoryResquest{
			User: "assistant",
			Bot:  sendtext["text"],
		})
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
