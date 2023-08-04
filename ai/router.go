package ai

import (
	"fmt"
	"net/http"

	"github.com/alexandrevicenzi/go-sse"
)

func Start() {
	sources = sse.NewServer(&sse.Options{
		Headers: map[string]string{
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "GET, OPTIONS",
			"Access-Control-Allow-Headers": "Keep-Alive,X-Requested-With,Cache-Control,Content-Type,Last-Event-ID",
		},
	})
	defer sources.Shutdown()
	//
	http.Handle("/stream/", sources)
	http.HandleFunc("/claude/send", ClaudeSend)
	http.HandleFunc("/openai/send", OpenaiSend)
	http.HandleFunc("/wenxin/send", WenxinSend)
	http.HandleFunc("/qianwen/send", QianWenSend)
	//
	fmt.Println("AI service started, listening on port: " + HttpPort)
	_ = http.ListenAndServe(":"+HttpPort, nil)
}