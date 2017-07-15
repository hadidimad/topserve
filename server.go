package topserve

import (
	"encoding/json"
	"fmt"
	"net"
)

type Publisher struct {
	Name        string
	Conn        net.Conn
	Subscribers map[*Subscriber]struct{}
}

type Subscriber struct {
	Conn net.Conn
}

type Message struct {
	Mode string
	Data interface{}
}

type Server struct {
	Listener   net.Listener
	Publishers map[*Publisher]struct{}
}

func (p *Publisher) RegisterSubscriber(c net.Conn) error {
	for i := range p.Subscribers {
		if c == i.Conn {
			return fmt.Errorf("you are a subscriber for" + p.Name)
		}
	}
	sub := &Subscriber{
		Conn: c,
	}
	p.Subscribers[sub] = struct{}{}
	return nil
}
func (p *Publisher) DeRegisterSubscriber(c net.Conn) error {
	for j := range p.Subscribers {
		if j.Conn == c {
			delete(p.Subscribers, j)
			return nil
		}
	}
	return fmt.Errorf("subscriber not found")
}

func (p *Publisher) Publish(v interface{}) error {
	for o := range p.Subscribers {
		var message Message
		message.Mode = "publish"
		m := make(map[string]interface{})
		m["name"] = p.Name
		m["value"] = v
		message.Data = m
		JSONval, err := json.Marshal(message)
		if err != nil {
			return err
		}
		_, err = o.Conn.Write(JSONval)
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}

func (s *Server) New(address string) (Server, error) {
	ln, err := net.Listen("tcp", address)
	if err != nil {
		return Server{}, err
	}
	return Server{
		Listener:   ln,
		Publishers: map[*Publisher]struct{}{},
	}, nil
}

func (s *Server) Publish(name string, v interface{}) error {
	for o := range s.Publishers {
		if o.Name == name {
			err := o.Publish(v)
			if err != nil {
				return err
			}
			return nil
		}
	}
	return fmt.Errorf("publisher not found")
}

func (s *Server) RegisterPublisher(name string, conn net.Conn) error {
	for i := range s.Publishers {
		if i.Name == name {
			return fmt.Errorf("this name Registered Before")
		}
	}

	pub := &Publisher{
		Name:        name,
		Conn:        conn,
		Subscribers: map[*Subscriber]struct{}{},
	}
	s.Publishers[pub] = struct{}{}
	return nil
}

func (s *Server) DeRegisterPublisher(name string, conn net.Conn) error {
	for i := range s.Publishers {
		if i.Name == name && i.Conn == conn {
			for i := range i.Subscribers {
				var msg Message
				msg.Mode = "delpublisher"
				msg.Data = name
				str, _ := json.Marshal(msg)
				_, err := i.Conn.Write(str)
				if err != nil {
					return err
				}
			}
			delete(s.Publishers, i)
			return nil
		}
	}

	return fmt.Errorf("publisher not found")
}

func (s *Server) RegisterSubscriber(name string, c net.Conn) error {
	for i := range s.Publishers {
		if i.Name == name {
			err := i.RegisterSubscriber(c)
			if err != nil {
				return err
			}
			return nil
		}
	}
	return fmt.Errorf("publisher not found")
}

func (s *Server) DeRegisterSubscriber(name string, c net.Conn) error {
	for i := range s.Publishers {
		if i.Name == name {
			err := i.DeRegisterSubscriber(c)
			if err != nil {
				return err
			}
			return nil
		}
	}
	return fmt.Errorf("publisher not found")
}

func (s *Server) AcceptConnection() (net.Conn, error) {
	c, err := s.Listener.Accept()
	if err != nil {
		var t net.Conn
		return t, err
	}
	return c, nil
}

func (s *Server) HandleConnection(c net.Conn) {
	for {
		msg := make([]byte, 256)
		readLen, err := c.Read(msg)
		if err != nil {
			fmt.Println(err)
			c.Close()
			break
		} else {
			var message Message
			var response Message
			response.Mode = "error"
			json.Unmarshal(msg[:readLen], &message)
			if message.Mode == "publisher" {
				err := s.RegisterPublisher(message.Data.(string), c)
				if err != nil {
					response.Data = err.Error()
				} else {
					response.Data = " "

				}
				b, _ := json.Marshal(response)
				_, err = c.Write(b)
				if err != nil {
					fmt.Println(err)
				}
			}
			if message.Mode == "subscriber" {
				err := s.RegisterSubscriber(message.Data.(string), c)
				if err != nil {
					response.Data = err.Error()
				} else {
					response.Data = " "
				}
				b, _ := json.Marshal(response)
				_, err = c.Write(b)
				if err != nil {
					fmt.Println(err)
				}
			}
			if message.Mode == "publish" {
				Data := message.Data.(map[string]interface{})
				fmt.Println(Data["name"], Data["value"])
				err := s.Publish(Data["name"].(string), Data["value"])
				if err != nil {
					response.Data = err.Error()
				} else {
					response.Data = " "
				}
				b, _ := json.Marshal(response)
				_, err = c.Write(b)
				if err != nil {
					fmt.Println(err)
				}
			}
			if message.Mode == "delsubscriber" {
				err := s.DeRegisterSubscriber(message.Data.(string), c)
				if err != nil {
					response.Data = err.Error()
				} else {
					response.Data = " "
				}
				b, _ := json.Marshal(response)
				_, err = c.Write(b)
				if err != nil {
					fmt.Println(err)
				}
			}
			if message.Mode == "delpublisher" {
				err := s.DeRegisterPublisher(message.Data.(string), c)
				if err != nil {
					response.Data = err.Error()
				} else {
					response.Data = " "
				}
				b, _ := json.Marshal(response)
				_, err = c.Write(b)
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}
}