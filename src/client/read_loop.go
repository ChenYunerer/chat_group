package client

import (
	"bufio"
	"chat_group/src/config"
	"chat_group/src/conn_msg"
	"chat_group/src/connect"
	"chat_group/src/log"
	"time"
)

func readLoop(conn *connect.Connection) {
	conf := config.GetInstance()
	bytes := make([]byte, 1024)
	reader := bufio.NewReader(conn.Conn)
	for {
		if conf.ReadTimeout.Seconds() != 0 {
			deadline := time.Now().Add(conf.WriteTimeout)
			err := conn.Conn.SetReadDeadline(deadline)
			if err != nil {
				log.Error(err)
			}
		}
		n, err := reader.Read(bytes)
		if err != nil {
			log.Error(err)
			conn.AddRetryTimes()
			connRetryTimes := conn.GetRetryTimes()
			log.Info("Read conn ", conn.RemoteAddress, " retry times is ", connRetryTimes)
			if connRetryTimes >= conf.RetryTimes {
				return
			}
			continue
		}
		if n == 0 {
			log.Error("no data read from reader")
			return
		}
		_, err = conn.Buffer.Write(bytes[:n])
		if err != nil {
			log.Error(err)
			return
		}
		bytess, err := conn_msg.DecodeData(conn.Buffer)
		if err != nil {
			log.Error(err)
			return
		}
		messages, err := conn_msg.DecodeMessage(bytess)
		if err != nil {
			log.Error(err)
			continue
		}
		for _, message := range messages {
			log.Info("receive conn_msg from client ", message)
			err = message.HandleMessage(conn)
			if err != nil {
				log.Error(err)
			}
		}
	}
}
