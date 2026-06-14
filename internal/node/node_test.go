package node

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/LeonardoEspinoza7373/Red-Hospitalaria-Distribuida/pkg/config"
)

var testIPs = map[int]string{
	4: "127.0.0.10",
	3: "127.0.0.11",
	2: "127.0.0.12",
	1: "127.0.0.13",
}

func initTestConfig() {
	config.IPToID = make(map[string]int)
	config.IDToIP = make(map[int]string)
	for id, ip := range testIPs {
		config.IPToID[ip] = id
		config.IDToIP[id] = ip
	}
}

type nodeConfig struct {
	id   int
	ip   string
	port int
}

func buildPeers(nodes []nodeConfig, selfID int) map[int]string {
	peers := make(map[int]string)
	for _, n := range nodes {
		if n.id != selfID {
			peers[n.id] = net.JoinHostPort(n.ip, strconv.Itoa(n.port))
		}
	}
	return peers
}

func testNode(t *testing.T, allNodes []nodeConfig, selfID int) *Node {
	t.Helper()
	var self nodeConfig
	for _, n := range allNodes {
		if n.id == selfID {
			self = n
			break
		}
	}

	n := NewWithPortAndPeers(self.ip, strconv.Itoa(self.port), buildPeers(allNodes, selfID))
	if n == nil {
		t.Fatalf("failed to create node id=%d", selfID)
	}
	n.Start()
	return n
}

func waitForCoordinator(t *testing.T, nodes []*Node, timeout time.Duration) (coordinatorID int, ok bool) {
	t.Helper()
	deadline := time.After(timeout)
	for {
		select {
		case <-deadline:
			return 0, false
		default:
			var highestID int
			allMatch := true
			first := -1

			for _, n := range nodes {
				n.mu.Lock()
				if n.State == Coordinator && n.ID > highestID {
					highestID = n.ID
				}
				if first == -1 {
					first = n.CoordinatorID
				} else if n.CoordinatorID != first {
					allMatch = false
				}
				n.mu.Unlock()
			}

			if highestID > 0 && allMatch && first == highestID {
				return first, true
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func waitForCoordinatorID(t *testing.T, nodes []*Node, expectedID int, timeout time.Duration) bool {
	t.Helper()
	deadline := time.After(timeout)
	for {
		select {
		case <-deadline:
			return false
		default:
			allMatch := true
			for _, n := range nodes {
				n.mu.Lock()
				if n.CoordinatorID != expectedID {
					allMatch = false
				}
				n.mu.Unlock()
				if !allMatch {
					break
				}
			}
			if allMatch {
				return true
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func stopNodes(nodes []*Node) {
	for _, n := range nodes {
		n.Stop()
	}
}

func TestBullyElection_AllStartSimultaneously(t *testing.T) {
	initTestConfig()
	basePort := 10010

	allCfgs := []nodeConfig{
		{id: 4, ip: testIPs[4], port: basePort + 4},
		{id: 3, ip: testIPs[3], port: basePort + 3},
		{id: 2, ip: testIPs[2], port: basePort + 2},
		{id: 1, ip: testIPs[1], port: basePort + 1},
	}

	nodes := make([]*Node, 0, 4)
	for _, cfg := range allCfgs {
		nodes = append(nodes, testNode(t, allCfgs, cfg.id))
	}
	defer stopNodes(nodes)

	coordID, ok := waitForCoordinator(t, nodes, 20*time.Second)
	if !ok {
		t.Fatal("no coordinator elected within timeout")
	}

	t.Logf("coordinator elected: id=%d", coordID)
	if coordID != 4 {
		t.Errorf("expected coordinator id 4 (highest priority), got %d", coordID)
	}

	allAgree := waitForCoordinatorID(t, nodes, coordID, 10*time.Second)
	if !allAgree {
		t.Error("not all nodes agree on coordinator")
	}
}

func TestBullyElection_CoordinatorFails(t *testing.T) {
	initTestConfig()
	basePort := 10020

	allCfgs := []nodeConfig{
		{id: 4, ip: testIPs[4], port: basePort + 4},
		{id: 3, ip: testIPs[3], port: basePort + 3},
		{id: 2, ip: testIPs[2], port: basePort + 2},
		{id: 1, ip: testIPs[1], port: basePort + 1},
	}

	nodes := make([]*Node, 0, 4)
	for _, cfg := range allCfgs {
		nodes = append(nodes, testNode(t, allCfgs, cfg.id))
	}
	defer stopNodes(nodes)

	coordID, ok := waitForCoordinator(t, nodes, 20*time.Second)
	if !ok {
		t.Fatal("no coordinator elected within timeout")
	}
	t.Logf("initial coordinator: id=%d", coordID)

	if coordID != 4 {
		t.Skip("initial coordinator not id=4, skipping failover test")
	}

	nodes[0].Stop()
	t.Log("coordinator (id=4) stopped")

	survivors := nodes[1:]
	newCoordID, ok := waitForCoordinator(t, survivors, 25*time.Second)
	if !ok {
		t.Fatal("no new coordinator elected after failover within timeout")
	}
	t.Logf("new coordinator after failover: id=%d", newCoordID)

	if newCoordID != 3 {
		t.Errorf("expected new coordinator id 3 (highest among survivors), got %d", newCoordID)
	}
}

func TestBullyElection_HigherNodeJoinsLate(t *testing.T) {
	initTestConfig()

	basePort := 10030
	allCfgs := []nodeConfig{
		{id: 4, ip: testIPs[4], port: basePort + 4},
		{id: 3, ip: testIPs[3], port: basePort + 3},
		{id: 2, ip: testIPs[2], port: basePort + 2},
		{id: 1, ip: testIPs[1], port: basePort + 1},
	}

	n1 := testNode(t, allCfgs, 1)
	defer n1.Stop()

	time.Sleep(500 * time.Millisecond)
	n2 := testNode(t, allCfgs, 2)
	defer n2.Stop()

	time.Sleep(500 * time.Millisecond)
	n3 := testNode(t, allCfgs, 3)
	defer n3.Stop()

	nodes := []*Node{n1, n2, n3}

	coordID, ok := waitForCoordinator(t, nodes, 20*time.Second)
	if !ok {
		t.Fatal("no coordinator elected within timeout")
	}
	t.Logf("coordinator after sequential startup: id=%d", coordID)
	if coordID != 3 {
		t.Errorf("expected coordinator id 3 (highest among running), got %d", coordID)
	}

	n4 := testNode(t, allCfgs, 4)
	defer n4.Stop()
	nodes = append(nodes, n4)

	allAgree := waitForCoordinatorID(t, nodes, 4, 20*time.Second)
	if !allAgree {
		t.Error("higher-ID node (4) did not take over as coordinator")
	}
	t.Log("node 4 correctly took over as coordinator")
}

func TestBullyElection_DisabledDoesNotPromoteOnTimeout(t *testing.T) {
	initTestConfig()

	n := NewWithPortAndPeers(testIPs[1], "10040", map[int]string{})
	if n == nil {
		t.Fatal("failed to create node")
	}

	n.mu.Lock()
	n.State = Candidate
	n.gotOK = true
	n.BullyEnabled = false
	n.mu.Unlock()

	n.handleElectionTimeout()

	n.mu.Lock()
	defer n.mu.Unlock()

	if n.State != Follower {
		t.Fatalf("expected follower state when bully is disabled, got %v", n.State)
	}
	if n.CoordinatorID != 0 {
		t.Fatalf("expected coordinator id to remain unchanged, got %d", n.CoordinatorID)
	}
}

func TestMain(m *testing.M) {
	config.HeartbeatInterval = 1 * time.Second
	config.HeartbeatTimeout = 3 * time.Second
	config.ElectionTimeoutBase = 300 * time.Millisecond
	config.ElectionTimeoutJitter = 200 * time.Millisecond
	config.DiscoveryTimeoutBase = 100 * time.Millisecond
	config.DiscoveryTimeoutJitter = 100 * time.Millisecond
	config.HeartbeatJitter = 0.1
	config.StartupDelayMax = 50 * time.Millisecond

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn})))

	code := m.Run()
	os.Exit(code)
}

func init() {
	fmt.Fprintf(os.Stderr, "node tests using fast timings\n")
}
