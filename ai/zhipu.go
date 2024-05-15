package ai

import (
	"doocli/ai/zhipu/model_api"
	"doocli/utils"
	"github.com/tidwall/gjson"
	"net/http"
)

func ZhiPuSend(w http.ResponseWriter, req *http.Request) {
	send := callSend(w, req)
	if send == nil {
		return
	}

	tmpKey := ZhipuKey
	tmpModel := ZhipuModel

	tmpValue := gjson.Get(send.extras, "zhipu_key")
	if tmpValue.Exists() {
		tmpKey = tmpValue.String()
	}
	tmpValue = gjson.Get(send.extras, "zhipu_model")
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
			"message": "ZhiPuKey is empty",
		})
		send.callRequest("sendtext", sendtext, tokens, true)
		return
	}

	if utils.InArray(send.text, clears) {
		send.zhipuContextClear()
		sendtext["text"] = "Operation Successful"
		send.callRequest("sendtext", sendtext, tokens, true)
		return
	}

	go func() {
		oc := send.zhipuContext()
		body := &model_api.PostParams{
			Stream:   true,
			Model:    tmpModel,
			Messages: oc.messages,
		}
		postResponse, err := model_api.BeCommonModelStream(ZhipuExpireAtTime, body, nil, tmpKey)
		if err != nil {
			sendtext["text"] = "err：" + err.Error()
			send.callRequest("sendtext", sendtext, tokens, true)
			return
		}

		client := getClient(send.id, true)
		client.message = send.text

		err = client.zhipuStream(postResponse)
		if err != nil {
			sendtext["text"] = "err：" + err.Error()
			send.callRequest("sendtext", sendtext, tokens, true)
			return
		}

		oc.messages = append(oc.messages, &model_api.Messages{
			Role:    "assistant",
			Content: client.message,
		})
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
