package protocol

import (
	"encoding/json"
	"fmt"
	"time"
)

type MessageType string

type LogEntry struct {
	NodeID    int    `json:"node_id"`
	Level     string `json:"level"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

const (
	Heartbeat   MessageType = "HEARTBEAT"
	Election    MessageType = "ELECTION"
	OK          MessageType = "OK"
	Coordinator MessageType = "COORDINATOR"
	LogEvent    MessageType = "LOG_EVENT"
)

type Message struct {
	Type          MessageType `json:"type"`
	NodeID        int         `json:"node_id"`
	CoordinatorID int         `json:"coordinator_id,omitempty"`
	Timestamp     int64       `json:"timestamp"`
	LogData       *LogEntry   `json:"log_data,omitempty"`
}

func NewMessage(msgType MessageType, nodeID int) Message {
	return Message{
		Type:      msgType,
		NodeID:    nodeID,
		Timestamp: time.Now().Unix(),
	}
}

func NewCoordinatorMessage(nodeID int) Message {
	msg := NewMessage(Coordinator, nodeID)
	msg.CoordinatorID = nodeID
	return msg
}

func NewHeartbeatMessage(nodeID, coordID int) Message {
	msg := NewMessage(Heartbeat, nodeID)
	msg.CoordinatorID = coordID
	return msg
}

func (m Message) Encode() ([]byte, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("encode message: %w", err)
	}
	data = append(data, '\n')
	return data, nil
}

func Decode(data []byte) (Message, error) {
	var m Message
	if err := json.Unmarshal(data, &m); err != nil {
		return m, fmt.Errorf("decode message: %w", err)
	}
	return m, nil
}
