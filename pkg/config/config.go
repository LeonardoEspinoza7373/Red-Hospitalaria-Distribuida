package config

import "time"

const Port = "5000"

const FrontendPort = "8080"

var IPToID = map[string]int{
	"192.168.1.10": 4,
	"192.168.1.11": 3,
	"192.168.1.12": 2,
	"192.168.1.13": 1,
	"192.168.1.14": 5,
}

var IDToIP = map[int]string{
	4: "192.168.1.10",
	3: "192.168.1.11",
	2: "192.168.1.12",
	1: "192.168.1.13",
	5: "192.168.1.14",
}

var (
	HeartbeatInterval    = 5 * time.Second
	HeartbeatTimeout     = 15 * time.Second
	ElectionTimeoutBase  = 2 * time.Second
	ElectionTimeoutJitter = 2 * time.Second
	DiscoveryTimeoutBase  = 1 * time.Second
	DiscoveryTimeoutJitter = 2 * time.Second
	HeartbeatJitter      = 0.2
	StartupDelayMax      = 3 * time.Second
	TimeSyncInterval     = 1 * time.Hour
)
