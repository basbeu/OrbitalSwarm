package extramessage

import (
	"gonum.org/v1/gonum/spatial/r3"
)

// SwarmInit initiates the mapping phase for the swarm. It carries the initials positions of the drones and the target positions. The two list should have the same length.
type SwarmInit struct {
	PatternID  string
	InitialPos []r3.Vec
	TargetPos  []r3.Vec
}
