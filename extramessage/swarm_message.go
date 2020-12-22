package extramessage

import "go.dedis.ch/cs438/orbitalswarm/utils"

// SwarmInit initiates the mapping phase for the swarm. It carries the initials positions of the drones and the target positions. The two list should have the same length.
type SwarmInit struct {
	initials []utils.Vec3d
	targets  []utils.Vec3d
}
