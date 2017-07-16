# topserve
    golang topic server and client package
###what is topic?
topic is a something that use to connect some programs together with events
a topic have this parts:
    *Event:
        >event is thing that happens and it have some values
    *Publisher:
        >this will notify Subscribers that event happens and give them values
    *Subscriber:
        >this will listen to Publisher and when publisher notify that event happened it will give values and proccess them
for example imagine you have a program that will fetch data when a server sends it
so here we have a event (fetchdata) and event value (data that comes from server)
on the other side you have two programs one will show values to user
and the other one will process values
you need to connect this three programs together using topic
first all of the programs need to be client in topic server 
in this case the data fetcher program is a publisher for "fetched data" event
and other two programs are  subscribers now lets see how to do it with topserve

###Server Code

```golang
    var server topserve.Server
	server, err := server.New(":7777") //create a new topic server on port :7777 using tcp connection
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	for {
		c, err := server.AcceptConnection()  //accepting new connection to clients
		if err != nil {
			fmt.Println(err)
		} else {
			go server.HandleConnection(c)//handle and procces the client messages this need to be routine
		}
	}
```

this is server code now lets see client

###Client Code

first you should make new client for the port that server run on it to recive server messages
```golang
    var client topserve.Client
    client = client.New(":7777") // my server run at port :7777 so i connect client to that port
    go client.HandleIncomings()  //this will give messages from server and procces them
```
so we have connected now we need to register our client as "test:test" event publisher in server
```golang
    err = client.RegisterPublisher("test:test")
	if err != nil {
		t.Log(err)
		t.Fail()
	}
```
so now our first client is publisher for "test:test" event so now we need another client for our second program to subscribe
the "test:test" event
```golang
    var client2 topserve.Client
    client2 = client2.New(":7777") // my server run at port :7777 so i connect client2 to that port
    go client2.HandleIncomings()  //this will give messages from server and procces them

    err = client2.RegisterSubscriber("test:test")//register as subscriber of "test:test" event in server
	if err != nil {
		t.Log(err)
		t.Fail()
	}

    err = client2.Subscribe("test:test", func(v interface{}, unixTime int64) {
		fmt.Println("test:test happens in ",unixTime," and the value is",v)
	})
```
 now you will notify when "test:test" happens but you need to some jobs to do when event happens 
```golang
    err = client2.Subscribe("test:test", func(v interface{}, unixTime int64) {
		fmt.Println("test:test happens in ",unixTime," and the value is",v)
	})//this function will run when "test:test" happens
    err = client2.Subscribe("test:test", func(v interface{}, unixTime int64) {
		fmt.Println("test:test happens in ",unixTime," and the value is",v)
	})//this function will run when "test:test" happens
```
now return to first client now imagine that "test:test" happens you need to message server that event happens
to publish the event

```golang
    err = client.PublishServer("test:test", value)//this will notify server "test:test" happens 
	if err != nil {
		t.Log(err)
		t.Fail()
	}
```
now server will notify subscriber clients of "test:test" and they will run subscribed functions of event

now you are tired of subscribing "test:test" event and you dont want to recive event any more so you can deRegister your
subscriber from "test:test" event in server
```golang
    err = client2.DeRegisterSubscriber("test:test")// remove you from "test:test" subscribers you wil not notify any more
	if err != nil {
		t.Log(err)
		t.Fail()
	}
```

now you dont want to publish "test:test" any more so you can deregister your publisher from server and 
the subscriber will automaticly delete
```golang
    err = client.DeRegisterPublisher("test:test")
	if err != nil {
		t.Log(err)
		t.Fail()
	}
```

you can see good example at topserve_test.go

at the end please help me to imporve this library 
special thanks from @ahmd.rz


