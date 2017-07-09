package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	c, err := net.Dial("tcp", ":7777")
	if err != nil {
		fmt.Println(err)
	}
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter text: ")
		text, _ := reader.ReadString('\n')
		fmt.Println(text)
		c.Write([]byte(text))
	}
}
