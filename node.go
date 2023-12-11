package main

import (
	"crypto/sha1"
	"fmt"
	"io"
	"math"
	"math/big"
	"os"
	"strconv"
	"sync"
)

type Key string
type NodeAddress string

type Node struct {
	mu          sync.Mutex
	Address     NodeAddress
	FingerTable []NodeAddress
	Predecessor NodeAddress
	Successors  []NodeAddress
	PK_A, PK_B  int64

	currentKey int64

	Bucket map[Key]string
}

func (n *Node) HandlePing(arguments *Args, reply *Reply) error {
	n.mu.Lock()
	//fmt.Println("In HandlePing")
	if arguments.Command == "CP" {
		reply.Reply = "CP reply"
	}
	n.mu.Unlock()
	return nil
}

func between(start *big.Int, elt *big.Int, end *big.Int, inclusive bool) bool {

	if end.Cmp(start) > 0 {
		return (start.Cmp(elt) < 0 && elt.Cmp(end) < 0) || (inclusive && elt.Cmp(end) == 0)
	} else {
		return start.Cmp(elt) < 0 || elt.Cmp(end) < 0 || (inclusive && elt.Cmp(end) == 0)
	}

}

func (n *Node) Get_predecessor(args *Args, reply *Reply) error {
	n.mu.Lock()
	reply.Reply = string(node.Predecessor)
	n.mu.Unlock()
	return nil

}

func (n *Node) GetNodeInfo(args *Args, reply *Reply) error {
	//  identifier, IP address, and port
	id := hashAddress(NodeAddress(args.Address))
	reply.Reply = id.String() + " " + args.Address
	return nil
}

func (n *Node) closest_predecing_node(id *big.Int) NodeAddress {

	for i := len(n.FingerTable) - 1; i >= 0; i-- {

		addH := hashAddress(n.Address)
		fingerH := hashAddress(n.FingerTable[i])

		if between(addH, fingerH, id, false) {
			return n.FingerTable[i]
		}

	}
	return n.Address

}

func hashAddress(elt NodeAddress) *big.Int {

	hasher := sha1.New()
	hasher.Write([]byte(elt))

	t := new(big.Int).SetBytes(hasher.Sum(nil))

	return new(big.Int).Mod(t, big.NewInt(int64(1024)))
	//return new(big.Int).SetBytes(hasher.Sum(nil))

}

func hashString(elt string) *big.Int {

	hasher := sha1.New()
	hasher.Write([]byte(elt))

	t := new(big.Int).SetBytes(hasher.Sum(nil))

	return new(big.Int).Mod(t, big.NewInt(int64(1024)))
	//return new(big.Int).SetBytes(hasher.Sum(nil))

}

func (n *Node) FindSuccessor(args *Args, reply *Reply) error {

	n.mu.Lock()
	addH := hashAddress(n.Address)

	ID := hashAddress(NodeAddress(args.Address))
	ID.Add(ID, big.NewInt(args.Offset))
	ID.Mod(ID, big.NewInt(int64(math.Pow(2, float64(FingerTableSize)))))

	succH := hashAddress(NodeAddress(n.Successors[0]))

	//If the ID is between self and immediate successor
	if between(addH, ID, succH, true) {
		reply.Found = true
		reply.Reply = string(n.Successors[0])
		//reply.Successors = n.Successors
	} else {
		//if the file is outside. Should return the closest preceding node before ID. Have to implement fix_fingers for this to work.
		//Right now it will return the next successor, jumping only 1 step on the ring. Search time is O(N), we want O(log(N))
		reply.Found = false
		reply.Reply = ""
		reply.Forward = string(n.closest_predecing_node(ID))

		//reply.Forward = string(n.Successors[0])
	}

	n.mu.Unlock()
	return nil

}

func (n *Node) Get_successors(args *Args, reply *Reply) error {

	n.mu.Lock()
	reply.Successors = node.Successors
	n.mu.Unlock()
	return nil

}

func (n *Node) create() {

	n.mu.Lock()
	n.Predecessor = ""
	n.Successors = append(n.Successors, n.Address)
	n.mu.Unlock()

}

func (n *Node) join(address NodeAddress) {
	n.mu.Lock()
	node.Predecessor = ""
	node.Successors = []NodeAddress{address}

	n.mu.Unlock()

}

func (n *Node) Notify(args *Args, reply *Reply) error {

	n.mu.Lock()
	addH := hashAddress(NodeAddress(args.Address))

	addressH := hashAddress(n.Address)

	preH := hashAddress(NodeAddress(n.Predecessor))

	if n.Predecessor == "" || (between(preH, addH, addressH, false)) {
		n.Predecessor = NodeAddress(args.Address)
		reply.Reply = "Success"
	} else {
		reply.Reply = "Fail"
	}
	n.mu.Unlock()
	return nil
}

func (n *Node) Store(args *FArgs, reply *FReply) error {
	fmt.Println("=================")
	content := args.Content
	org := args.Address
	filename := args.Filename + org + "ii"

	k := HandleKeys(filename, org)
	if k == "" {
		fmt.Println("key oopsie")
		return nil
	}

	dec, err := DecryptMessage([]byte(k), string(content))
	if err != nil {
		fmt.Println("errroror: ", err.Error())
	}

	err = os.WriteFile(filename, []byte(dec), 0777)
	if err != nil {
		fmt.Println("problem writing file")
	}

	// // secure files on other nodes
	// tmpFileame := "tmp_" + filename

	// // err = os.WriteFile(tmpFileame, []byte(content), 0777)
	// // if err != nil {
	// // 	fmt.Println("problem writing file")
	// // }

	// // err = os.WriteFile(tmpFileame, []byte(content), 0777)
	// var dests []NodeAddress

	// for _, a := range node.FingerTable {
	// 	if !slices.Contains(dests, a) {
	// 		dests = append(dests, a)
	// 	}
	// }

	// for _, a := range dests {
	// 	dest := a
	// 	reply := FReply{}
	// 	arguments := FArgs{string(add), tmpFileame, content}

	// 	ok := call(string(dest), "Node.Store", &arguments, &reply)
	// 	if !ok {
	// 		fmt.Println("cano reach the node")

	// 	}

	// }

	return nil
}

func (n *Node) HandleRequest(args *KArgs, reply *KReply) error {

	//filename := args.Filename
	// generator = args.Generator
	// prime = args.Prime

	n.PK_B = args.PublicKey
	// fmt.Println("PK_B: ", n.PK_B)

	ret := math.Mod(math.Pow(float64(generator), float64(*SK)), float64(prime))
	fmt.Println("PK: ", ret)

	reply.PublicKey = int64(ret)

	return nil
}

func (n *Node) GetKey(args *KArgs, reply *KReply) error {

	secret := int64(math.Mod(math.Pow(float64(n.PK_B), float64(*SK)), float64(args.Prime)))
	fmt.Println("Secret: ", secret)

	secretExt := strconv.FormatInt(secret, 10)
	for len(secretExt) < 32 {
		secretExt = secretExt + secretExt
	}
	secretExt = secretExt[:32]

	//fmt.Println(secretExt)

	eKey, err := EncryptMessage([]byte(secretExt), *secretKey)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	// dKey, err := DecryptMessage([]byte(secretExt), eKey)
	// if err != nil {
	// 	fmt.Println("Error decrypting ", err)
	// 	return nil
	// }

	// fmt.Println(eKey, dKey)

	reply.EncKey = eKey

	return nil
}

func (n *Node) GetFile(args *KArgs, reply *KReply) error {

	f, err := os.Open(args.Filename)
	if err != nil {
		return nil
	}

	content, err := io.ReadAll(f)
	if err != nil {
		return nil
	}

	reply.Content = string(content)
	return nil
}
