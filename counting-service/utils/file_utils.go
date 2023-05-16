package utils

import (
	"errors"
	"log"
	"os"
	"strconv"
	"strings"
)

func DirExists(dname string) bool {
	_, err := os.Stat(dname)
	return !errors.Is(err, os.ErrNotExist)
}

func FileExists(fname string) bool {
	info, err := os.Stat(fname)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func GetLastCountFromFile(filename string) int64 {
	// Open file
	file, err := os.Open(filename)
	Check(err)
	defer file.Close()

	// Create empty buffer
	buf := make([]byte, 32)
	stat, err := os.Stat(filename)
	Check(err)

	// Get file size and read last n bytes into buffer
	start := stat.Size() - 32
	_, err = file.ReadAt(buf, start)
	if err != nil {
		// File read error, e.g. there aren't 64 bytes available to read
		log.Println(err)
		return 0
	}

	// Convert buffer to string, trim outside whitespace, split into chunks
	s := strings.TrimSpace(string(buf))
	splits := strings.Split(s, " ")

	// The last chunk is the final recorded count
	lastCount, err := strconv.Atoi(splits[len(splits)-1])
	if err != nil {
		log.Println(err)
		return 0
	}

	return int64(lastCount)
}
