package main

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var running, created bool
var localPort string
var node *Node
var add NodeAddress

var address *string
var addressPort *int
var joinAddress *string
var joinPort *int
var timeStablize *int
var timeFixFingers *int
var timeCheckPredecessor *int
var successorAmount *int
var identifier *string

var FingerTableSize = 10

var m sync.Mutex
var connections = make(map[string]*rpc.Client)

var prime int64 = 479
var generator int64 = 13
var SK *int
var secretKey *string

func main() {

	address = flag.String("a", "", "Address")
	addressPort = flag.Int("p", -1, "Port")
	joinAddress = flag.String("ja", "", "Join address")
	joinPort = flag.Int("jp", -1, "Join port")
	timeStablize = flag.Int("ts", -1, "Time before stabilize call")
	timeFixFingers = flag.Int("tff", -1, "Time before fix fingers call")
	timeCheckPredecessor = flag.Int("tcp", -1, "Time before check predecessor is called")
	successorAmount = flag.Int("r", -1, "The amount of immediate successor stored")
	identifier = flag.String("i", "", "The string identifier of a node")
	SK = flag.Int("sk", 0, "Key")
	secretKey = flag.String("e", "", "end key")

	flag.Parse()
	*address = strings.TrimSpace(*address)
	*joinAddress = strings.TrimSpace(*joinAddress)
	*identifier = strings.TrimSpace(*identifier)
	localPort = ":" + strconv.Itoa(*addressPort)

	if (len(*joinAddress) == 0 && *joinPort >= 0) || (len(*joinAddress) > 0 && *joinPort < 0) {
		fmt.Printf("You have to provide both -ja and -jp flags")
		return
	}
	if (*addressPort < 0 || *timeStablize < 1 || *timeCheckPredecessor < 1 || *timeFixFingers < 1) ||
		(*timeStablize > 60000 || *timeCheckPredecessor > 60000 || *timeFixFingers > 60000) {
		fmt.Println("Invalid arguments")
		return
	}

	add = NodeAddress(*address + localPort)
	node = &Node{Address: add, Successors: []NodeAddress{}, Predecessor: "", FingerTable: []NodeAddress{}, Bucket: make(map[Key]string)}

	server(*address, ":"+strconv.Itoa(*addressPort))

	//if not joining a ring

	//if joining a ring

	//fmt.Printf("%v %v %v %v %v %v %v %v %v ", *address, *addressPort, *joinAddress, *joinPort, *timeStablize, *timeFixFingers, *timeCheckPredecessor, *successorAmount, *identifier)

	if len(*joinAddress) > 0 && *joinPort > 0 {
		//Joining an existing ring
		add := NodeAddress(*joinAddress + ":" + strconv.Itoa(*joinPort))
		join(add)
	} else {
		//Creating a new ring
		args := []string{*address + ":" + strconv.Itoa(*addressPort)}
		create(args)

	}

	go loopCP(time.Duration(*timeCheckPredecessor))
	go loopStab(time.Duration(*timeStablize))
	go loopFF(time.Duration(*timeFixFingers))

	res := bufio.NewReader(os.Stdin)
	var s string
	running = true
	created = false

	m := make(map[string]func([]string))
	m["help"] = help
	m["quit"] = quit
	m["dump"] = dump
	m["StoreFile"] = StoreFile
	m["LookUp"] = LookUp
	m["PrintState"] = PrintState
	m["a"] = a
	for running {

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

func a(args []string) {
	aa := []string{}
	aa = append(aa, "StoreFile")
	aa = append(aa, "a")
	StoreFile(aa)
}

func LookUp(args []string) {
	add := findFile(args)
	fmt.Println(add)
}

func findFile(args []string) string {
	filename := args[1]

	fmt.Println("file id: ", hashString(filename))

	reply := Reply{}
	arguments := Args{"", filename, 0}

	add := node.Address
	flag := false

	for !flag {
		ok := call(string(add), "Node.FindSuccessor", &arguments, &reply)
		if !ok {
			fmt.Println("Failed to fix fingers")
			return ""
		}
		switch found := reply.Found; found {

		//if the file maps between self and successor then reply.Reply = node.Successor[0]
		case true:

			flag = true
			break
		//if the file maps somewhere else then we have to forward the request to a better node
		case false:
			add = NodeAddress(reply.Forward)
			fmt.Println("next id to check :", hashString(reply.Forward))
			break
		}

	}

	//Print out the correct address to store the file in. Dependent on hashString(filename)
	//fmt.Println(reply.Reply)
	return reply.Reply

}

func StoreFile(args []string) {
	filename := args[1]
	dest := findFile(args)

	file, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println("file open error: " + err.Error())
	}

	content := string(file)

	enc, err := EncryptMessage([]byte(*secretKey), content)
	if err != nil {
		fmt.Println("Error encrypting: " + err.Error())

	}

	reply := FReply{}
	arguments := FArgs{string(add), filename, enc}

	ok := call(string(dest), "Node.Store", &arguments, &reply)
	if !ok {
		fmt.Println("cano reach the node")
		return
	}

}

func HandleKeys(filename string, dest string) string {
	PK := int64(math.Mod(math.Pow(float64(generator), float64(*SK)), float64(prime)))

	args := KArgs{filename, PK, prime, generator, 0}
	reply := KReply{}

	ok := call(dest, "Node.HandleRequest", &args, &reply)
	if !ok {
		fmt.Println("Error requesting")
		return ""
	}

	secret := int64(math.Mod(math.Pow(float64(reply.PublicKey), float64(*SK)), float64(prime)))

	secretExt := strconv.FormatInt(secret, 10)
	for len(secretExt) < 32 {
		secretExt = secretExt + secretExt
	}
	secretExt = secretExt[:32]

	ok = call(dest, "Node.GetKey", &args, &reply)
	if !ok {
		fmt.Println("Error requesting")
		return ""
	}
	fmt.Println(reply.EncKey)

	dKey, err := DecryptMessage([]byte(secretExt), reply.EncKey)
	if err != nil {
		fmt.Println("Error decrypting ", err)
		return ""
	}

	return dKey

}

func SendRequest(address string, filename string, PK int64, prime int64, generator int64, s int64) error {

	args := KArgs{filename, PK, prime, generator, s}
	reply := KReply{}
	fmt.Println("GENERATOR: ", generator)
	fmt.Println("PRIME: ", prime)

	ok := call(address, "Node.HandleRequest", &args, &reply)
	if !ok {
		fmt.Println("Error requesting")
		return nil
	}

	secret := int64(math.Mod(math.Pow(float64(reply.PublicKey), float64(*SK)), float64(prime)))
	fmt.Println("PK_A: ", reply.PublicKey)
	fmt.Println("PK_A: ", *SK)
	fmt.Println("PK_A: ", prime)
	fmt.Println("Secret: ", secret)

	secretExt := strconv.FormatInt(secret, 10)
	for len(secretExt) < 32 {
		secretExt = secretExt + secretExt
	}
	secretExt = secretExt[:32]

	args = KArgs{filename, PK, prime, generator, s}
	reply = KReply{}

	ok = call(address, "Node.GetKey", &args, &reply)
	if !ok {
		fmt.Println("Error requesting")
		return nil
	}
	fmt.Println(reply.EncKey)

	dKey, err := DecryptMessage([]byte(secretExt), reply.EncKey)
	if err != nil {
		fmt.Println("Error decrypting ", err)
		return nil
	}

	fmt.Println(dKey)

	args = KArgs{filename, PK, prime, generator, s}
	reply = KReply{}

	ok = call(address, "Node.GetFile", &args, &reply)
	if !ok {
		fmt.Println("Error requesting")
		return nil
	}

	text, err := DecryptMessage([]byte(dKey), reply.Content)
	if err != nil {
		fmt.Println("Error decrypting ", err)
		return nil

	}

	fmt.Println("decrypted?")
	fmt.Println(text)
	return nil

}

func EncryptMessage(key []byte, message string) (string, error) {
	byteMsg := []byte(message)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("could not create new cipher: %v", err)
	}

	cipherText := make([]byte, aes.BlockSize+len(byteMsg))
	iv := cipherText[:aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return "", fmt.Errorf("could not encrypt: %v", err)
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], byteMsg)

	return base64.StdEncoding.EncodeToString(cipherText), nil
}

func DecryptMessage(key []byte, message string) (string, error) {
	cipherText, err := base64.StdEncoding.DecodeString(message)
	if err != nil {
		return "", fmt.Errorf("could not base64 decode: %v", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("could not create new cipher: %v", err)
	}

	if len(cipherText) < aes.BlockSize {
		return "", fmt.Errorf("invalid ciphertext block size")
	}

	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(cipherText, cipherText)

	return string(cipherText), nil
}

func PrintState(args []string) {
	// dump(args)
	node.mu.Lock()
	fmt.Println(getNodeInfo(node.Address))

	fmt.Println("SUCCESSORS:")
	for _, s := range node.Successors {
		fmt.Println(getNodeInfo(s))
	}

	fmt.Println("FINGER TABLE:")
	for _, f := range node.FingerTable {
		fmt.Println(getNodeInfo(f))
	}

	node.mu.Unlock()
}

func getNodeInfo(address NodeAddress) string {
	id := hashAddress(address)
	return id.String() + " => " + string(address)

}

func loopCP(t time.Duration) {
	for {
		cp([]string{})
		time.Sleep(t * time.Millisecond)
	}

}

func loopFF(t time.Duration) {
	for {
		fix_fingers()
		time.Sleep(t * time.Millisecond)
	}

}

func loopStab(t time.Duration) {

	for {
		stabilize([]string{})
		time.Sleep(t * time.Millisecond)

	}

}

func help(args []string) {
	fmt.Print("Help arrived!\n")
}

func quit(args []string) {
	running = false
	m.Lock()
	defer m.Unlock()
	fmt.Println(len(connections))
	for add, conn := range connections {
		err := conn.Close()
		if err != nil {
			fmt.Println("error closing :", add, err)
		}
	}

	fmt.Print("Quitting!\n")
}

func cp(args []string) {

	arguments := Args{"CP", string(node.Address), 0}
	reply := Reply{}

	if string(node.Predecessor) == "" {
		return
	}

	ok := call(string(node.Predecessor), "Node.HandlePing", &arguments, &reply)
	if !ok {
		node.mu.Lock()
		fmt.Println("Can not connect to predecessor")
		node.Predecessor = NodeAddress("")
		node.mu.Unlock()
		return
	}

}

func fix_fingers() {
	if len(node.FingerTable) == 0 {
		node.mu.Lock()
		node.FingerTable = []NodeAddress{node.Successors[0]}
		node.mu.Unlock()

		return
	}

	node.FingerTable = []NodeAddress{}
	for next := 1; next <= FingerTableSize; next++ {
		offset := int64(math.Pow(2, float64(next)-1))
		add := node.Address
		flag := false
		for !flag {

			reply := Reply{}
			args := Args{"", string(node.Address), offset}

			ok := call(string(add), "Node.FindSuccessor", &args, &reply)
			if !ok {
				fmt.Println("Failed to fix fingers : ")
				return
			}
			//fmt.Println(reply.Found)

			switch found := reply.Found; found {
			case true:
				node.mu.Lock()
				node.FingerTable = append(node.FingerTable, NodeAddress(reply.Reply))
				//fmt.Println("SUCC: " + reply.Reply)
				flag = true
				node.mu.Unlock()
				break
			case false:
				//fmt.Println("FORWARD: " + reply.Forward)
				add = NodeAddress(reply.Forward)
				break

			}

		}

	}

}

func stabilize(args []string) {

	arguments := Args{"", string(node.Address), 0}
	reply := Reply{}

	ok := call(string(node.Successors[0]), "Node.Get_predecessor", &arguments, &reply)
	if !ok {
		fmt.Println("Could not connect to successor")
		dump([]string{})
		node.mu.Lock()
		node.Successors = node.Successors[1:]
		node.mu.Unlock()
		return
	}
	node.mu.Lock()
	addH := hashAddress(node.Address)
	addressH := hashAddress(NodeAddress(reply.Reply))
	succH := hashAddress(node.Successors[0])

	if between(addH, addressH, succH, false) && reply.Reply != "" {
		node.Successors = []NodeAddress{NodeAddress(reply.Reply)}
	}

	node.mu.Unlock()
	arguments = Args{"", string(node.Address), 0}
	reply = Reply{}
	ok = call(string(node.Successors[0]), "Node.Get_successors", &arguments, &reply)
	if !ok {
		//fmt.Println("Call failed to successor in stabilize <2>")
	}
	node.mu.Lock()
	//fmt.Println(reply.Successors)
	node.Successors = []NodeAddress{node.Successors[0]}
	node.Successors = append(node.Successors, reply.Successors[:len(reply.Successors)]...)
	if len(node.Successors) > *successorAmount {
		node.Successors = node.Successors[:*successorAmount]
	}
	node.mu.Unlock()

	arguments = Args{"Stabilize", string(node.Address), 0}
	reply = Reply{}
	notify([]string{})

}

func notify(args []string) {

	arguments := Args{"Notify", string(node.Address), 0}
	reply := Reply{}

	ok := call(string(node.Successors[0]), "Node.Notify", &arguments, &reply)
	if !ok {
		//fmt.Println("Call failed in notify")
	}

}

func server(address string, port string) {

	rpc.Register(node)
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", port)
	if err != nil {
		panic("Error starting RPC")
	}

	go http.Serve(l, nil)
	fmt.Println("Created node at address: " + address + localPort)

}

func create(args []string) {
	if created {
		fmt.Println("Node already created")
		return
	}

	node.create()

}

func join(address NodeAddress) {

	reply := Reply{}
	args := Args{"", string(node.Address), 0}

	add := address

	flag := false

	for !flag {

		call(string(add), "Node.FindSuccessor", &args, &reply)
		//fmt.Println(reply.Reply)

		switch found := reply.Found; found {
		case true:
			node.join(NodeAddress(reply.Reply))
			//fmt.Println("True")
			flag = true
			break
		case false:
			add = NodeAddress(reply.Forward)
			//fmt.Println("False")
			break

		}

	}

}

func call(address string, method string, args interface{}, reply interface{}) bool {

	m.Lock()
	defer m.Unlock()

	c, ok := connections[address]
	if ok { // if already connection to address
		err := c.Call(method, args, reply)
		if err == nil {
			return true
		}

		fmt.Println("CALL ERROR: ", err)
		delete(connections, address)
		return false
	}

	//then := time.Now().Nanosecond()
	c, err := rpc.DialHTTP("tcp", address)
	if err != nil {
		//fmt.Println("Could not connect, return false")
		fmt.Println("error :", err)
		return false
	}
	connections[address] = c

	err = c.Call(method, args, reply)

	//now := time.Now().Nanosecond()
	//fmt.Println(now - then)
	if err == nil {
		return true
	}

	fmt.Println("CALL ERROR: ", err)
	return false

}

func put(args []string) {

}

func putrandom(args []string) {

}

func get(args []string) {

}

// func delete(args []string) {

// }

func dump(args []string) {

	fmt.Println("Address: " + node.Address)
	fmt.Println("ID: " + hashAddress(node.Address).String())
	fmt.Print("Finger table: ")
	fmt.Println(node.FingerTable)
	fmt.Println("Predecessor: " + node.Predecessor)
	fmt.Print("Successors: ")
	fmt.Println(node.Successors)
	fmt.Print("Bucket: ")
	fmt.Println(node.Bucket)

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
