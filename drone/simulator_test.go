package drone

import (
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"gonum.org/v1/gonum/spatial/r3"
)

type mockDrone struct {
	res []r3.Vec
}

func newMockDrone() *mockDrone {
	return &mockDrone{}
}

func (d *mockDrone) UpdateLocation(location r3.Vec) {
	d.res = append(d.res, location)
}

func TestSimulator(t *testing.T) {
	drone := newMockDrone()
	simulator := NewSimulator(drone)
	timeOneStep := 1

	starting := r3.Vec{X: 0, Y: 0, Z: 0}
	path := []r3.Vec{
		r3.Vec{X: 1, Y: 0, Z: 0},
		r3.Vec{X: 0, Y: 1, Z: 0},
		r3.Vec{X: 0, Y: 0, Z: 1},
		r3.Vec{X: -1, Y: 0, Z: 0},
	}
	simulator.launchSimulation(timeOneStep, 4, starting, path)

	expected := []r3.Vec{
		r3.Vec{X: 0.25, Y: 0, Z: 0},
		r3.Vec{X: 0.5, Y: 0, Z: 0},
		r3.Vec{X: 0.75, Y: 0, Z: 0},
		r3.Vec{X: 1, Y: 0, Z: 0},
		r3.Vec{X: 1, Y: 0.25, Z: 0},
		r3.Vec{X: 1, Y: 0.5, Z: 0},
		r3.Vec{X: 1, Y: 0.75, Z: 0},
		r3.Vec{X: 1, Y: 1, Z: 0},
		r3.Vec{X: 1, Y: 1, Z: 0.25},
		r3.Vec{X: 1, Y: 1, Z: 0.5},
		r3.Vec{X: 1, Y: 1, Z: 0.75},
		r3.Vec{X: 1, Y: 1, Z: 1},
		r3.Vec{X: 0.75, Y: 1, Z: 1},
		r3.Vec{X: 0.5, Y: 1, Z: 1},
		r3.Vec{X: 0.25, Y: 1, Z: 1},
		r3.Vec{X: 0, Y: 1, Z: 1},
	}

	time.Sleep(time.Second * time.Duration(len(path)+1))

	log.Println(expected)
	log.Println(drone.res)
	require.Equal(t, len(expected), len(drone.res))

	for i := range expected {
		require.Equal(t, expected[i], drone.res[i])
	}
}
