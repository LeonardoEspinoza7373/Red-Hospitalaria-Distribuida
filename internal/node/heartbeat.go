package node

import (
	"context"
	"time"

	"github.com/LeonardoEspinoza7373/Red-Hospitalaria-Distribuida/internal/protocol"
	"github.com/LeonardoEspinoza7373/Red-Hospitalaria-Distribuida/pkg/config"
)

func (n *Node) startHeartbeats() {
	n.mu.Lock()
	if n.heartbeatCancel != nil {
		n.heartbeatCancel()
	}
	ctx, cancel := context.WithCancel(n.ctx)
	n.heartbeatCancel = cancel
	n.mu.Unlock()

	n.log.Info("starting heartbeat sender", "category", "system")

	go func() {
		ticker := time.NewTicker(config.HeartbeatInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				n.mu.Lock()
				isCoord := n.State == Coordinator
				coordID := n.CoordinatorID
				myID := n.ID
				n.mu.Unlock()

				if !isCoord {
					return
				}

				msg := protocol.NewHeartbeatMessage(myID, coordID)
				n.broadcast(msg)
				n.log.Debug("sent HEARTBEAT", "category", "heartbeat")
			}
		}
	}()
}

func (n *Node) resetHeartbeatWatch() {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.heartbeatWatchTimer != nil {
		n.heartbeatWatchTimer.Stop()
	}

	n.heartbeatWatchTimer = time.AfterFunc(config.HeartbeatTimeout, func() {
		select {
		case n.events <- Event{Type: HeartbeatWatchTimeout}:
		case <-n.ctx.Done():
		}
	})
}

func (n *Node) handleHeartbeat(fromID, coordID int) {
	n.mu.Lock()
	isCoord := n.State == Coordinator
	myID := n.ID
	myCoordID := n.CoordinatorID
	n.mu.Unlock()

	if isCoord && coordID != myID {
		if myID > coordID {
			n.log.Warn("detected lower-ID coordinator, asserting dominance", "category", "bully")
			n.broadcast(protocol.NewCoordinatorMessage(myID))
			return
		}
		n.mu.Lock()
		n.State = Follower
		n.CoordinatorID = coordID
		n.mu.Unlock()
		n.log.Info("stepping down for higher-ID coordinator", "new_coord", coordID, "category", "bully")
		n.resetHeartbeatWatch()
		return
	}

	if coordID < myID {
		n.mu.Lock()
		enabled := n.BullyEnabled
		n.mu.Unlock()
		if enabled {
			n.log.Info("detected lower-ID coordinator, starting election", "category", "bully")
			n.startElection()
		} else {
			n.log.Warn("detected lower-ID coordinator but bully disabled, ignoring", "category", "bully")
		}
		return
	}

	if coordID != myCoordID {
		n.mu.Lock()
		n.CoordinatorID = coordID
		n.mu.Unlock()
		n.log.Info("coordinator updated", "new_coord", coordID, "category", "bully")
	}

	n.resetHeartbeatWatch()
}

func (n *Node) handleHeartbeatWatchTimeout() {
	n.mu.Lock()
	if n.State == Coordinator {
		n.mu.Unlock()
		return
	}
	bullyEnabled := n.BullyEnabled
	n.mu.Unlock()

	n.log.Warn("heartbeat timeout - coordinator may be down", "category", "bully")
	if bullyEnabled {
		n.startElection()
	} else {
		n.log.Warn("bully algorithm disabled - not starting election, proxy will return 503", "category", "bully")
	}
}
