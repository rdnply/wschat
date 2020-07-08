package message

import "encoding/json"

type Message struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Message string `json:"message"`
}

func (m *Message) ToSend() []byte {
	b, err := json.Marshal(*m)
	if err != nil {
		return nil
	}

	return b
}
