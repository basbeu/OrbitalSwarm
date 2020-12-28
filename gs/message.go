package gs

import "go.dedis.ch/cs438/orbitalswarm/utils"

type Message interface{}

type TargetMessage struct {
	targets []utils.Vec3d
}

type InitMessage struct {
	Identifier string
	Drones     []utils.Vec3d
}
