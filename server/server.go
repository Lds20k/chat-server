package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

type client chan<- string // canal de mensagem

var (
	entering = make(chan client)
	leaving  = make(chan client)
	messages = make(chan string)
)

func broadcaster() {
	clients := make(map[client]bool) // todos os clientes conectados
	for {
		select {
		case msg := <-messages:
			// broadcast de mensagens. Envio para todos
			for cli := range clients {
				cli <- msg
			}
		case cli := <-entering:
			clients[cli] = true
		case cli := <-leaving:
			delete(clients, cli)
			close(cli)
		}
	}
}

func clientWriter(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		fmt.Fprintln(conn, msg)
	}
}

func handleConn(conn net.Conn) {
	ch := make(chan string)
	go clientWriter(conn, ch)

	nick := conn.RemoteAddr().String()
	ch <- "vc é " + nick
	messages <- nick + " chegou!"
	entering <- ch

	input := bufio.NewScanner(conn)
	for input.Scan() {
		var text string = input.Text()
		if text[0:1] == "/" {
			var command = strings.Split(text[1:], " ")[0]
			var arguments = strings.Split(text, " ")[1:]
			switch command {
			case "nick":
				if len(arguments) == 1 {
					var old_nick = nick
					nick = arguments[0]
					fmt.Println("\"" + old_nick + "\" change nick to \"" + nick + "\"")
				}
			default:
				var error_message string = "command \"" + command + "\" not found"
				fmt.Println(error_message)
				messages <- error_message
			}

		} else {
			messages <- nick + ": " + text
		}
	}

	leaving <- ch
	messages <- nick + " se foi "
	conn.Close()
}

func main() {
	fmt.Println("Iniciando servidor...")
	listener, err := net.Listen("tcp", "localhost:3000")
	if err != nil {
		log.Fatal(err)
	}
	go broadcaster()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		go handleConn(conn)
	}
}
