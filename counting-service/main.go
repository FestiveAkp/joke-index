// https://github.com/gempir/go-twitch-irc
//
// go run main.go
// tail -f data.log
// go run main.go | tee -a $(date -u +\%Y-\%m-\%d).log
// http "127.0.0.1:8080/events?stream={channel}"

package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/FestiveAkp/jji/counting-service/utils"
	"github.com/gempir/go-twitch-irc/v4"
	"github.com/r3labs/sse/v2"
)

var count int64
var channel = "moistcr1tikal"

func listenToChat() {
	client := twitch.NewAnonymousClient()

	client.OnConnect(func() {
		log.Println("Connected to chat.")
	})

	client.OnPrivateMessage(func(message twitch.PrivateMessage) {
		// fmt.Println(message.Message)

		if strings.Contains(message.Message, "LUL") || strings.Contains(message.Message, "KEKW") {
			atomic.AddInt64(&count, -1)
		} else {
			atomic.AddInt64(&count, 1)
		}
	})

	client.Join(channel)

	log.Fatal(client.Connect())
}

func startServer(server sse.Server) {
	port := ":8080"
	mux := http.NewServeMux()

	mux.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		server.ServeHTTP(w, r)
	})

	log.Println("Started web server on " + port)
	log.Fatal(http.ListenAndServe(port, mux))
}

func main() {
	log.Println("Starting counting-service...")

	// Create the data directory if it doesn't exist
	if !utils.DirExists("data/") {
		log.Println("Data directory not found, creating directory...")
		err := os.Mkdir("data/", os.ModePerm)
		utils.Check(err)
	}

	// This is the file that will be storing our stream of data
	dataFile := filepath.Join("data", channel+"-data.log")

	// Initialize the in-memory count with the last recorded value if it's available
	if utils.FileExists(dataFile) {
		lastCount := utils.GetLastCountFromFile(dataFile)
		log.Println("Found previous count of", lastCount)
		atomic.AddInt64(&count, lastCount)
	}

	// Start worker for listening to chat and updating counts
	go listenToChat()

	// Set up logger to append data to stdout and log file
	f, err := os.OpenFile(dataFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	utils.Check(err)
	defer f.Close()
	mw := io.MultiWriter(os.Stdout, f)
	fileLogger := log.New(mw, "", log.LstdFlags|log.LUTC)

	// Set up the server handler for pushing updates using Server-Sent Events
	server := sse.New()
	server.CreateStream(channel)

	// Run the web server in the background
	go startServer(*server)

	// Report the current count every second
	for range time.Tick(time.Second) {
		fileLogger.Println(count)

		now := strconv.FormatInt(time.Now().Unix(), 10)
		server.Publish(channel, &sse.Event{Data: []byte(now + " " + strconv.FormatInt(count, 10))})
	}
}
