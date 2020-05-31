package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"github.com/xpzouying/graft"
)

func panicError(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	ci := graft.ClusterInfo{Name: "zy_cluster", Size: 3}
	opts := &nats.DefaultOptions
	opts.Servers = []string{"localhost:4222"}
	rpc, err := graft.NewNatsRpc(opts)
	panicError(err)

	errChan := make(chan error)
	stateChangeChan := make(chan graft.StateChange)
	handler := graft.NewChanHandler(stateChangeChan, errChan)

	logfile := fmt.Sprintf("/tmp/graft-%d.log", rand.Int31n(100))
	node, err := graft.New(ci, handler, rpc, logfile)
	panicError(err)
	defer node.Close()

	for {
		if node.State() == graft.LEADER {
			// Process as a LEADER
			logrus.Infof("node(%v) is leader", node.Id())
		}

		select {
		case sc := <-stateChangeChan:
			// Process a state change
			logrus.Warnf("node(%v) state changed: %v", node.Id(), sc)
		case err := <-errChan:
			// Process an error, log etc.
			logrus.Errorf("node(%v) got error: %v", node.Id(), err)
		default:
		}

		logrus.Infof("node(%s) - current leader:%s - current_term:%v", node.Id(), node.Leader(), node.CurrentTerm())

		time.Sleep(1 * time.Second)
	}
}
