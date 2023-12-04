package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

var flag bool

func main() {

	res := bufio.NewReader(os.Stdin)
	var s string
	flag = true

	m := make(map[string]func([]string))
	m["help"] = help
	m["quit"] = quit
	m["port"] = port
	for flag {

		fmt.Print("::> ")
		s, _ = res.ReadString('\n')

		s = strings.TrimSpace(s)
		args := strings.Split(s, " ")
		fmt.Println("This is your text: ", s)

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
	fmt.Println(args[1])
}

func create(args []string) {

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
