package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/gorilla/websocket"
	"github.com/moby/moby/client"
	"github.com/robert-nix/ansihtml"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"unicode"
)

var sessions []*websocket.Conn

func main() {
	name := os.Getenv("CONTAINER")
	if name == "" {
		fmt.Println("CONTAINER env-var is not set")
		os.Exit(1)
	}
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}

	l, err := cli.ContainerLogs(context.Background(), name, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true, Follow: true, Tail: "all"})
	if err != nil {
		panic(err)
	}

	line := make(chan string)

	go func() {
		reader := bufio.NewReader(l)
		for {
			readLine, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println(err)
			} else {
				line <- strings.TrimSpace(strings.TrimLeft(readLine, string([]byte{0x1, 0x0000, 0x00c})))
			}
		}
	}()

	go func() {
		for {
			msg := <-line
			msg = string(ansihtml.ConvertToHTML([]byte(msg))[1:])
			clean := strings.Map(func(r rune) rune {
				if unicode.IsGraphic(r) {
					return r
				}
				return -1
			}, msg)

			fmt.Printf("%q\n", clean)
			fmt.Println(len(clean))

			clean = strings.Map(func(r rune) rune {
				if unicode.IsPrint(r) {
					return r
				}
				return -1
			}, msg)
			for _, session := range sessions {
				fmt.Printf("message: %x ", []byte(msg))
				err := session.WriteMessage(websocket.TextMessage, []byte(clean))
				if err != nil {
					fmt.Println(err.Error())
				}
			}
		}
	}()

	htmlFile, err := os.Open("website.html")
	if err != nil {
		panic(err)
	}

	html, err := ioutil.ReadAll(htmlFile)
	if err != nil {
		panic(err)
	}
	h := string(html)

	htmlFile.Close()

	http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, h)
	}))

	http.Handle("/ws", http.HandlerFunc(handler))
	http.ListenAndServe(":8080", nil)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("New connection")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	sessions = append(sessions, conn)
	conn.SetCloseHandler(func(code int, text string) error {
		fmt.Println("Closed session")
		return nil
	})
}
