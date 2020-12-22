package drone

import (
	"fmt"
	"math"

	"go.dedis.ch/cs438/orbitalswarm/gossip"
	"go.dedis.ch/cs438/orbitalswarm/utils"
)

type Swarm struct {
	drones []*Drone
	stop   chan struct{}
}

func NewSwarm(numDrones, firstUIPort, firstGossipPort, antiEntropy, routeTimer, paxosRetry int, baseUIAddress, baseGossipAddress string) (*Swarm, []utils.Vec3d) {
	swarm := Swarm{
		drones: make([]*Drone, numDrones),
		stop:   make(chan struct{}),
	}

	gossipAddresses := make([]string, numDrones)
	UIAddresses := make([]string, numDrones)
	positions := make([]utils.Vec3d, numDrones)
	line := 0
	column := 0
	edge := int(math.Sqrt(float64(numDrones)))
	for i := 0; i < numDrones; i++ {
		gossipAddress := fmt.Sprintf("%s:%d", baseGossipAddress, firstGossipPort+i)
		gossipAddresses[i] = gossipAddress
		UIAddress := fmt.Sprintf("%s:%d", baseUIAddress, firstUIPort+i)
		UIAddresses[i] = UIAddress
		positions[i] = utils.NewVec3d(float64(line), float64(column), 0)
		column = (column + 1) % edge
		if column == 0 {
			line++
		}
	}

	fac := gossip.GetFactory()
	for i := 0; i < numDrones; i++ {
		name := fmt.Sprintf("drone%d", i)
		g, err := fac.New(gossipAddresses[i], name, antiEntropy, routeTimer, numDrones, i, paxosRetry)

		if err != nil {
			panic(err)
		}
		peers := make([]string, numDrones)
		copy(peers, gossipAddresses)
		peers = append(peers[:i], peers[i+1:]...)
		swarm.drones[i] = NewDrone(name, UIAddresses[i], gossipAddresses[i], g, peers, positions[i])
	}

	return &swarm, positions
}

func (swarm *Swarm) Run() {
	for _, drone := range swarm.drones {
		ready := make(chan struct{})
		go drone.gossiper.Run(ready)
		defer drone.gossiper.Stop()
		<-ready

		go drone.Run()
	}
	<-swarm.stop
}

func (swarm *Swarm) Stop() {
	close(swarm.stop)
}
