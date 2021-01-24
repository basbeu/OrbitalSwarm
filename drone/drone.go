// ========== CS-438 orbitalswarm Skeleton ===========

package drone

import (
	"strconv"
	"sync"

	"go.dedis.ch/cs438/orbitalswarm/paxos/blk"

	"go.dedis.ch/cs438/orbitalswarm/drone/consensus"
	"go.dedis.ch/cs438/orbitalswarm/drone/mapping"
	"go.dedis.ch/cs438/orbitalswarm/pathgenerator"
	"gonum.org/v1/gonum/spatial/r3"

	"go.dedis.ch/cs438/orbitalswarm/gossip"

	"go.dedis.ch/onet/v3/log"
)

type state int

const (
	IDLE state = iota
	READY
	MAPPING
	GENERATING_PATH
	MOVING
)

type Drone struct {
	droneID uint32
	status  state

	position r3.Vec
	target   r3.Vec
	path     []r3.Vec

	gossiper        *gossip.Gossiper
	consensusClient consensus.ConsensusClient
	targetsMapper   mapping.TargetsMapper
	pathGenerator   pathgenerator.PathGenerator
	simulator       *simulator

	muxFly sync.Mutex
}

func NewDrone(droneID uint32, g *gossip.Gossiper, addresses []string, position r3.Vec, targetsMapper mapping.TargetsMapper, consensusClient consensus.ConsensusClient, pathGenerator pathgenerator.PathGenerator) *Drone {

	d := &Drone{
		droneID: droneID,
		status:  IDLE,

		position: position,

		gossiper:        g,
		consensusClient: consensusClient,
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
				d.status = READY

				if d.consensusClient.IsProposer() {
					go func() {
						patternID := msg.Rumor.Extra.SwarmInit.PatternID
						dronePos := msg.Rumor.Extra.SwarmInit.InitialPos

						targets := d.mapTarget(patternID, dronePos, msg.Rumor.Extra.SwarmInit.TargetPos)

						d.generatePaths(patternID, dronePos, targets)

						d.fly()
					}()
				}
			} else {
				// log.Printf("Handle")
				blockContainer := d.consensusClient.HandleExtraMessage(d.gossiper, msg.Rumor.Extra)

				if blockContainer != nil {
					if blockContainer.Type == blk.BlockPathStr {
						blockContent := blockContainer.GetContent().(*blk.PathBlockContent)

						d.path = blockContent.Paths[d.droneID]
						go d.fly()
					}
				}
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

func (d *Drone) mapTarget(patternID string, initialPos, targetsPos []r3.Vec) []r3.Vec {
	log.Printf("%s Swarm init received", d.gossiper.GetIdentifier())
	//Begin mapping phase
	d.status = MAPPING
	log.Printf("%s Start mapping", d.gossiper.GetIdentifier())
	target := d.targetsMapper.MapTargets(initialPos, targetsPos)
	targets := target
	// targets := d.consensusClient.ProposeTargets(d.gossiper, patternID, target)
	d.target = targets[d.droneID]
	return targets
}

func (d *Drone) generatePaths(patternID string, dronePos, targets []r3.Vec) {
	d.status = GENERATING_PATH
	log.Printf("%s Generate path", d.gossiper.GetIdentifier())
	chanPath := d.pathGenerator.GeneratePath(dronePos, targets)
	pathsGenerated := <-chanPath
	log.Printf("%s Propose path", d.gossiper.GetIdentifier())
	paths := d.consensusClient.ProposePaths(d.gossiper, patternID, pathsGenerated)
	d.path = paths[d.droneID]
}

func (d *Drone) fly() {
	d.muxFly.Lock()
	defer d.muxFly.Unlock()

	if d.status != IDLE && d.status != MOVING {
		log.Printf(d.gossiper.GetIdentifier() + "Start simulation")
		d.status = MOVING
		done := d.simulator.launchSimulation(1, 4, d.position, d.path)
		<-done

		log.Printf("Simulation ended")
		d.gossiper.AddMessage(strconv.FormatUint(uint64(d.GetDroneID()), 10))

		d.status = IDLE
	}
}
