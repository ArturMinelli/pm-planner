package message

// Message is a dot-notation i18n key with optional interpolation params.
type Message struct {
	Key    string            `json:"key"`
	Params map[string]string `json:"params,omitempty"`
}

// New builds a message with the given key and params.
func New(key string, params map[string]string) Message {
	if params == nil {
		return Message{Key: key}
	}
	copied := make(map[string]string, len(params))
	for key, value := range params {
		copied[key] = value
	}
	return Message{Key: key, Params: copied}
}

// Ptr returns a pointer to a new message.
func Ptr(key string, params map[string]string) *Message {
	message := New(key, params)
	return &message
}
