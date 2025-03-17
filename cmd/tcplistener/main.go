package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

const port = ":42069"

func main() {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("error listening for TCP traffic: %s\n", err.Error())
	}
	defer listener.Close()

	fmt.Println("Listening for TCP traffic on", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("error: %s\n", err.Error())
		}
		fmt.Println("Accepted connection from", conn.RemoteAddr())

		lineChan := getLinesChannel(conn)

		for line := range lineChan {
			fmt.Println(line)
		}
		fmt.Println("Connection to ", conn.RemoteAddr(), "closed")
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
