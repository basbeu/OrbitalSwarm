package drone

import (
	"time"

	"gonum.org/v1/gonum/spatial/r3"
)

type simulator struct {
	drone *Drone
	done  chan struct{}
}

func NewSimulator(drone *Drone) *simulator {
	return &simulator{
		drone: drone,
		done:  make(chan struct{}),
	}
}

func (s *simulator) launchSimulation(singleMoveTime int, refreshFrequency int, location r3.Vec, path []r3.Vec) {
	go func() {
		currentLocation := location
		sleepDuration := time.Duration(1000/refreshFrequency) * time.Millisecond
		for _, move := range path {
			stepMove := move.Scale(float64(singleMoveTime) / float64(refreshFrequency))
			for step := 1; step < singleMoveTime*refreshFrequency; step++ {
				time.Sleep(sleepDuration)
				currentLocation = location.Add(stepMove.Scale(float64(step)))
				s.drone.UpdateLocation(currentLocation)
			}
			currentLocation = currentLocation.Add(move)
		}
		close(s.done)
	}()
}
