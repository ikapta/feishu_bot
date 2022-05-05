package bot

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Bot interface {
  SendText(content string) error
  SendPost(p Post, ps ...Post) error
  SendImage(imageKey string) error
  SendCard(bgColor CardTitleBgColor, cfg CardConfig, c Card, more ...Card) error
}

type BotOpt struct {
	webhook string
	secretKey string
}

func NewBot(webhook, secretKey string) Bot {
	b := &BotOpt{
		secretKey: strings.TrimSpace(secretKey),
	}

	if !strings.Contains(webhook, "open.feishu.cn") {
    b.webhook = fmt.Sprintf(WebhookFormat, webhook)
  } else {
    b.webhook = webhook
  }

	return b
}

func (b *BotOpt) SendText(content string) error {
	return b.send(NewText(content))
}

func (b *BotOpt) SendImage(imageKey string) error {
  return b.send(NewImage(imageKey))
}

func (b *BotOpt) SendCard(bgColor CardTitleBgColor, cfg CardConfig, c Card, more ...Card) error {
  return b.send(NewCard(bgColor, cfg, c, more...))
}

func (b *BotOpt) SendPost(p Post, ps ...Post) error {
	return b.send(NewPost(p, ps...))
}

func (b *BotOpt) send(msg map[string]interface{}) (err error) {
	if b.secretKey != "" {
		ts := time.Now().Unix()
		signed, err := genSign(b.secretKey, ts)
		if err != nil {
			return err
		}
		msg["timestamp"] = ts
		msg["sign"] = signed
	}

	var msgBody []byte
	if msgBody, err = json.Marshal(msg); err != nil {
		return err
	}

  err = execPost(b.webhook, msgBody)
	return err
}
