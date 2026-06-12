package node

import (
	"net"
	"time"

	"github.com/LeonardoEspinoza7373/Red-Hospitalaria-Distribuida/internal/protocol"
	"github.com/LeonardoEspinoza7373/Red-Hospitalaria-Distribuida/internal/transport"
	"github.com/LeonardoEspinoza7373/Red-Hospitalaria-Distribuida/pkg/config"
)

func (n *Node) startElection() {
	n.mu.Lock()
	if !n.BullyEnabled {
		n.log.Warn("election blocked: bully algorithm disabled")
		n.mu.Unlock()
		return
	}

	if n.State == Coordinator || n.State == Candidate {
		n.mu.Unlock()
		return
	}

	n.State = Candidate
	n.gotOK = false
	n.electionSeq++
	seq := n.electionSeq
	n.mu.Unlock()

	n.log.Info("starting bully election", "seq", seq)

	higherPeers := n.getHigherPeers()

	if len(higherPeers) == 0 {
		n.becomeCoordinator()
		return
	}

	msg := protocol.NewMessage(protocol.Election, n.ID)
	for _, peerAddr := range higherPeers {
		addr := peerAddr
		go func() {
			if err := transport.SendMessage(addr, msg); err != nil {
				n.log.Warn("send ELECTION failed", "peer", addr, "error", err)
			}
		}()
	}

	time.AfterFunc(config.ElectionTimeout, func() {
		select {
		case n.events <- Event{Type: ElectionResponseTimeout}:
		case <-n.ctx.Done():
		}
	})
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
		n.log.Warn("election timed out but bully disabled - staying follower")
		n.mu.Lock()
		if n.State == Candidate {
			n.State = Follower
		}
		n.gotOK = false
		n.mu.Unlock()
		return
	}

	if gotOK {
		n.log.Info("OK received during election, waiting for new coordinator")
		n.scheduleCoordinatorWatch()
	} else {
		n.log.Info("no OK received, becoming coordinator")
		n.becomeCoordinator()
	}
}

func (n *Node) scheduleCoordinatorWatch() {
	n.mu.Lock()
	if n.heartbeatWatchTimer != nil {
		n.heartbeatWatchTimer.Stop()
	}
	n.heartbeatWatchTimer = time.AfterFunc(config.HeartbeatTimeout, func() {
		select {
		case n.events <- Event{Type: HeartbeatWatchTimeout}:
		case <-n.ctx.Done():
		}
	})
	n.mu.Unlock()
}

func (n *Node) handleElection(fromID int, fromAddr string) {
	if fromID >= n.ID {
		return
	}

	peerAddr, ok := n.Peers[fromID]
	if !ok {
		host, _, err := net.SplitHostPort(fromAddr)
		if err != nil {
			n.log.Warn("cannot determine peer address for OK", "from_id", fromID, "error", err)
			return
		}
		peerAddr = net.JoinHostPort(host, n.Port)
	}

	msg := protocol.NewMessage(protocol.OK, n.ID)
	if err := transport.SendMessage(peerAddr, msg); err != nil {
		n.log.Warn("send OK failed", "peer", peerAddr, "error", err)
	}

	n.mu.Lock()
	shouldStart := n.State != Candidate && n.State != Coordinator && n.BullyEnabled
	n.mu.Unlock()

	if shouldStart {
		n.startElection()
	}
}

func (n *Node) handleOK(fromID int) {
	n.mu.Lock()
	if n.State == Candidate {
		n.gotOK = true
		n.log.Debug("received OK", "from", fromID)
	}
	n.mu.Unlock()
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
		n.log.Warn("coordinator promotion blocked: bully algorithm disabled")
		return
	}
	n.State = Coordinator
	n.CoordinatorID = n.ID
	n.mu.Unlock()

	n.log.Info("became coordinator", "node_id", n.ID)

	msg := protocol.NewCoordinatorMessage(n.ID)
	n.broadcast(msg)

	n.startHeartbeats()
	n.startHTTPServer()
}

func (n *Node) handleCoordinator(coordID int) {
	if coordID == n.ID {
		return
	}

	n.mu.Lock()
	currentCoord := n.CoordinatorID
	currentState := n.State
	n.mu.Unlock()

	if coordID > currentCoord {
		n.mu.Lock()
		n.CoordinatorID = coordID
		n.State = Follower
		n.mu.Unlock()

		n.log.Info("accepting new coordinator", "coordinator_id", coordID)
		n.stopHTTPServer()
		n.resetHeartbeatWatch()
	} else if coordID < n.ID && currentState == Coordinator {
		n.log.Warn("lower-ID node claims coordinator, asserting myself")
		n.broadcast(protocol.NewCoordinatorMessage(n.ID))
	}
}
