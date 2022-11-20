package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

type client chan<- string // canal de mensagem

type user struct {
	nick   string
	chanel chan string
}

var (
	entering = make(chan client)
	leaving  = make(chan client)
	messages = make(chan string)
	users    = make(map[string]user)
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

	user_obj := user{
		nick:   conn.RemoteAddr().String(),
		chanel: ch,
	}
	users[user_obj.nick] = user_obj
	fmt.Println(users)

	ch <- "vc é " + user_obj.nick
	messages <- user_obj.nick + " chegou!"
	entering <- ch

	input := bufio.NewScanner(conn)
main_loop:
	for input.Scan() {
		var text string = input.Text()
		if len(text) > 0 && text[0:1] == "/" {
			var command = strings.Split(text[1:], " ")[0]
			var arguments = strings.Split(text, " ")[1:]
			switch command {
			case "nick":
				if len(arguments) == 1 {
					if _, ok := users[arguments[0]]; !ok {
						var old_nick = user_obj.nick
						user_obj.nick = arguments[0]
						fmt.Println("\"" + old_nick + "\" change nick to \"" + user_obj.nick + "\"")

						delete(users, old_nick)
						users[user_obj.nick] = user_obj
					} else {
						user_obj.chanel <- "[server] Name already in use!"
					}
				} else {
					user_obj.chanel <- "[server] Invalid arguments!"
				}
			case "exit":
				if len(arguments) == 0 {
					fmt.Println("\"" + user_obj.nick + "\" exit from server")
					break main_loop
				} else {
					user_obj.chanel <- "[server] Invalid arguments!"
				}
			case "private":
				if len(arguments) >= 2 {
					fmt.Print("private to " + arguments[0] + ": ")
					if user_s, ok := users[arguments[0]]; ok {
						fmt.Println("-send-")
						user_s.chanel <- "[private] " + user_obj.nick + ": " + strings.Join(arguments[1:], " ")
						user_obj.chanel <- "[server] Message sended!"
					} else {
						fmt.Println("-not found-")
						user_obj.chanel <- "[server] User not found!"
					}
				} else {
					user_obj.chanel <- "[server] Invalid arguments!"
				}
			default:
				var error_message string = "command \"" + command + "\" not found"
				fmt.Println(error_message)
				messages <- error_message
			}

		} else {
			messages <- user_obj.nick + ": " + text
		}
	}

	leaving <- ch
	messages <- user_obj.nick + " saiu do servidor"
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
