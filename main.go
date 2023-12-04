package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"strings"
)

var flag, created bool
var addressPort string = ":3410"

func main() {

	res := bufio.NewReader(os.Stdin)
	var s string
	flag = true
	created = false

	m := make(map[string]func([]string))
	m["help"] = help
	m["quit"] = quit
	m["port"] = port
	m["create"] = create
	m["ping"] = ping
	for flag {

		fmt.Print("::> ")
		s, _ = res.ReadString('\n')

		s = strings.TrimSpace(s)
		args := strings.Split(s, " ")

		f, ok := m[args[0]]
		if ok {
			f(args)
		}

	}

	return
}

func help(args []string) {
	fmt.Print("Help arrived!\n")
}

func quit(args []string) {
	flag = false
	fmt.Print("Quitting!\n")
}

func port(args []string) {
	if len(args) < 2 {
		fmt.Println("Not enough arguments for port!")
	}
	if created {
		fmt.Println("Node already created")
		return
	}

	addressPort = ":" + args[1]
	fmt.Println(args[1])
}

func create(args []string) {
	if created {
		fmt.Println("Node already created")
		return
	}

	node := NewNode(getLocalAddress())

	rpc.Register(node)
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", addressPort)
	if err != nil {
		panic("Error starting RPC")
	}

	go http.Serve(l, nil)
	created = true
	fmt.Println("Created node at address: " + getLocalAddress() + addressPort)

}

func call(address string, method string, args interface{}, reply interface{}) bool {
	c, err := rpc.DialHTTP("tcp", address)
	if err != nil {
		fmt.Println("Could not connect, return false")
		return false
	}
	defer c.Close()

	err = c.Call(method, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false

}

func join(args []string) {

}

func put(args []string) {

}

func putrandom(args []string) {

}

func get(args []string) {

}

func delete(args []string) {

}

func dump(args []string) {

}

func ping(args []string) {

	reply := Reply{}
	arguments := Args{"Ping"}

	ok := call(args[1], "Node.HandlePing", &arguments, &reply)
	if ok {
		fmt.Printf("reply: %v\n", reply.Reply)
	} else {
		fmt.Printf("Call failed\n")
	}
	r := reply.Reply
	fmt.Println(r)
}

func getLocalAddress() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	fmt.Println(conn.LocalAddr().(*net.UDPAddr).IP.String())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}
