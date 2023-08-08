package ai

import (
	"context"
	"doocli/ai/claude"
	"doocli/ai/claude/vars"
	"doocli/db"
	"doocli/utils"
	"github.com/tidwall/gjson"
	"net/http"
)

func ClaudeSend(w http.ResponseWriter, req *http.Request) {
	send := callSend(w, req)
	if send == nil {
		return
	}
	tmpToken := ClaudeToken
	tmpAgency := ClaudeAgency
	tmpValue := gjson.Get(send.extras, "claude_token")
	if tmpValue.Exists() {
		tmpToken = tmpValue.String()
	}
	tmpValue = gjson.Get(send.extras, "claude_agency")
	if tmpValue.Exists() {
		tmpAgency = tmpValue.String()
	}
	if tmpToken == "" {
		writeJson(w, map[string]string{
			"code":    "400",
			"message": "ClaudeToken is empty",
		})
		send.callRequest("sendtext", map[string]string{
			"update_id":   send.id,
			"update_mark": "no",
			"dialog_id":   send.dialogId,
			"text":        "claude token is empty",
			"text_type":   "md",
			"silence":     "yes",
		}, map[string]string{
			"version": send.version,
			"token":   send.token,
		})
		return
	}

	organizationKey := "organization_" + send.dialogId + "_" + send.msgUid
	conversationKey := "conversation_" + send.dialogId + "_" + send.msgUid
	if utils.InArray(send.text, clears) {
		_ = db.DelConfig(organizationKey)
		_ = db.DelConfig(conversationKey)
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

	options := claude.NewDefaultOptions(tmpToken, "", vars.Model4WebClaude2)
	if tmpAgency != "" {
		options.Agency = tmpAgency
	}
	options.OrganizationId = db.GetConfigString(organizationKey)
	options.ConversationId = db.GetConfigString(conversationKey)
	chat, err := claude.New(options)
	if err != nil {
		writeJson(w, map[string]string{
			"code":    "400",
			"message": err.Error(),
		})
		send.callRequest("sendtext", map[string]string{
			"update_id":   send.id,
			"update_mark": "no",
			"dialog_id":   send.dialogId,
			"text":        err.Error(),
			"text_type":   "md",
			"silence":     "yes",
		}, map[string]string{
			"version": send.version,
			"token":   send.token,
		})
		return
	}

	go func() {
		response, err := chat.Reply(context.Background(), send.text, nil)
		var message string
		if err != nil {
			message = err.Error()
		} else {
			_ = db.SetConfig(organizationKey, chat.GetOptions().OrganizationId)
			_ = db.SetConfig(conversationKey, chat.GetOptions().ConversationId)
			client := getClient(send.id, true)
			client.claudeResponse(response)
			client.sendMessage("done")
			message = client.message
			client.remove()
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
	}()
	//
	writeJson(w, map[string]string{
		"code":   "200",
		"msg_id": send.id,
	})
}
