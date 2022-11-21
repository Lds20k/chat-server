package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:3000")
	fmt.Println("-- connected --")
	if err != nil {
		log.Fatal(err)
	}

	exit := false

	go func() {
		for !exit {
			io.Copy(os.Stdout, conn)
			fmt.Println("-- disconnected --")
			retry := true
			for !exit && retry {
				time.Sleep(time.Second)
				fmt.Println("-- retry connection --")
				conn, err = net.Dial("tcp", "localhost:3000")
				if err == nil {
					fmt.Println("-- reconnected --")
					retry = false
				}

			}
		}
		fmt.Println("-- exit --")
		conn.Close()
		os.Exit(0)
	}()

	for {
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')

		if text == "/exit\n" {
			exit = true
		}

		if _, err := conn.Write([]byte(text)); err != nil {
			log.Println(err)
			return
		}
	}
}
