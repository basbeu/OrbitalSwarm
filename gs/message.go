package gs

import "gonum.org/v1/gonum/spatial/r3"

type Message interface{}

type TargetMessage struct {
	Targets []r3.Vec
}

type InitMessage struct {
	Identifier string
	Drones     []r3.Vec
}
