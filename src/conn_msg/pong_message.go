package conn_msg

import "touch_fish_missile/src/connect"

type PongMessage struct {
	Content MessageContent
}

func (msg *PongMessage) ServerHandleMessage(conn *connect.Connection) error {
	conn.ResetRetryTimes()
	return nil
}

func (msg *PongMessage) ClientHandleMessage(conn *connect.Connection) error {
	return nil
}

func NewPongMessage() PongMessage {
	return PongMessage{
		Content: MessageContent{MessageType: "PONG"},
	}
}
