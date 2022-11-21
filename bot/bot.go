package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
)

// https://www.educative.io/answers/how-to-reverse-a-string-in-golang
func reverse(str string) (result string) {
	for _, v := range str {
		result = string(v) + result
	}
	return
}

func main() {
	conn, err := net.Dial("tcp", "localhost:3000")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected!")

	exit := make(chan os.Signal, 1)
	tmp := make([]byte, 1024)
	stop := false

	if _, err := conn.Write([]byte("/nick [bot](inverter)\n")); err != nil {
		log.Println(err)
		return
	}

	signal.Notify(exit, os.Interrupt)
	go func() {
		<-exit
		stop = true
		fmt.Println("[local] Bot exiting!")
		conn.Close()
		os.Exit(0)
	}()

	for {
		n, err := conn.Read(tmp)
		if err != nil {
			if err != io.EOF && !stop {
				fmt.Println("read error:", err)
			}
			break
		}

		buffer := strings.TrimSuffix(string(tmp[:n]), "\n")
		fmt.Println(buffer)
		user_text := strings.SplitN(buffer, ": ", 2)

		if len(user_text) >= 2 {
			if user_text[0] != "[bot](inverter)" {
				user := "{" + user_text[0] + "}"
				text := user_text[1]
				if _, err := conn.Write([]byte(user + " -> " + reverse(text) + "\n")); err != nil {
					log.Println(err)
					return
				}
			}
		}
	}
}
