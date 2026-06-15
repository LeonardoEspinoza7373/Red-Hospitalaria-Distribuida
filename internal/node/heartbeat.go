package node

import "time"

func (n *Node) scheduleTimeSync() {
	time.AfterFunc(2*time.Second, func() {
		n.syncWithCoordinator()
	})
}
