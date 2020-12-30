package mapping

import (
	"testing"

	"gonum.org/v1/gonum/mat"

	"github.com/stretchr/testify/require"
	"gonum.org/v1/gonum/spatial/r3"
)

func TestInitMatrix(t *testing.T) {
	mapper := NewHungarianMapper()

	initialPos := []r3.Vec{
		r3.Vec{X: 1, Y: 0, Z: 0},
		r3.Vec{X: 2, Y: 0, Z: 0},
	}
	targetPos := []r3.Vec{
		r3.Vec{X: 3, Y: 0, Z: 0},
		r3.Vec{X: 4, Y: 0, Z: 0},
	}

	expectedMatrix := mat.NewDense(2, 2, []float64{
		2, 3,
		1, 2,
	})

	matrix := mapper.initMatrix(initialPos, targetPos)

	require.True(t, mat.Equal(expectedMatrix, matrix))

	initialPos = []r3.Vec{
		r3.Vec{X: 1, Y: 0, Z: 0},
		r3.Vec{X: 0, Y: 2, Z: 0},
		r3.Vec{X: 0, Y: 0, Z: 3},
	}
	targetPos = []r3.Vec{
		r3.Vec{X: 1, Y: 1, Z: 1},
		r3.Vec{X: 2, Y: 2, Z: 2},
		r3.Vec{X: 3, Y: 3, Z: 3},
	}

	expectedMatrix = mat.NewDense(3, 3, []float64{
		1, 3, 4,
		1, 2, 4,
		2, 3, 4,
	})

	matrix = mapper.initMatrix(initialPos, targetPos)

	require.True(t, mat.Equal(expectedMatrix, matrix))
}
