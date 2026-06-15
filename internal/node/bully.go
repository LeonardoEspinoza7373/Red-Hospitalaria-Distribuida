package node

import (
	"math/rand"
	"net"
	"time"

	"github.com/LeonardoEspinoza7373/Red-Hospitalaria-Distribuida/internal/protocol"
	"github.com/LeonardoEspinoza7373/Red-Hospitalaria-Distribuida/pkg/config"
)

// All functions in this file are called exclusively from the event loop goroutine.
// They may acquire n.mu to read/write state, but never hold it across side effects.

func (n *Node) getHigherPeers() []string {
	var higher []string
	for peerID, addr := range n.Peers {
		if peerID > n.ID {
			higher = append(higher, addr)
		}
	}
	return higher
}

// ─── Discovery ─────────────────────────────────────────────────────────

func (n *Node) handleStartupDiscovery() {
	n.beginDiscovery()
}

func (n *Node) beginDiscovery() {
	n.mu.Lock()
	if !n.BullyEnabled {
		n.mu.Unlock()
		return
	}
	if n.State != Follower {
		n.mu.Unlock()
		return
	}
	n.State = Discovering
	n.mu.Unlock()

	n.log.Info("beginning coordinator discovery", "category", "bully")

	if len(n.Peers) == 0 {
		n.log.Info("no peers, skipping discovery", "category", "bully")
		n.mu.Lock()
		n.State = Follower
		n.mu.Unlock()
		n.beginElection()
		return
	}

	msg := protocol.NewMessage(protocol.CoordinatorQuery, n.ID)
	for _, addr := range n.Peers {
		peerAddr := addr
		go n.safeSend(peerAddr, msg)
	}

	timeout := randomTimeout(config.DiscoveryTimeoutBase, config.DiscoveryTimeoutJitter)
	timer := time.AfterFunc(timeout, func() {
		select {
		case n.events <- Event{Type: DiscoveryTimeout}:
		case <-n.ctx.Done():
		}
	})
	n.mu.Lock()
	n.discoveryTimer = timer
	n.mu.Unlock()
}

func (n *Node) handleDiscoveryTimeout() {
	n.mu.Lock()
	if n.State != Discovering {
		n.mu.Unlock()
		return
	}
	// If we learned about a coordinator via ACK in the meantime
	if n.CoordinatorID > 0 {
		n.State = Follower
		n.mu.Unlock()
		n.log.Info("discovery complete, coordinator already known", "coordinator_id", n.CoordinatorID, "category", "bully")
		return
	}
	n.State = Follower
	n.mu.Unlock()

	n.log.Info("discovery timeout — no coordinator found, starting election", "category", "bully")
	n.beginElection()
}

// ─── Coordinator Query / Ack ───────────────────────────────────────────

func (n *Node) handleCoordinatorQuery(fromID int, fromAddr string, msgEpoch int) {
	n.mu.Lock()
	knownCoord := n.CoordinatorID
	epoch := n.currentEpoch
	n.mu.Unlock()

	ack := protocol.NewMessage(protocol.CoordinatorAck, n.ID)
	ack.CoordinatorID = knownCoord
	ack.Epoch = epoch

	peerAddr, ok := n.Peers[fromID]
	if !ok {
		host, _, err := net.SplitHostPort(fromAddr)
		if err != nil {
			return
		}
		peerAddr = net.JoinHostPort(host, n.Port)
	}
	go n.safeSend(peerAddr, ack)
}

func (n *Node) handleCoordinatorAck(fromID, coordID, msgEpoch int) {
	n.mu.Lock()
	if n.State != Discovering {
		n.mu.Unlock()
		return
	}
	if coordID == 0 {
		n.mu.Unlock()
		return
	}
	// Never accept a coordinator with lower priority than ourselves
	if coordID < n.ID {
		n.mu.Unlock()
		n.log.Debug("ignoring coordinator ACK with lower-ID", "coord", coordID, "my_id", n.ID, "category", "bully")
		return
	}
	if msgEpoch < n.currentEpoch {
		n.mu.Unlock()
		return
	}
	if n.CoordinatorID > 0 && coordID <= n.CoordinatorID {
		n.mu.Unlock()
		return
	}
	if coordID > n.currentEpoch {
		n.currentEpoch = msgEpoch
	}
	n.CoordinatorID = coordID
	n.State = Follower
	n.gotOK = false
	n.lastHeartbeatAt = n.Now()
	if n.discoveryTimer != nil {
		n.discoveryTimer.Stop()
		n.discoveryTimer = nil
	}
	n.mu.Unlock()

	n.log.Info("discovered coordinator via query", "coordinator_id", coordID, "epoch", msgEpoch, "category", "bully")
	n.scheduleTimeSync()
}

// ─── Election ──────────────────────────────────────────────────────────

// beginElection starts a new election. Must be called from event loop.
func (n *Node) beginElection() {
	n.mu.Lock()
	if !n.BullyEnabled {
		n.mu.Unlock()
		n.log.Warn("election blocked: bully algorithm disabled", "category", "bully")
		return
	}
	if n.State == Coordinator || n.State == Candidate || n.State == Discovering {
		n.mu.Unlock()
		return
	}
	n.State = Candidate
	n.gotOK = false
	n.electionSeq++
	n.currentElectionID = (int64(n.currentEpoch) << 32) | int64(n.electionSeq)
	seq := n.electionSeq
	epoch := n.currentEpoch
	electionID := n.currentElectionID

	if n.electionTimer != nil {
		n.electionTimer.Stop()
		n.electionTimer = nil
	}
	n.mu.Unlock()

	n.log.Info("starting bully election", "seq", seq, "epoch", epoch, "election_id", electionID, "category", "bully")

	higherPeers := n.getHigherPeers()
	if len(higherPeers) == 0 {
		n.becomeCoordinator()
		return
	}

	msg := protocol.NewMessage(protocol.Election, n.ID)
	msg.Epoch = epoch
	msg.ElectionID = electionID

	for _, addr := range higherPeers {
		peerAddr := addr
		go n.safeSend(peerAddr, msg)
	}

	timeout := randomTimeout(config.ElectionTimeoutBase, config.ElectionTimeoutJitter)
	timer := time.AfterFunc(timeout, func() {
		select {
		case n.events <- Event{Type: ElectionResponseTimeout}:
		case <-n.ctx.Done():
		}
	})
	n.mu.Lock()
	n.electionTimer = timer
	n.mu.Unlock()
}

func (n *Node) handleElectionTimeout() {
	n.mu.Lock()
	if n.State != Candidate {
		n.mu.Unlock()
		return
	}
	gotOK := n.gotOK
	bullyEnabled := n.BullyEnabled
	n.mu.Unlock()

	if !bullyEnabled {
		n.log.Warn("election timed out but bully disabled — staying follower", "category", "bully")
		n.mu.Lock()
		if n.State == Candidate {
			n.State = Follower
		}
		n.gotOK = false
		n.mu.Unlock()
		return
	}

	if gotOK {
		n.log.Info("OK received during election, waiting for new coordinator", "category", "bully")
		n.mu.Lock()
		n.State = Follower
		n.lastHeartbeatAt = n.Now()
		n.mu.Unlock()
	} else {
		n.log.Info("no OK received, becoming coordinator", "category", "bully")
		n.becomeCoordinator()
	}
}

func (n *Node) becomeCoordinator() {
	n.mu.Lock()
	if n.State == Coordinator {
		n.mu.Unlock()
		return
	}
	if !n.BullyEnabled {
		n.State = Follower
		n.mu.Unlock()
		n.log.Warn("coordinator promotion blocked: bully algorithm disabled", "category", "bully")
		return
	}
	n.State = Coordinator
	n.CoordinatorID = n.ID
	n.currentEpoch++
	epoch := n.currentEpoch
	n.mu.Unlock()

	n.log.Info("became coordinator", "node_id", n.ID, "epoch", epoch, "category", "bully")

	msg := protocol.NewCoordinatorMessage(n.ID)
	msg.Epoch = epoch
	n.broadcast(msg)

	n.scheduleHeartbeat()
}

func (n *Node) handleStartupElection() {
	n.mu.Lock()
	hasCoord := n.CoordinatorID != 0
	lowerCoord := n.CoordinatorID > 0 && n.CoordinatorID < n.ID
	n.mu.Unlock()
	if !hasCoord {
		n.log.Info("no coordinator found, starting election", "category", "bully")
		n.beginElection()
	} else if lowerCoord {
		n.log.Info("current coordinator has lower priority, starting election", "category", "bully")
		n.beginElection()
	}
}

// ─── Message Handlers ──────────────────────────────────────────────────

func (n *Node) handleElectionMsg(fromID int, fromAddr string, msgEpoch int, msgElectionID int64) {
	n.mu.Lock()
	if msgEpoch > n.currentEpoch {
		n.currentEpoch = msgEpoch
	}
	myID := n.ID
	bullyEnabled := n.BullyEnabled
	myState := n.State
	hasActiveCoord := n.CoordinatorID > 0 && n.State == Follower
	currentEpoch := n.currentEpoch
	n.mu.Unlock()

	// Always reply OK with the election ID so the sender can match it
	peerAddr, ok := n.Peers[fromID]
	if !ok {
		host, _, err := net.SplitHostPort(fromAddr)
		if err != nil {
			return
		}
		peerAddr = net.JoinHostPort(host, n.Port)
	}

	okMsg := protocol.NewMessage(protocol.OK, myID)
	okMsg.Epoch = currentEpoch
	okMsg.ElectionID = msgElectionID
	go n.safeSend(peerAddr, okMsg)

	// If sender is lower-ID, we have no active coordinator, and we're not
	// already in an election or coordinator, start our own election
	if fromID < myID && bullyEnabled && myState != Coordinator && myState != Candidate && !hasActiveCoord {
		n.log.Debug("starting own election after receiving ELECTION from lower-ID node", "from", fromID, "category", "bully")
		n.beginElection()
	}
}

func (n *Node) handleOKMsg(fromID int, msgEpoch int, msgElectionID int64) {
	n.mu.Lock()
	candidate := n.State == Candidate
	epochMatch := msgEpoch == n.currentEpoch
	electionMatch := msgElectionID == n.currentElectionID
	if candidate && epochMatch && electionMatch {
		n.gotOK = true
	}
	n.mu.Unlock()

	if !candidate {
		return
	}
	if !epochMatch || !electionMatch {
		n.log.Debug("ignoring stale OK", "from", fromID, "msg_epoch", msgEpoch,
			"msg_election", msgElectionID, "current_election", n.currentElectionID, "category", "bully")
		return
	}

	n.log.Debug("received OK", "from", fromID, "category", "bully")
}

func (n *Node) acceptCoordinator(coordID, msgEpoch int) {
	n.mu.Lock()
	if msgEpoch > n.currentEpoch {
		n.currentEpoch = msgEpoch
	}
	n.CoordinatorID = coordID
	cancelling := n.State == Candidate || n.State == Discovering
	n.State = Follower
	n.gotOK = false
	n.lastHeartbeatAt = n.Now()
	if n.electionTimer != nil {
		n.electionTimer.Stop()
		n.electionTimer = nil
	}
	if n.discoveryTimer != nil {
		n.discoveryTimer.Stop()
		n.discoveryTimer = nil
	}
	n.mu.Unlock()

	if cancelling {
		n.log.Info("cancelling election, accepting coordinator", "coordinator_id", coordID, "epoch", msgEpoch, "category", "bully")
	}
	n.scheduleTimeSync()
}

func (n *Node) handleCoordinatorMsg(coordID int, msgEpoch int) {
	if coordID == n.ID {
		return
	}

	n.mu.Lock()
	currentCoord := n.CoordinatorID
	currentState := n.State
	currentEpoch := n.currentEpoch
	n.mu.Unlock()

	// Higher-ID node always wins — cancel any election and accept
	if coordID > n.ID {
		n.log.Info("accepting higher-ID coordinator", "coordinator_id", coordID, "epoch", msgEpoch, "category", "bully")
		n.acceptCoordinator(coordID, msgEpoch)
		return
	}

	// coordID < n.ID
	if msgEpoch > currentEpoch {
		n.mu.Lock()
		n.currentEpoch = msgEpoch
		n.mu.Unlock()
	}

	// If we already know a coordinator with >= ID, ignore
	if coordID <= currentCoord {
		return
	}

	// A lower-ID node claims to be coordinator and we don't have a better one
	if currentState == Coordinator {
		n.log.Warn("lower-ID node claims coordinator, asserting myself", "category", "bully")
		n.mu.Lock()
		epoch := n.currentEpoch
		n.mu.Unlock()
		msg := protocol.NewCoordinatorMessage(n.ID)
		msg.Epoch = epoch
		n.broadcast(msg)
	} else if currentState == Follower {
		n.log.Debug("lower-ID coordinator received, starting election", "from", coordID, "category", "bully")
		n.beginElection()
	}
	// If Candidate, our existing election will handle it
}

func (n *Node) handleHeartbeatMsg(fromID, coordID int, msgEpoch int) {
	n.mu.Lock()
	myID := n.ID
	currentState := n.State
	currentCoord := n.CoordinatorID
	currentEpoch := n.currentEpoch
	n.mu.Unlock()

	// Advance epoch if this heartbeat carries a higher one (cluster-wide term)
	if msgEpoch > currentEpoch {
		n.mu.Lock()
		n.currentEpoch = msgEpoch
		currentEpoch = msgEpoch
		n.mu.Unlock()
	}

	// If heartbeat claims a higher-ID coordinator than me, always accept
	if coordID > myID {
		n.log.Debug("accepting higher-ID coordinator via heartbeat", "coordinator_id", coordID, "epoch", msgEpoch, "category", "bully")
		n.acceptCoordinator(coordID, msgEpoch)
		return
	}

	// Heartbeat claiming I'm the coordinator — ignore (we'd know)
	if coordID == myID {
		return
	}

	// coordID < myID — lower-ID node claims to be coordinator
	if coordID <= currentCoord {
		return
	}

	// This lower-ID coordinator is higher than our current — contest or start election
	if currentState == Coordinator {
		n.log.Warn("lower-ID coordinator via heartbeat, asserting myself", "category", "bully")
		n.mu.Lock()
		epoch := n.currentEpoch
		n.mu.Unlock()
		msg := protocol.NewCoordinatorMessage(myID)
		msg.Epoch = epoch
		n.broadcast(msg)
	} else if currentState == Follower {
		n.log.Info("lower-ID coordinator via heartbeat, starting election", "category", "bully")
		n.beginElection()
	}
}

// ─── Heartbeat Tick (coordinator only) ─────────────────────────────────

func (n *Node) handleHeartbeatTick() {
	n.mu.Lock()
	if n.State != Coordinator {
		n.mu.Unlock()
		return
	}
	epoch := n.currentEpoch
	n.mu.Unlock()

	msg := protocol.NewHeartbeatMessage(n.ID, n.ID)
	msg.Epoch = epoch
	n.broadcast(msg)
	n.log.Debug("sent HEARTBEAT", "category", "heartbeat")

	n.scheduleHeartbeat()
}

func (n *Node) scheduleHeartbeat() {
	delay := config.HeartbeatInterval
	if config.HeartbeatJitter > 0 {
		jitter := time.Duration(float64(delay) * config.HeartbeatJitter)
		delay += time.Duration(rand.Int63n(int64(jitter*2))) - jitter
	}
	if delay < 100*time.Millisecond {
		delay = 100 * time.Millisecond
	}

	time.AfterFunc(delay, func() {
		select {
		case n.events <- Event{Type: HeartbeatTick}:
		case <-n.ctx.Done():
		}
	})
}
