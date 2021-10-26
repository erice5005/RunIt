package runit

import (
	"log"
	"sync"
	"time"

	zmq "github.com/pebbe/zmq4"
)

type NodeConfig struct { // Configuration info for a node
	ZType               zType    // zmq type, for the connection config
	NType               NodeType // the runit node type
	Url                 string
	TransmissionBuffers int
	Id                  string
}

type NodeStat struct {
	val interface{}
	mut sync.RWMutex
}

func (ns *NodeStat) Set(v interface{}) {
	ns.mut.Lock()
	defer ns.mut.Unlock()

	ns.val = v
}

func (ns *NodeStat) Get() interface{} {
	ns.mut.RLock()
	defer ns.mut.RUnlock()

	return ns.val
}

type Node struct { // Core unit, handles connecting and maintaining connections
	Config        NodeConfig
	socket        *zmq.Socket
	socketMut     sync.Mutex
	stats         map[string]*NodeStat //map of stats for a given node, each has its own mutex
	input         chan Payload
	output        chan Payload
	inputHandler  func(*Node, Payload)
	outputHandler func(*Node, Payload)
	connect       chan string
	disconnect    chan string
	connections   map[string]map[string]NodeStat
	connectionMut sync.RWMutex
	// connectHandler    func(string)
	// disconnectHandler func(string)
}

func NewNode(c NodeConfig) *Node {
	s, err := zmq.NewSocket(zmq.Type(c.ZType))
	if o := handleErr(err, 0); o == 1 {
		return nil
	}
	s.SetIdentity(c.Id)

	switch c.ZType {
	case 0:
		err = s.Bind(c.Url)
	case 1:
		err = s.Connect(c.Url)
	}

	if o := handleErr(err, 0); o == 1 {
		return nil
	}

	return &Node{
		socket: s,
		Config: c,
		stats:  map[string]*NodeStat{},
		input:  make(chan Payload, c.TransmissionBuffers),
		output: make(chan Payload, c.TransmissionBuffers),
	}
}

// func unmarshalFrame(data []string) []MessageFrame {
// 	o := make([]MessageFrame, 0)

// 	for _, d := range data {
// 		var f MessageFrame
// 		handleErr(json.Unmarshal([]byte(d), &f), 0)
// 		o = append(o, f)
// 	}

// 	return o
// }

func (n *Node) GetSocket() *zmq.Socket {
	n.socketMut.Lock()
	defer n.socketMut.Unlock()
	return n.socket
}

func (n *Node) Receive() {
	for {
		msg, err := n.GetSocket().RecvMessage(0)
		if o := handleErr(err, 0); o == 1 {
			return
		}

		n.input <- DecodeToPayload(msg)

	}
}

func (n *Node) SetReceiveHandler(handler func(*Node, Payload)) {
	n.inputHandler = handler
}

func (n *Node) GetReceiveHandler() func(*Node, Payload) {
	return n.inputHandler
}

func (n *Node) HandleReceive() {
	go n.Receive()
	for in := range n.input {
		n.inputHandler(n, in)
	}
}

func (n *Node) Send(payload [][]byte) {
	n.GetSocket().SendMessage(payload)
}

func (n *Node) SetSendHandler(handler func(*Node, Payload)) {
	n.outputHandler = handler
}
func (n *Node) GetSendHandler() func(*Node, Payload) {
	return n.outputHandler
}

func (n *Node) AddOutput(p Payload) {
	n.output <- p
}

func (n *Node) HandleSend() {
	for out := range n.output {
		n.Send(out.Encode())
		n.outputHandler(n, out)
	}

}

func (n *Node) HandleDisconnectStream() {
	for d := range n.disconnect {
		n.GetSocket().Close()
		n.connectionMut.Lock()
		delete(n.connections, d)
		n.connectionMut.Unlock()
	}
}

// func (n *Node) SetDisconnect(handler func(string)) {
// 	n.disconnectHandler = handler
// }

// func (n *Node) HandleDisconnect(val string) {
// 	n.disconnectHandler(val)
// }

func (n *Node) HandleConnectStream() {
	for c := range n.connect {
		n.GetSocket().Connect(c)
		n.connectionMut.Lock()
		n.connections[c] = NewConnectedStats()
		n.connectionMut.Unlock()
		n.AddOutput(Payload{
			Target: c,
			T:      Connect,
		})
	}
}

func NewConnectedStats() map[string]NodeStat {
	return map[string]NodeStat{
		"connected": {
			val: time.Now(),
		},
		"last_recieved": {
			val: time.Now(),
		},
	}
}

func (n *Node) Run() {
	go n.HandleReceive()
	go n.HandleConnectStream()
	go n.HandleDisconnectStream()
	go n.HandleSend()
	log.Printf("Running\n")
}

// func (n *Node) SetConnect(handler func(string)) {
// 	n.connectHandler = handler
// }

// func (n *Node) HandleConnect(val string) {
// 	n.connectHandler(val)
// }
