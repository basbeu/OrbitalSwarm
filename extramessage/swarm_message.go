package extramessage

import "go.dedis.ch/cs438/orbitalswarm/utils"

// SwarmInit initiates the mapping phase for the swarm. It carries the initials positions of the drones and the target positions. The two list should have the same length.
type SwarmInit struct {
	Initials []utils.Vec3d
	Targets  []utils.Vec3d
}