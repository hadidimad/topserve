package topserve

import (
	"encoding/json"
	"fmt"
	"net"
)

//Publisher publishers use for notify subscibers about event
type Publisher struct {
	Name        string
	Conn        net.Conn
	Subscribers map[*Subscriber]struct{}
}

//Subscriber will give the event
type Subscriber struct {
	Conn net.Conn
}

//Message use to commiunicate between server and client
type Message struct {
	Mode string
	Data interface{}
}

//Server struct
type Server struct {
	Listener   net.Listener
	Publishers map[*Publisher]struct{}
}

//RegisterSubscriber will add client to this pubilsher subscribers by giving client connection
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

//DeRegisterSubscriber will remove c client from publisher subscribers
func (p *Publisher) DeRegisterSubscriber(c net.Conn) error {
	for j := range p.Subscribers {
		if j.Conn == c {
			delete(p.Subscribers, j)
			return nil
		}
	}
	return fmt.Errorf("subscriber not found")
}

//Publish will notify subscriber that event happens
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

//New will make new server
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

//Publish will notify publisher that event happend
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

//RegisterPublisher will add new publisher to server
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

//DeRegisterPublisher will remove this publisher from server
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

//RegisterSubscriber will add new Subscriber for a publisher in server
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

//DeRegisterSubscriber will remove subscriber from this publisher in server
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

//AcceptConnection will accpet new connections to clients
func (s *Server) AcceptConnection() (net.Conn, error) {
	c, err := s.Listener.Accept()
	if err != nil {
		var t net.Conn
		return t, err
	}
	return c, nil
}

func (s *Server) End() {
	for i := range s.Publishers {
		for j := range i.Subscribers {
			i.DeRegisterSubscriber(j.Conn)
		}
		s.DeRegisterPublisher(i.Name, i.Conn)
		i.Conn.Close()
	}

}

//HandleConnection will recive messages from client and procces them this should be routine after accepted new connection
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
