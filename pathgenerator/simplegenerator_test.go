package pathgenerator

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

func TestGenerateBasicPath_NoFloorAndNoPrecisionSteps(t *testing.T) {
	from := []r3.Vec{
		r3.Vec{X: 0, Y: 0, Z: 0},
		r3.Vec{X: 1, Y: 0, Z: 0},
		r3.Vec{X: 0, Y: 0, Z: 1},
		r3.Vec{X: 1, Y: 0, Z: 1},
	}
	dest := []r3.Vec{
		r3.Vec{X: -2, Y: 1, Z: 0}, // 3
		r3.Vec{X: 2, Y: 2, Z: 0},  // 3
		r3.Vec{X: 0, Y: 2, Z: 1},  // 2
		r3.Vec{X: 2, Y: 2, Z: 1},  // 3
	}
	res := generateBasicPath(from, dest)

	expected := [][]r3.Vec{
		[]r3.Vec{
			r3.Vec{X: 0, Y: 1, Z: 0},
			r3.Vec{X: -1, Y: 0, Z: 0},
			r3.Vec{X: -1, Y: 0, Z: 0},
		},
		[]r3.Vec{
			r3.Vec{X: 0, Y: 1, Z: 0},
			r3.Vec{X: 0, Y: 1, Z: 0},
			r3.Vec{X: 1, Y: 0, Z: 0},
		},
		[]r3.Vec{
			r3.Vec{X: 0, Y: 1, Z: 0},
			r3.Vec{X: 0, Y: 1, Z: 0},
			r3.Vec{X: 0, Y: 0, Z: 0},
		},
		[]r3.Vec{
			r3.Vec{X: 0, Y: 1, Z: 0},
			r3.Vec{X: 0, Y: 1, Z: 0},
			r3.Vec{X: 1, Y: 0, Z: 0},
		},
	}

	time.Sleep(time.Second * time.Duration(3))

	log.Println(expected)
	log.Println(res)
	require.Equal(t, len(expected), len(res))
	for i := range expected {
		require.Equal(t, len(expected[i]), len(res[i]))
		for j := range expected[i] {
			require.Equal(t, expected[i][j], res[i][j])
		}
	}

	require.Equal(t, true, validatePaths(from, dest, res))
}

func TestGenerateBasicPath_WithFloorStep(t *testing.T) {
	from := []r3.Vec{
		r3.Vec{X: 0.12, Y: 0.5, Z: 0.9},
		r3.Vec{X: 1, Y: 0, Z: 0},
		r3.Vec{X: 0, Y: 0, Z: 1},
	}
	dest := []r3.Vec{
		r3.Vec{X: -2, Y: 1, Z: 0}, // 3
		r3.Vec{X: 2, Y: 2, Z: 0},  // 3
		r3.Vec{X: 0, Y: 2, Z: 1},  // 2
	}
	res := generateBasicPath(from, dest)

	expected := [][]r3.Vec{
		[]r3.Vec{
			r3.Vec{X: -0.06, Y: -0.25, Z: -0.45},
			r3.Vec{X: -0.06, Y: -0.25, Z: -0.45},
			r3.Vec{X: 0, Y: 1, Z: 0},
			r3.Vec{X: -1, Y: 0, Z: 0},
			r3.Vec{X: -1, Y: 0, Z: 0},
		},
		[]r3.Vec{
			r3.Vec{X: 0, Y: 0, Z: 0},
			r3.Vec{X: 0, Y: 0, Z: 0},
			r3.Vec{X: 0, Y: 1, Z: 0},
			r3.Vec{X: 0, Y: 1, Z: 0},
			r3.Vec{X: 1, Y: 0, Z: 0},
		},
		[]r3.Vec{
			r3.Vec{X: 0, Y: 0, Z: 0},
			r3.Vec{X: 0, Y: 0, Z: 0},
			r3.Vec{X: 0, Y: 1, Z: 0},
			r3.Vec{X: 0, Y: 1, Z: 0},
			r3.Vec{X: 0, Y: 0, Z: 0},
		},
	}

	time.Sleep(time.Second * time.Duration(3))

	log.Println("Expected")
	log.Println(expected)
	log.Println("Res")
	log.Println(res)
	require.Equal(t, len(expected), len(res))
	for i := range expected {
		require.Equal(t, len(expected[i]), len(res[i]))
		for j := range expected[i] {
			require.Equal(t, expected[i][j], res[i][j])
		}
	}

	require.Equal(t, true, validatePaths(from, dest, res))
}

func TestGenerateBasicPath_WithPrecisionStep(t *testing.T) {
	from := []r3.Vec{
		r3.Vec{X: 0, Y: 0, Z: 0},
		r3.Vec{X: 1, Y: 0, Z: 0},
		r3.Vec{X: 0, Y: 0, Z: 1},
	}
	dest := []r3.Vec{
		r3.Vec{X: -2, Y: 1, Z: 0},       // 3
		r3.Vec{X: 2, Y: 2, Z: 0},        // 3
		r3.Vec{X: 0.5, Y: 2.12, Z: 1.2}, // 2
	}
	res := generateBasicPath(from, dest)

	expected := [][]r3.Vec{
		[]r3.Vec{
			r3.Vec{X: 0, Y: 1, Z: 0},
			r3.Vec{X: -1, Y: 0, Z: 0},
			r3.Vec{X: -1, Y: 0, Z: 0},
			r3.Vec{X: 0, Y: 0, Z: 0},
		},
		[]r3.Vec{
			r3.Vec{X: 0, Y: 1, Z: 0},
			r3.Vec{X: 0, Y: 1, Z: 0},
			r3.Vec{X: 1, Y: 0, Z: 0},
			r3.Vec{X: 0, Y: 0, Z: 0},
		},
		[]r3.Vec{
			r3.Vec{X: 0, Y: 1, Z: 0},
			r3.Vec{X: 0, Y: 1, Z: 0},
			r3.Vec{X: 0, Y: 0, Z: 0},
			r3.Vec{X: 0.5, Y: 0.12, Z: 0.2},
		},
	}

	time.Sleep(time.Second * time.Duration(3))

	log.Println("Expected")
	log.Println(expected)
	log.Println("Res")
	log.Println(res)
	require.Equal(t, len(expected), len(res))

	const float64EqualityThreshold = 1e-9

	for i := range expected {
		require.Equal(t, len(expected[i]), len(res[i]))
		for j := range expected[i] {
			require.Equal(t, true, r3.Norm(expected[i][j].Sub(res[i][j])) <= float64EqualityThreshold)
		}
	}

	require.Equal(t, true, validatePaths(from, dest, res))
}
func TestValidatePaths_exchange(t *testing.T) {
	from := []r3.Vec{
		r3.Vec{X: 0, Y: 0, Z: 0},
		r3.Vec{X: 1, Y: 0, Z: 0},
	}
	dest := []r3.Vec{
		r3.Vec{X: 1, Y: 1, Z: 0}, // 3
		r3.Vec{X: 0, Y: 1, Z: 0}, // 3
	}
	res := generateBasicPath(from, dest)

	time.Sleep(time.Second * time.Duration(3))

	require.Equal(t, false, validatePaths(from, dest, res))
}

func TestValidatePaths_cross(t *testing.T) {
	from := []r3.Vec{
		r3.Vec{X: 0, Y: 0, Z: 0},
		r3.Vec{X: 2, Y: 0, Z: 0},
	}
	dest := []r3.Vec{
		r3.Vec{X: 2, Y: 1, Z: 0}, // 3
		r3.Vec{X: 0, Y: 1, Z: 0}, // 3
	}
	res := generateBasicPath(from, dest)

	time.Sleep(time.Second * time.Duration(3))

	require.Equal(t, false, validatePaths(from, dest, res))
}
