package topserve

import (
	"encoding/json"
	"fmt"
	"net"
	"time"
)

type handler func(v interface{}, unixTime int64)

//Client struct
type Client struct {
	c               net.Conn
	topic           map[string][]handler
	YourPublishers  map[string]struct{}
	YourSubscribers map[string]struct{}
	Err             chan string
}

//New will make new client and connect it to server in  address
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

//Subscribe will subscribe handle function to event
func (cl *Client) Subscribe(name string, h handler) error {
	for i := range cl.topic {
		if i == name {
			cl.topic[name] = append(cl.topic[name], h)
			return nil
		}
	}
	return fmt.Errorf("publisher not finded")
}

//RegisterSubscriber will register you as event subscriber in server to recive event
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

//DeRegisterSubscriber will remove you from the event subscribers so you will not recive event
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

//Publish should run when event happens this will run event handlers
func (cl *Client) Publish(name string, v interface{}) error {
	for i := range cl.topic {
		if i == name {
			handlers := cl.topic[name]
			for _, i := range handlers {
				i(v, time.Now().Unix())
			}
		}
	}
	return nil
}

//PublishServer will notify server that this event happend
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

//RegisterPublisher will Register you as publisher of event at server
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

//DeRegisterPublisher will remove you from server publisher for this event
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

func (cl *Client) End() {
	for i := range cl.YourPublishers {
		cl.DeRegisterPublisher(i)
	}
	for i := range cl.YourSubscribers {
		cl.DeRegisterSubscriber(i)
	}
	cl.c.Close()
}

//HandleIncomings will recive and process them this should be routine after making new client
func (cl *Client) HandleIncomings() {
	for {
		msg := make([]byte, 256)
		readLen, err := cl.c.Read(msg)
		if err != nil {
			fmt.Println(err)
			cl.c.Close()
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
						}
					}
				}
			}
		}
	}
}
