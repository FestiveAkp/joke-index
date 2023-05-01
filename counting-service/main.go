// https://github.com/gempir/go-twitch-irc
//
// go run main.go
// tail -f data.log
// go run main.go | tee -a $(date -u +\%Y-\%m-\%d).log

package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gempir/go-twitch-irc/v4"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func fileExists(fname string) bool {
	info, err := os.Stat(fname)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func getLastCountFromFile(filename string) int64 {
	// Open file
	file, err := os.Open(filename)
	check(err)
	defer file.Close()

	// Create empty buffer
	buf := make([]byte, 32)
	stat, err := os.Stat(filename)
	check(err)

	// Get file size and read last n bytes into buffer
	start := stat.Size() - 32
	_, err = file.ReadAt(buf, start)
	if err != nil {
		// File read error, e.g. there aren't 64 bytes available to read
		fmt.Println(err)
		return 0
	}

	// Convert buffer to string, trim outside whitespace, split into chunks
	s := strings.TrimSpace(string(buf))
	splits := strings.Split(s, " ")

	// The last chunk is the final recorded count
	lastCount, err := strconv.Atoi(splits[len(splits)-1])
	if err != nil {
		fmt.Println(err)
		return 0
	}

	return int64(lastCount)
}

var count int64
var channel = "theprimeagen"

func listenToChat() {
	client := twitch.NewAnonymousClient()

	client.OnConnect(func() {
		fmt.Println("Connected to chat.")
	})

	client.OnPrivateMessage(func(message twitch.PrivateMessage) {
		// fmt.Println(message.Message)

		if strings.Contains(message.Message, "LUL") || strings.Contains(message.Message, "KEKW") {
			atomic.AddInt64(&count, -1)
		} else {
			// Safely increment the count
			atomic.AddInt64(&count, 1)
		}
	})

	client.Join(channel)

	err := client.Connect()
	check(err)
}

func main() {
	fmt.Println("Starting counting-service...")

	// Create the data directory if it doesn't exist
	if _, err := os.Stat("data/"); errors.Is(err, os.ErrNotExist) {
		fmt.Println("Data directory not found, creating directory...")
		err := os.Mkdir("data/", os.ModePerm)
		check(err)
	}

	// This is the file that will be receiving our stream of data
	dataFile := filepath.Join("data", channel+"-data.log")

	// Initialize the count with the last recorded value if it's available
	if fileExists(dataFile) {
		lastCount := getLastCountFromFile(dataFile)
		fmt.Println("Found previous count of", lastCount)

		atomic.AddInt64(&count, lastCount)
	}

	// Start worker for listening to chat and updating counts
	go listenToChat()

	// Set up logger to append data to stdout and log file
	f, err := os.OpenFile(dataFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	check(err)
	defer f.Close()
	mw := io.MultiWriter(os.Stdout, f)
	logger := log.New(mw, "", log.LstdFlags|log.LUTC)

	// Report the current count every second
	for range time.Tick(time.Second) {
		logger.Println(count)
	}
}
