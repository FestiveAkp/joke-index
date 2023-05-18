package twitch

import (
	"log"
	"strings"
	"sync/atomic"

	irc "github.com/gempir/go-twitch-irc/v4"
)

func ListenToIRC(count *int64, channel string) {
	client := irc.NewAnonymousClient()

	client.OnConnect(func() {
		log.Println("Connected to chat.")
	})

	client.OnPrivateMessage(func(message irc.PrivateMessage) {
		// fmt.Println(message.Message)

		if strings.Contains(message.Message, "+2") {
			atomic.AddInt64(count, +2)
		} else if strings.Contains(message.Message, "-2") {
			atomic.AddInt64(count, -2)
		}
	})

	client.Join(channel)

	log.Fatal(client.Connect())
}
