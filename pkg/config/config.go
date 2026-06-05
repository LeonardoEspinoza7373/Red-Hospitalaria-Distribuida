package config

import "time"

const Port = "5000"

const ProxyAddr = "192.168.1.14:" + Port

const FrontendPort = "8080"

var IPToID = map[string]int{
	"192.168.1.10": 4,
	"192.168.1.11": 3,
	"192.168.1.12": 2,
	"192.168.1.13": 1,
}

var IDToIP = map[int]string{
	4: "192.168.1.10",
	3: "192.168.1.11",
	2: "192.168.1.12",
	1: "192.168.1.13",
}

var (
	HeartbeatInterval = 5 * time.Second
	HeartbeatTimeout  = 15 * time.Second
	ElectionTimeout   = 3 * time.Second
	StartupDelay      = 200 * time.Millisecond
)
