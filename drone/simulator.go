package drone

import (
	"time"

	"gonum.org/v1/gonum/spatial/r3"
)

type interface_drone interface {
	UpdateLocation(r3.Vec)
}

type simulator struct {
	drone interface_drone
	done  chan struct{}
}

func NewSimulator(drone interface_drone) *simulator {
	return &simulator{
		drone: drone,
		done:  make(chan struct{}),
	}
}

func (s *simulator) launchSimulation(singleMoveTime int, refreshFrequency int, location r3.Vec, path []r3.Vec) <-chan struct{} {
	go func() {
		sleepDuration := time.Duration(1000/refreshFrequency) * time.Millisecond
		for _, move := range path {
			stepMove := move.Scale(float64(singleMoveTime) / float64(refreshFrequency))
			for step := 1; step <= singleMoveTime*refreshFrequency; step++ {
				time.Sleep(sleepDuration)
				tempLocation := location.Add(stepMove.Scale(float64(step)))
				s.drone.UpdateLocation(tempLocation)
			}
			location = location.Add(move)
		}
		close(s.done)
	}()
	return s.done
}
