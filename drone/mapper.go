package drone

import (
	"go.dedis.ch/cs438/orbitalswarm/utils"
)

//
type targetsMapper interface {
	mapTargets(initials []utils.Vec3d, targets []utils.Vec3d) map[string]utils.Vec3d
}

type hungarianGraphConsensus struct {
}

func newHungarianGraphConsensus() *hungarianGraphConsensus {
	return &hungarianGraphConsensus{}
}

func (mapper *hungarianGraphConsensus) mapTargets(initials []utils.Vec3d, targets []utils.Vec3d) map[string]utils.Vec3d {
	//TODO
	return make(map[string]utils.Vec3d)
}
