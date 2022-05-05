package event

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const (
  URL_VERIFICATION = "url_verification"
  IM_MESSAGE_RECEIVE_V1 = "im.message.receive_v1"
)

type EventOptions struct {
  Encrypt_key string
  MsgBody []byte
  UrlVerificationEvent func(challenge string) error
  ReceiveTextEventHandler func(receiveEvent ReceiveEventType) error
  ReceiveAnyEventHandler func(receiveEvent ReceiveEventType) error
}

// partial json field that defined in https://open.feishu.cn/document/uAjLw4CM/ukTMukTMukTM/reference/im-v1/message/events/receive
type ReceiveEventType struct {
  Schema string `json:"schema"`
  Header struct {
    Event_id string `json:"event_id"`
    Event_type string `json:"event_type"`
    Tenant_key string `json:"tenant_key"`
    App_id string `json:"app_id"`
  }
  Event struct {
    Message struct {
      Chat_id string `json:"chat_id"`
      Content string `json:"content"`
      Message_id string `json:"message_id"`
      Message_type string `json:"message_type"`
    }
    Sender struct {
      Sender_id struct {
        Open_id string `json:"open_id"`
        User_id string `json:"user_id"`
        Union_id string `json:"union_id"`
      }
    }
  }
}

func Setup(options EventOptions) error {
  var (
    encrypt_key = options.Encrypt_key
    msgBody = options.MsgBody
    urlVerificationEvent = options.UrlVerificationEvent
    receiveTextEventHandler = options.ReceiveTextEventHandler
    receiveAnyEventHandler = options.ReceiveAnyEventHandler
  )

  encryptMsgJson := struct {
		Encrypt string `json:"encrypt"`
	}{}

  if err := json.Unmarshal(msgBody, &encryptMsgJson); err != nil {
    return fmt.Errorf("Parse error: %v\n", err)
  }

  if len(encryptMsgJson.Encrypt) == 0 {
    return fmt.Errorf("Error: missing encrypt!")
  }

  decryptMsgText, err := decrypt(encryptMsgJson.Encrypt, encrypt_key)
  if err != nil {
    return fmt.Errorf("Decrypt error: %v\n", err)
  }

  var receiveEventMap = make(map[string]interface{})

  if err := json.Unmarshal([]byte(decryptMsgText), &receiveEventMap); err != nil {
    return fmt.Errorf("Parse error: %v\n", err)
  }

  // check if url_verification, else is receive msg event.
  // https://open.feishu.cn/document/ukTMukTMukTM/uYDNxYjL2QTM24iN0EjN/event-subscription-configure-/request-url-configuration-case
  if receiveEventMap["type"] == URL_VERIFICATION {
    if urlVerificationEvent == nil {
      return fmt.Errorf("UrlVerificationEvent is nil")
    }

    return urlVerificationEvent(receiveEventMap["challenge"].(string))
  }

  // process receive msg event.
  // the field defined in https://open.feishu.cn/document/uAjLw4CM/ukTMukTMukTM/reference/im-v1/message/events/receive
  decryptEventMsg := ReceiveEventType{}
  if err := json.Unmarshal([]byte(decryptMsgText), &decryptEventMsg); err != nil {
    return fmt.Errorf("Parse error: %v\n", err)
  }

  var (
    schema = decryptEventMsg.Schema
    eventType = decryptEventMsg.Header.Event_type
    msgType = decryptEventMsg.Event.Message.Message_type
  )

  if len(schema) == 0 {
    return fmt.Errorf("Error: schema is nil!")
  }

  if eventType == IM_MESSAGE_RECEIVE_V1 {
    if msgType == "text" {
      if receiveTextEventHandler == nil {
        return fmt.Errorf("receiveTextEventHandler is nil")
      }
      return receiveTextEventHandler(decryptEventMsg)
    }

    if (receiveAnyEventHandler == nil) {
      return fmt.Errorf(`Error: the msg_type for [%s] is ignored, if you need pls impl func ReceiveAnyEventHandler!`, msgType)
    }
    return receiveAnyEventHandler(decryptEventMsg)
  }

  return nil
}

func GetTextContent(receiveEvent ReceiveEventType) (string, error) {
  var m = make(map[string]string)
  err := json.Unmarshal([]byte(receiveEvent.Event.Message.Content), &m)

  if err != nil {
    return "", fmt.Errorf("Parse error: %v\n", err)
  }
  return m["text"], err
}

func decrypt(encrypt string, key string) (string, error) {
	buf, err := base64.StdEncoding.DecodeString(encrypt)
	if err != nil {
		return "", fmt.Errorf("base64StdEncode Error[%v]", err)
	}
	if len(buf) < aes.BlockSize {
		return "", errors.New("cipher  too short")
	}
	keyBs := sha256.Sum256([]byte(key))
	block, err := aes.NewCipher(keyBs[:sha256.Size])
	if err != nil {
		return "", fmt.Errorf("AESNewCipher Error[%v]", err)
	}
	iv := buf[:aes.BlockSize]
	buf = buf[aes.BlockSize:]
	// CBC mode always works in whole blocks.
	if len(buf)%aes.BlockSize != 0 {
		return "", errors.New("ciphertext is not a multiple of the block size")
	}
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(buf, buf)
	n := strings.Index(string(buf), "{")
	if n == -1 {
		n = 0
	}
	m := strings.LastIndex(string(buf), "}")
	if m == -1 {
		m = len(buf) - 1
	}
	return string(buf[n : m+1]), nil
}