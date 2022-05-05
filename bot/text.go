package bot

func NewText(text string) map[string]interface{} {
	msgType := "text"
	return map[string]interface{}{
		"msg_type": msgType,
		"content": map[string]string{
			msgType: text,
		},
	}
}
