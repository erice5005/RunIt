package runit

import (
	"log"
	"testing"
)

func Test_NodeStatGet(t *testing.T) {
	tests := []struct {
		name        string
		input       interface{}
		expectedVal interface{}
	}{
		{
			name:        "gets int",
			input:       int(1),
			expectedVal: int(1),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns_t := &NodeStat{
				val: tt.input,
			}

			expect_v := ns_t.Get()

			if expect_v != tt.expectedVal {
				t.Fail()
			}

		})
	}
}

func Test_BasicServerClient(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		server *Node
		client *Node
	}{
		{
			name: "server 1 and client 1: hello world",
			server: &Node{
				Config: NodeConfig{
					ZType: 4,
					Id:    "Server",
				},
				inputHandler: func(n *Node, p Payload) {
					log.Printf("Server Recieved: %v\n", p)
				},
				outputHandler: func(n *Node, p Payload) {
					log.Printf("Server Sending: %v\n", p)
				},
			},
			client: &Node{
				Config: NodeConfig{
					ZType: 5,
					Id:    "Client",
				},
				inputHandler: func(n *Node, p Payload) {
					log.Printf("Client Recieved: %v\n", p)
				},
				outputHandler: func(n *Node, p Payload) {
					log.Printf("Client Sending: %v\n", p)
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s1 := NewNode(tt.server.Config)
			if s1 == nil {
				t.Fail()
			}
			c1 := NewNode(tt.client.Config)
			if c1 == nil {
				t.Fail()
			}
			s1.SetReceiveHandler(tt.server.GetReceiveHandler())
			s1.SetSendHandler(tt.server.GetSendHandler())

			c1.SetReceiveHandler(tt.client.GetReceiveHandler())
			c1.SetSendHandler(tt.client.GetSendHandler())

			// Todo: connect? Looks to be only coming from server

			s1.Run()
			c1.Run()

			for i := 0; i < 10; i++ {
				t.Log(i)
				s1.AddOutput(Payload{
					Target: "Client",
					T:      4,
					Data: map[string]interface{}{
						"val": i,
					},
				})
			}
		})
	}
}
