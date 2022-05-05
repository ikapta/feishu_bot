package event

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

const (
  APP_ACCESS_TOKEN_URI = "https://open.feishu.cn/open-apis/auth/v3/app_access_token/internal"
  TENANT_ACCESS_TOKEN_URI = "https://open.feishu.cn/open-apis/auth/v3/tenant_access_token"
  REPLY_MESSAGE_URI = "https://open.feishu.cn/open-apis/im/v1/messages/:message_id/reply"
  SEND_MESSAGE_URI = "https://open.feishu.cn/open-apis/im/v1/messages?receive_id_type=:receive_id_type"
)

type BotEvent interface {
  TextReplyByMsgId(msg, msgId string) error
  TextReplyByChatId(msg, receive_id string) error
  ProcessReply(isReplyMsg bool, tenant_access_token, receive_id, msgType string, content interface{}) error
}

type BotEventOpt struct {
  App_id string
  App_secret string
  Tenant_access_token string // token expire in 2 hour as default, can be set to session
}

func NewBotEvent(opts BotEventOpt) (BotEvent, error) {
  b := &BotEventOpt{
    App_id: opts.App_id,
    App_secret: opts.App_secret,
    Tenant_access_token: opts.Tenant_access_token,
  }

  var err error

  if len(b.Tenant_access_token) == 0 {
    b.Tenant_access_token, err = get_app_access_token(b.App_id, b.App_secret)
  }

  return b, err;
}


// reply msg defined in https://open.feishu.cn/document/uAjLw4CM/ukTMukTMukTM/reference/im-v1/message/reply
func (b *BotEventOpt) TextReplyByMsgId(msg, msgId string) error {
  jsonText := map[string]string{
    "text": msg,
  }

  jsonStr, _ := json.Marshal(jsonText)

  err := b.ProcessReply(
    true,
    b.Tenant_access_token,
    msgId,
    "text",
    fmt.Sprintf("%s", jsonStr),
  )

  return err
}

// send msg defined in https://open.feishu.cn/document/uAjLw4CM/ukTMukTMukTM/reference/im-v1/message/create
func (b *BotEventOpt) TextReplyByChatId(msg, receive_id string) error {
  jsonText := map[string]string{
    "text": msg,
  }

  jsonStr, _ := json.Marshal(jsonText)

  err := b.ProcessReply(
    false,
    b.Tenant_access_token,
    receive_id,
    "text",
    fmt.Sprintf("%s", jsonStr),
  )

  return err
}

/**
 * isReplyMsg: true for reply msg, false for send msg
 * receive_id: chat_id|open_id for send msg or msg_id for reply msg
 */
func (b *BotEventOpt) ProcessReply(isReplyMsg bool, tenant_access_token, receive_id, msgType string, content interface{}) error {
  var uri string
  if (isReplyMsg) {
    uri = strings.ReplaceAll(REPLY_MESSAGE_URI, ":message_id", receive_id)
  } else {
    //receive_id_type defined in https://open.feishu.cn/document/uAjLw4CM/ukTMukTMukTM/reference/im-v1/message/create#bc6d1214
    uri = strings.ReplaceAll(SEND_MESSAGE_URI, ":receive_id_type", "chat_id")
  }

  body := map[string]interface{}{
    "receive_id": receive_id,
    "msg_type": msgType,
    "content": content,
  }
  bodystr, _ := json.Marshal(body)

  client := &http.Client{}
  req, err := http.NewRequest("POST", uri, bytes.NewReader(bodystr))

  if err != nil {
    return err
  }

  req.Header.Set("Authorization", "Bearer " + tenant_access_token)
  req.Header.Set("Content-Type", "application/json")

  resp, err := client.Do(req)
  defer resp.Body.Close()

  if err != nil {
    return err
  }

  respBody, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    return err
  }

  if (resp.StatusCode != 200) {
    var data map[string]interface{}
    json.Unmarshal(respBody, &data)
    return fmt.Errorf("send message error: %s", data["msg"])
  }

  return nil
}

// https://open.feishu.cn/document/ukTMukTMukTM/ukDNz4SO0MjL5QzM/auth-v3/auth/app_access_token_internal
func get_app_access_token(app_id, app_secret string) (string, error) {
  body := map[string]string{
    "app_id": app_id,
    "app_secret": app_secret,
  }
  bodystr, _ := json.Marshal(body)

  client := &http.Client{}
  req, err := http.NewRequest("POST", APP_ACCESS_TOKEN_URI, bytes.NewReader(bodystr))
  if err != nil {
    return "", err
  }

  req.Header.Set("Content-Type", "application/json; charset=utf-8")

  resp, err := client.Do(req)
  if err != nil {
    return "", err
  }

  defer resp.Body.Close()

  if err != nil {
    return "", err
  }

  respBody, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    return "", err
  }

  var data map[string]interface{}
  if err := json.Unmarshal(respBody, &data); err != nil {
    return "", err
  }

  // may be you need -> data["app_access_token"]

  if (data["tenant_access_token"] == "") {
    return "", fmt.Errorf("get tenant_access_token error: %s", data["msg"])
  }

  return data["tenant_access_token"].(string), nil
}

