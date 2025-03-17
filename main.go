package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

const inputFilePath = "messages.txt"

func main() {
	f, err := os.Open(inputFilePath)
	if err != nil {
		log.Fatalf("could not open %s: %v\n", inputFilePath, err)
	}

	fmt.Printf("Reading data from %s\n", inputFilePath)
	fmt.Println("=====================================")

	lineChan := getLinesChannel(f)

	for line := range lineChan {
		fmt.Printf("read: %s\n", line)
	}
}

func getLinesChannel(f io.ReadCloser) <-chan string {
	lineChan := make(chan string)

	go func() {
		defer close(lineChan)
		defer f.Close()

		var currentLine string

		for {
			buf := make([]byte, 8, 8)
			n, err := f.Read(buf)
			if err != nil {
				if currentLine != "" {
					lineChan <- currentLine + string(buf[:n])
				}
				if errors.Is(err, io.EOF) {
					break
				}
				fmt.Printf("error: %v\n", err)
				break
			}
			str := string(buf[:n])

			parts := strings.Split(str, "\n")
			for i := range len(parts) - 1 {
				lineChan <- currentLine + parts[i]
				currentLine = ""
			}
			currentLine += parts[len(parts)-1]
		}
	}()

	return lineChan
}
