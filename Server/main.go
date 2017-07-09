package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

type Publisher struct {
	name        string
	Conn        net.Conn
	Subscribers map[*Subscriber]struct{}
}

type Subscriber struct {
	Conn net.Conn
}

type Topic struct {
	Publishers map[*Publisher]struct{}
}

func (t *Topic) RegisterPublisher(name string, conn net.Conn) error {
	for i := range t.Publishers {
		if i.name == name {
			return fmt.Errorf("this name Registered Before")
		}
		if i.Conn == conn {
			return fmt.Errorf("you registered a Publisher before")
		}
	}
	pub := &Publisher{
		name:        name,
		Conn:        conn,
		Subscribers: map[*Subscriber]struct{}{},
	}
	t.Publishers[pub] = struct{}{}
	return nil
}

func handleIncoming(c net.Conn, topic *Topic) {
	for {
		msg := make([]byte, 128)
		_, err := c.Read(msg)
		if err != nil {
			fmt.Println(err)
			c.Close()
			break
		} else {
			if strings.Contains(string(msg), "Register") {
				///////register new Publisher part
				if strings.Contains(string(msg), "Publisher") {
					str := strings.Replace(string(msg), "Register", "", -1)
					str = strings.Replace(str, "Publisher", "", -1)
					for i := 0; i < len(str); i++ {
						if str[i] == 0 {
							str = str[0:i]
							break
						}
					}
					str = strings.Replace(str, " ", "", -1)
					str = str[:len(str)-1]
					err := topic.RegisterPublisher(str, c)
					if err != nil {
						c.Write([]byte(err.Error()))
						fmt.Println("failed to Register" + str + "as Publsiher")
					} else {
						fmt.Println(str + "registered as Publisher")
					}
				}
				//////register new subscriber
				if strings.Contains(string(msg), "to") {

				}
			} else if strings.Contains(string(msg), "Deregister") {

			}
		}
	}
}

func main() {
	ln, err := net.Listen("tcp", ":7777")
	if err != nil {
		fmt.Println("err listen to port")
		os.Exit(-1)
	}
	topic := Topic{
		Publishers: map[*Publisher]struct{}{},
	}
	for {
		c, err := ln.Accept()
		if err != nil {
			fmt.Println("error making connection")
		}
		go handleIncoming(c, &topic)
		for i := range topic.Publishers {
			fmt.Println(i.name, i.Conn)
		}
	}
}
