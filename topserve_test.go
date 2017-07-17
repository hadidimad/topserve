package topserve

import (
	"fmt"
	"testing"
	"time"
)

func TestPublishValues(t *testing.T) {
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
	err = client1.RegisterPublisher("test:test")
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	err = client2.RegisterSubscriber("test:test")
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	var num int = 0
	value := make(map[string]interface{})
	value["value1"] = 46
	value["value2"] = "string"
	err = client2.Subscribe("test:test", func(v interface{}, unixTime int64) {
		num++
		m := v.(map[string]interface{})
		fmt.Println(v)
		if m["value1"].(float64) != 46 {
			t.Log("value did not recive from client 1")
			t.Fail()
		}
		if m["value2"] != "string" {
			t.Log("value did not recive from client 1")
			t.Fail()
		}
	})
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	err = client1.PublishServer("test:test", value)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	time.Sleep(time.Millisecond * 10)
	if num == 0 {
		t.Log("subscriber didnt run")
		t.Fail()
	}
	err = client2.DeRegisterSubscriber("test:test")
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	err = client1.PublishServer("test:test", value)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	if num == 2 {
		t.Log("subscriber runs after delete")
		t.Fail()
	}
	err = client1.DeRegisterPublisher("test:test")
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	err = client2.RegisterSubscriber("test:test")
	if err == nil {
		t.Fail()
	}
	client1.End()
	client2.End()
}
