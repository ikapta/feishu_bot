package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/ikapta/bot"
	"github.com/ikapta/event"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "hello world")
}

func htmlHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/html")
    html := `<!doctype  html>
    <META  http-equiv="Content-Type"  content="text/html"  charset="utf-8">
    <html  lang="zhCN">
      <head>
        <title>Golang</title>
        <meta  name="viewport"  content="width=device-width,  initial-scale=1.0,  maximum-scale=1.0,  user-scalable=0;"  />
      </head>

      <body>
          <div id="app">Welcome!</div>
      </body>
    </html>`
    fmt.Fprintf(w, html)
}

func botEventHandler(w http.ResponseWriter, r *http.Request) {
  isUrlVerify := false

  var conf = map[string]string{
    "app_id": "//todo fill your app_id",
    "app_secret": "//todo fill your app_secret",
    "encrypt_key": "//todo fill your encrypt_key",
  }

  b, _ := event.NewBotEvent(event.BotEventOpt{
  	App_id:              conf["app_id"],
  	App_secret:          conf["app_secret"],
  	Tenant_access_token: "",
  })

  reqBody, _ := ioutil.ReadAll(r.Body)

  _ = event.Setup(event.EventOptions{
  	Encrypt_key: conf["encrypt_key"],
  	MsgBody:     reqBody,
  	UrlVerificationEvent: func(challenge string) error {
      isUrlVerify = true

      jsonBytes, _ := json.MarshalIndent(&map[string]string{
        "challenge": challenge,
      }, "", "   ")

      w.Write(jsonBytes)
      return nil
  	},
  	ReceiveTextEventHandler: func(receiveEvent event.ReceiveEventType) error {
  		open_id := receiveEvent.Event.Sender.Sender_id.Open_id
  		chat_id := receiveEvent.Event.Message.Chat_id
  		receive_text, _ := event.GetTextContent(receiveEvent)
  		b.TextReplyByChatId("收到消息：" + receive_text + bot.AtUserInPost(open_id), chat_id)
  		// signal = true
  		return nil
  	},
  })

  if !isUrlVerify {
    fmt.Fprintf(w, "msg event")
  }
}

func main() {
    mux := http.NewServeMux()

    mux.Handle("/", http.HandlerFunc(indexHandler))

    mux.Handle("/hello", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "hello world")
    }))

    mux.Handle("/bot/event", http.HandlerFunc(botEventHandler))

    http.ListenAndServe(":8003", mux)
}