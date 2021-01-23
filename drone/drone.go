// ========== CS-438 orbitalswarm Skeleton ===========

package drone

import (
	"net"
	"strconv"
	"sync"

	"go.dedis.ch/cs438/orbitalswarm/drone/mapping"
	"go.dedis.ch/cs438/orbitalswarm/pathgenerator"
	"gonum.org/v1/gonum/spatial/r3"

	"go.dedis.ch/cs438/orbitalswarm/gossip"

	"github.com/rs/zerolog/log"
)

type state int

const (
	MAPPING state = iota
	GENERATING_PATH
	MOVING
	IDLE
)

// Drone is responsible to be the glue between the gossiping protocol and
// the ui, dispatching responses and messages etc
type Drone struct {
	sync.Mutex
	droneID         uint32
	status          state
	uiAddress       string
	identifier      string
	gossipAddress   string
	gossiper        *gossip.Gossiper
	cliConn         net.Conn
	position        r3.Vec
	target          r3.Vec
	consensusClient *ConsensusClient
	targetsMapper   mapping.TargetsMapper
	pathGenerator   pathgenerator.PathGenerator
	path            []r3.Vec

	simulator *simulator
}

// NewDrone returns the controller that sets up the gossiping state machine
// as well as the web routing. It uses the same gossiping address for the
// identifier.
func NewDrone(droneID uint32, identifier, uiAddress, gossipAddress string,
	g *gossip.Gossiper, addresses []string, position r3.Vec, targetsMapper mapping.TargetsMapper, consensusClient *ConsensusClient, pathGenerator pathgenerator.PathGenerator) *Drone {

	d := &Drone{
		droneID:         droneID,
		status:          IDLE,
		identifier:      identifier,
		uiAddress:       uiAddress,
		gossipAddress:   gossipAddress,
		gossiper:        g,
		consensusClient: consensusClient,
		position:        position,
		targetsMapper:   targetsMapper,
		pathGenerator:   pathGenerator,
	}
	g.AddAddresses(addresses...)

	g.RegisterCallback(d.HandleGossipMessage)
	return d
}

// Run ...
func (d *Drone) Run() {
	d.simulator = NewSimulator(d)
}

// UpdateLocation of the drone
func (d *Drone) UpdateLocation(location r3.Vec) {
	d.gossiper.AddPrivateMessage(gossip.PrivateMessageData{
		Location: location,
		DroneID:  d.droneID,
	}, "GS", d.gossiper.GetIdentifier(), 10)
	d.position = location
}

// HandleGossipMessage handle specific messages concerning the drone
func (d *Drone) HandleGossipMessage(origin string, msg gossip.GossipPacket) {

	if msg.Rumor != nil {
		if msg.Rumor.Extra != nil {
			if msg.Rumor.Extra.SwarmInit != nil && d.status == IDLE {
				go func() {
					log.Printf("%s Swarm init received", d.identifier)
					//Begin mapping phase
					d.status = MAPPING
					dronePos := msg.Rumor.Extra.SwarmInit.InitialPos
					patternID := msg.Rumor.Extra.SwarmInit.PatternID
					log.Printf("%s Start mapping", d.identifier)
					target := d.targetsMapper.MapTargets(dronePos, msg.Rumor.Extra.SwarmInit.TargetPos)
					targets := target
					// targets, _ := d.consensusClient.ProposeTargets(d.gossiper, patternID, target)
					d.target = targets[d.droneID]

					d.status = GENERATING_PATH
					log.Printf("%s Generate path", d.identifier)
					chanPath := d.pathGenerator.GeneratePath(dronePos, targets)
					pathsGenerated := <-chanPath
					log.Printf("%s Propose path", d.identifier)
					paths, _ := d.consensusClient.ProposePaths(d.gossiper, patternID, pathsGenerated)
					d.path = paths[d.droneID]

					log.Printf("Start simulation")
					d.status = MOVING
					done := d.simulator.launchSimulation(1, 4, d.position, paths[d.droneID])
					<-done

					log.Printf("Simulation ended")
					d.gossiper.AddMessage(strconv.FormatUint(uint64(d.GetDroneID()), 10))

					d.status = IDLE
				}()
			} else {
				// log.Printf("Handle")
				d.consensusClient.HandleExtraMessage(d.gossiper, msg.Rumor.Extra)
			}
		}
	}
}

func (d *Drone) GetTarget() r3.Vec {
	return d.target
}
func (d *Drone) GetDroneID() uint32 {
	return d.droneID
}
