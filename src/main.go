package main

import (
	"bufio"
	"chat_group/src/config"
	"chat_group/src/conn_msg"
	"chat_group/src/connect"
	"chat_group/src/serialization"
	log "github.com/sirupsen/logrus"
	"net"
	"os"
	"os/signal"
	"sync"
	"time"
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})
}

func main() {
	go listenerConn()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	log.Info("Interrupt")
}

func listenerConn() {
	conf := config.GetInstance()
	listener, err := net.Listen(conf.Network, conf.GetAddress())
	if err != nil {
		log.Error(err)
	}
	log.Info("listen start")
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Error(err)
		}
		log.Info("accept a conn")
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	connection := connect.NewConnection(conn)
	connPool := connect.GetConnectionPoolInstant()
	connPool.AddConnection(connection)
	defer connection.Conn.Close()
	defer connPool.RemoveConnection(connection)
	log.Info("handle conn address is ", connection.RemoteAddress)
	quit := make(chan struct{})
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(quit)
		readLoop(connection)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		writeLoop(connection, quit)
	}()
	wg.Wait()
}

func readLoop(conn *connect.Connection) {
	bytes := make([]byte, 1024)
	for {
		reader := bufio.NewReader(conn.Conn)
		n, err := reader.Read(bytes)
		if err != nil {
			log.Error(err)
			return
		}
		if n == 0 {
			log.Error("no data read from reader")
			return
		}
		message, err := serialization.DecodeMessage(bytes)
		if err != nil {
			log.Info(err)
			continue
		}
		log.Info("receive conn_msg from client ", message)
		message.HandleMessage(*conn)
		//connect.GetConnectionPoolInstant().SendToOthers(*conn, bytes)
	}
}

func writeLoop(conn *connect.Connection, quit chan struct{}) {
	conf := config.GetInstance()
	pingTimer := time.NewTicker(conf.PingDuration)
	for {
		select {
		case messageBytes := <-conn.SendMessageChan:
			log.Info("send conn_msg to ", conn.RemoteAddress)
			n, err := conn.Conn.Write(messageBytes)
			if err != nil {
				log.Error(err)
				return
			}
			if n == 0 {
				log.Error("send data error")
				return
			}
		case <-pingTimer.C:
			pingMessage := conn_msg.NewPingMessage()
			bytes, err := serialization.EncodeMessage(&pingMessage)
			if err != nil {
				log.Error(err)
				continue
			}
			conn.SendMessageChan <- bytes
		case <-quit:
			return
		}
	}
}
