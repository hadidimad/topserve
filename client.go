package topserve

import (
	"encoding/json"
	"fmt"
	"net"
	"time"
)

type handler func(v interface{}, unixTime int64)

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

func (cl *Client) DeRegisterSubscriber(name string) error {
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
	delete(cl.YourSubscribers, name)
	delete(cl.topic, name)
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

func (cl *Client) DeRegisterPublisher(name string) error {
	msg := Message{
		Mode: "delpublisher",
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
				if message.Mode == "delpublisher" {
					for i := range cl.YourSubscribers {
						if i == message.Data.(string) {
							delete(cl.YourSubscribers, message.Data.(string))
							delete(cl.topic, message.Data.(string))
							fmt.Println("publisher "+message.Data.(string), " deleted")
						}
					}
				}
			}
		}
	}
}
