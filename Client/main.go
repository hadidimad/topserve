package main

import (
	//"bufio"
	"fmt"
	"net"
	"time"

	//"os"
	"encoding/json"
	"reflect"
)

type handler func(v interface{}, unixTime int64)

type Message struct {
	Mode string
	Data interface{}
}

type Client struct {
	c               net.Conn
	topic           map[string][]handler
	YourPublishers  map[string]struct{}
	YourSubscribers map[string]struct{}
	Err             chan string
}

func (cl *Client) New(address string) (Client, error) {
	c, err := net.Dial("tcp", address)
	if err != nil {
		return Client{}, err
	}
	return Client{
		c:               c,
		topic:           make(map[string][]handler),
		YourPublishers:  map[string]struct{}{},
		YourSubscribers: map[string]struct{}{},
		Err:             make(chan string),
	}, nil
}

func (cl *Client) Subscribe(name string, h handler) error {
	for i := range cl.topic {
		if i == name {
			cl.topic[name] = append(cl.topic[name], h)
			return nil
		}
	}
	return fmt.Errorf("publisher not finded")
}

func (cl *Client) RegisterSubscriber(name string) error {
	msg := Message{
		Mode: "subscriber",
		Data: name,
	}
	str, _ := json.Marshal(msg)
	_, err := cl.c.Write([]byte(str))
	if err != nil {
		return err
	}
	Serr := <-cl.Err
	if " " != Serr {
		return fmt.Errorf(Serr)
	}
	cl.YourSubscribers[name] = struct{}{}
	cl.topic[name] = []handler{}
	return nil
}

func (cl *Client) DeregisterSubscriber(name string) error {
	var finded bool
	for i := range cl.YourSubscribers {
		if i == name {
			finded = true
			break
		}
	}
	if !finded {
		return fmt.Errorf("you are not subscriber for %s", name)
	}
	msg := Message{
		Mode: "delsubscriber",
		Data: name,
	}
	str, _ := json.Marshal(msg)
	_, err := cl.c.Write([]byte(str))
	if err != nil {
		return err
	}
	Serr := <-cl.Err
	if " " != Serr {
		return fmt.Errorf(Serr)
	}
	delete(cl.YourPublishers, name)
	delete(cl.topic, name)
	return nil
}

func (cl *Client) UnSubscribe(name string, h handler) error {
	if cl.topic[name] == nil {
		return fmt.Errorf("this publisher not found")
	}
	handlers := cl.topic[name]
	address := reflect.ValueOf(h).Pointer()
	index := -1
	for i := range handlers {
		if reflect.ValueOf(handlers[i]).Pointer() == address {
			index = i
			break
		}
	}
	if index == -1 {
		return fmt.Errorf("handler not found")
	}
	handlers = append(handlers[:index], handlers[index+1:]...)
	cl.topic[name] = handlers
	return nil
}

func (cl *Client) Publish(name string, v interface{}) error {
	for i := range cl.topic {
		if i == name {
			handlers := cl.topic[name]
			for _, i := range handlers {
				go i(v, time.Now().Unix())
			}
		}
	}
	return nil
}

func (cl *Client) PublishServer(name string, v interface{}) error {
	for i := range cl.YourPublishers {
		if i == name {
			var message Message
			message.Mode = "publish"
			m := make(map[string]interface{})
			m["name"] = name
			m["value"] = v
			message.Data = m
			JSONval, err := json.Marshal(message)
			if err != nil {
				return err
			}
			_, err = cl.c.Write(JSONval)
			if err != nil {
				return err
			}
			Serr := <-cl.Err
			if " " != Serr {
				return fmt.Errorf(Serr)
			}
			return nil
		}
	}
	return fmt.Errorf("you have not registered this publisher")
}

func (cl *Client) RegisterPublisher(name string) error {
	msg := Message{
		Mode: "publisher",
		Data: name,
	}
	str, _ := json.Marshal(msg)
	_, err := cl.c.Write([]byte(str))
	if err != nil {
		return err
	}
	Serr := <-cl.Err
	if " " != Serr {
		return fmt.Errorf(Serr)
	}

	cl.YourPublishers[name] = struct{}{}
	return nil
}

func (cl *Client) HandleIncomings() {
	for {
		msg := make([]byte, 256)
		readLen, err := cl.c.Read(msg)
		if err != nil {
			fmt.Println(err)
		} else {
			var message Message
			err = json.Unmarshal(msg[:readLen], &message)
			if err != nil {
				fmt.Println(err)
			} else {
				if message.Mode == "error" {
					cl.Err <- message.Data.(string)
				}
				if message.Mode == "publish" {
					Data := message.Data.(map[string]interface{})
					err := cl.Publish(Data["name"].(string), Data["value"])
					if err != nil {
						fmt.Println(err)
					}
				}
			}
		}
	}
}

func main() {
	var client Client
	client, err := client.New(":7777")
	if err != nil {
		fmt.Println(err)
	}
	go client.HandleIncomings()
	for {
		var input string
		fmt.Scan(&input)
		if input == "rp" {
			err := client.RegisterPublisher("test:test")
			if err != nil {
				fmt.Println(err)
			}
		}
		if input == "p" {
			client.PublishServer("test:test", 23)
		}
		if input == "sub" {
			err := client.Subscribe("test:test", func(v interface{}, unixTime int64) {
				fmt.Println("test:test happens in ", unixTime, "and the value is ", v)
			})
			if err != nil {
				fmt.Println(err)
			}
		}
		if input == "rsub" {
			err := client.RegisterSubscriber("test:test")
			if err != nil {
				fmt.Println(err)
			}
		}
		if input == "drsub" {
			err := client.DeregisterSubscriber("test:test")
			if err != nil {
				fmt.Println(err)
			}
		}
		if input == "usub" {
			err := client.UnSubscribe("test:test", func(v interface{}, unixTime int64) {
				fmt.Println("test:test happens in ", unixTime, "and the value is ", v)
			})
			if err != nil {
				fmt.Println(err)
			}
		}
		fmt.Println("task Done")
	}

}
