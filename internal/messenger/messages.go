package messenger

import (
	"fmt"
)

type Message struct {
	Text string
	User string
	Time string
}

func (msg *Message) string() string {
	if msg.Time == "" {
		return msg.Text
	}
	return fmt.Sprintf("[%s][%s]:%s\n", msg.Time, msg.User, msg.Text)
}
