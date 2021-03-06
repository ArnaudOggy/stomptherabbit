package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/CanalTP/stomptherabbit/internal/rabbitmq"
	"github.com/CanalTP/stomptherabbit/internal/webstomp"
)

func main() {
	c, err := config()
	if err != nil {
		log.Fatalf("failed to load configuration: %s", err)
	}

	m := rabbitmq.NewAmqpManager(c.RabbitMQ.URL, c.RabbitMQ.Exchange.Name, c.RabbitMQ.ContentType)
	defer m.Close()

	done := make(chan struct{})
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)

		<-c
		close(done)
	}()

	opts := webstomp.Options{
		Protocol:    c.Webstomp.Protocol,
		Login:       c.Webstomp.Login,
		Passcode:    c.Webstomp.Passcode,
		SendTimeout: c.Webstomp.SendTimeout,
		RecvTimeout: c.Webstomp.RecvTimeout,
		Target:      c.Webstomp.Target,
		Destination: c.Webstomp.Destination,
	}

	var wsClient *webstomp.Client
	go func() {
		wsClient = webstomp.NewClient(opts)
		wsClient.Consume(func(msg []byte) {
			m.Send(msg)
		})
	}()

	fmt.Println(c.ToString())
	fmt.Println("Waiting for messages...")
	<-done
	fmt.Println("Gracefully exiting...")
	wsClient.Disconnect()
}
