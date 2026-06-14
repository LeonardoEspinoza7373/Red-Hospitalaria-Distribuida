package node

import (
	"time"

	"github.com/LeonardoEspinoza7373/Red-Hospitalaria-Distribuida/pkg/config"
)

// resetHeartbeatWatch cancels any pending heartbeat watch timer and starts a new one.
// When the timer fires, it posts HeartbeatWatchTimeout to the events channel.
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

// scheduleTimeSync schedules an initial time sync after a short delay.
func (n *Node) scheduleTimeSync() {
	time.AfterFunc(2*time.Second, func() {
		n.syncWithCoordinator()
	})
}
