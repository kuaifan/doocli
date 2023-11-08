package ai

import (
	"context"
	"doocli/utils"
	"net/http"

	"github.com/hitosea/go-wenxin/baidubce"
	aicustomv1 "github.com/hitosea/go-wenxin/gen/go/baidubce/ai_custom/v1"
	baidubcev1 "github.com/hitosea/go-wenxin/gen/go/baidubce/v1"
	"github.com/tidwall/gjson"
)

func WenxinSend(w http.ResponseWriter, req *http.Request) {
	send := callSend(w, req)
	if send == nil {
		return
	}

	tmpKey := WenxinKey
	tmpSecret := WenxinSecret
	tmpModel := WenxinModel
	tmpValue := gjson.Get(send.extras, "wenxin_key")
	if tmpValue.Exists() {
		tmpKey = tmpValue.String()
	}
	tmpValue = gjson.Get(send.extras, "wenxin_secret")
	if tmpValue.Exists() {
		tmpSecret = tmpValue.String()
	}
	tmpValue = gjson.Get(send.extras, "wenxin_model")
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
		sendtext["text"] = "OpenaiKey is empty"
		writeJson(w, map[string]string{"code": "400", "message": sendtext["text"]})
		send.callRequest("sendtext", sendtext, tokens)
		return
	}

	if utils.InArray(send.text, clears) {
		send.wenxinContextClear()
		sendtext["text"] = "Operation Successful"
		send.callRequest("sendtext", sendtext, tokens)
		return
	}

	go func() {
		wenxinClient, err := baidubce.New(baidubce.WithTokenRequest(&baidubcev1.TokenRequest{
			GrantType:    "client_credentials",
			ClientId:     tmpKey,
			ClientSecret: tmpSecret,
		}))
		if err != nil {
			writeJson(w, map[string]string{"code": "400", "message": err.Error()})
			sendtext["text"] = err.Error()
			send.callRequest("sendtext", sendtext, tokens)
			return
		}

		oc := send.wenxinContext()
		stream, err := wenxinClient.ChatStream(context.Background(), &aicustomv1.ChatCompletionsRequest{
			User:     "wenxin_" + send.dialogId + "_" + send.msgUid,
			Messages: oc.messages,
		}, tmpModel)
		if err != nil {
			writeJson(w, map[string]string{"code": "400", "message": err.Error()})
			sendtext["text"] = err.Error()
			send.callRequest("sendtext", sendtext, tokens)
			return
		}

		defer func(stream aicustomv1.WenxinworkshopService_ChatCompletionsStreamClient) {
			_ = stream.CloseSend()
		}(stream)
		client := getClient(send.id, true)
		client.wenxinStream(stream)
		if client.message == "" {
			client.message = "empty"
		}
		sendtext["text"] = client.message
		oc.messages = append(oc.messages, &aicustomv1.Message{
			Role:    "assistant",
			Content: sendtext["text"],
		})
		client.sendMessage("done")
		client.remove()

		send.callRequest("sendtext", sendtext, tokens)
	}()
	//
	writeJson(w, map[string]string{
		"code":   "200",
		"msg_id": send.id,
	})
}
