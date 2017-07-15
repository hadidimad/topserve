package topserve

import (
	"fmt"
	"testing"
)

func Test(t *testing.T) {
	var server Server
	server, err := server.New(":7777")
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	go func() {
		for {
			c, err := server.AcceptConnection()
			if err != nil {
				fmt.Println(err)
			} else {
				go server.HandleConnection(c)
			}
		}
	}()
	var client1 Client
	client1, err = client1.New(":7777")
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	go client1.HandleIncomings()
	var client2 Client
	client2, err = client2.New(":7777")
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	go client2.HandleIncomings()
}
