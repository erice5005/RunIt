package runit

import (
	"encoding/json"
	"reflect"
	"strconv"

	zmq "github.com/pebbe/zmq4"
)

type MessageType int64 // Defines the types of messages that can be sent

const (
	Id         MessageType = 0
	KeepAlive  MessageType = 1
	Connect    MessageType = 2
	Disconnect MessageType = 3
	Data       MessageType = 4
)

type zType zmq.Type // Wrap zmq4 types

const (
	Router       zType = 0 // Generally used as a server type
	Dealer       zType = 1 // Generally used as a client type
	Publisher    zType = 2
	Subscriber   zType = 3
	Debug_Router zType = 4
	Debug_Dealer zType = 5
)

type NodeType int64

const (
	Basic NodeType = 0
)

// type MessageFrame struct {
// 	T       MessageType `json:"t"`
// 	Payload []byte      `json:"payload"`
// }

type Payload struct {
	Target string
	T      MessageType
	Data   map[string]interface{}
}

func (p Payload) Encode() [][]byte {
	out := make([][]byte, 0)

	val := reflect.TypeOf(p)

	for i := 0; i < val.NumField(); i++ {
		marshed, _ := json.Marshal(val.Field(i))
		out = append(out, marshed)
	}

	return out
}

func DecodeToPayload(dataset []string) Payload {
	p := Payload{
		Target: dataset[0],
	}
	tNum, err := strconv.Atoi(dataset[2])
	handleErr(err, 0)
	p.T = MessageType(int64(tNum))

	var datamap map[string]interface{}
	err = json.Unmarshal([]byte(dataset[3]), &datamap)
	handleErr(err, 0)
	p.Data = datamap

	return p
}

func (p Payload) UnmarshalDatamap(target interface{}) interface{} {
	jsonString, err := json.Marshal(p.Data)
	handleErr(err, 0)
	json.Unmarshal(jsonString, &target)

	return target
}
