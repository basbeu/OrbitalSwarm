package drone

import (
	"fmt"
	"math"

	"go.dedis.ch/cs438/orbitalswarm/drone/consensus"
	"go.dedis.ch/cs438/orbitalswarm/drone/mapping"
	"go.dedis.ch/cs438/orbitalswarm/gossip"
	"go.dedis.ch/cs438/orbitalswarm/pathgenerator"
	"gonum.org/v1/gonum/spatial/r3"
)

// Swarm represents a collections of drones that runs together
type Swarm struct {
	drones []*Drone
	stop   chan struct{}
}

// NewSwarm creates and returns an new Swarm, but do not start the drones
func NewSwarm(numDrones, numPaxosDrone, firstUIPort, firstGossipPort, antiEntropy, routeTimer, paxosRetry int, baseUIAddress, baseGossipAddress string) (*Swarm, []r3.Vec) {
	swarm := Swarm{
		drones: make([]*Drone, numDrones),
		stop:   make(chan struct{}),
	}

	// Drone parameters initialisation
	gossipAddresses := make([]string, numDrones)
	UIAddresses := make([]string, numDrones)
	positions := make([]r3.Vec, numDrones)
	line := 0
	column := 0
	space := 2

	edge := int(math.Sqrt(float64(numDrones)))
	for i := 0; i < numDrones; i++ {
		gossipAddress := fmt.Sprintf("%s:%d", baseGossipAddress, firstGossipPort+i)
		gossipAddresses[i] = gossipAddress
		UIAddress := fmt.Sprintf("%s:%d", baseUIAddress, firstUIPort+i)
		UIAddresses[i] = UIAddress
		positions[i] = r3.Vec{X: float64(line * space), Y: 0, Z: float64(column * space)}
		column = (column + 1) % edge
		if column == 0 {
			line++
		}
	}

	// Drone creation
	fac := gossip.GetFactory()
	for i := 0; i < numDrones; i++ {
		name := fmt.Sprintf("drone%d", i)
		g, err := fac.New(gossipAddresses[i], name, antiEntropy, routeTimer, numDrones)

		if err != nil {
			panic(err)
		}
		peers := make([]string, numDrones)
		copy(peers, gossipAddresses)
		peers = append(peers[:i], peers[i+1:]...)

		var consensusCli consensus.ConsensusClient

		if i < numPaxosDrone {
			consensusCli = consensus.NewConsensusParticipant(numPaxosDrone, i, paxosRetry)
		} else {
			consensusCli = consensus.NewConsensusReader(numPaxosDrone, i, paxosRetry)
		}

		swarm.drones[i] = NewDrone(uint32(i), g, peers, positions[i], mapping.NewHungarianMapper(), consensusCli, pathgenerator.NewSimplePathGenerator())
	}

	return &swarm, positions
}

// Run the drones composing the drones, this function is blocking until the the stop function is called
func (s *Swarm) Run() {
	for _, drone := range s.drones {
		ready := make(chan struct{})
		go drone.gossiper.Run(ready)
		defer drone.gossiper.Stop()
		<-ready

		go drone.Run()
	}
	<-s.stop
}

// Stop every drone composing the Swarm
func (s *Swarm) Stop() {
	close(s.stop)
}

// DronesAddresses return the drone addresses
func (s *Swarm) DronesAddresses() []string {
	addresses := make([]string, len(s.drones))
	for i, d := range s.drones {
		addresses[i] = d.gossiper.GetLocalAddr()
	}
	return addresses
}

// TO TEST, maybe not useful/good to keep it
func (s *Swarm) DroneTargets() []r3.Vec {
	targets := make([]r3.Vec, len(s.drones))
	for _, d := range s.drones {
		targets[d.GetDroneID()] = d.GetTarget()
	}
	return targets
}
