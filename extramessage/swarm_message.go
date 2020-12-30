package extramessage

import "go.dedis.ch/cs438/orbitalswarm/utils"

// SwarmInit initiates the mapping phase for the swarm. It carries the initials positions of the drones and the target positions. The two list should have the same length.
type SwarmInit struct {
	InitialPos []utils.Vec3d
	TargetPos  []utils.Vec3d
}
