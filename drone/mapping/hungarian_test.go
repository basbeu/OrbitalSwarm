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
		r3.Vec{1, 0, 0},
		r3.Vec{2, 0, 0},
	}
	targetPos := []r3.Vec{
		r3.Vec{3, 0, 0},
		r3.Vec{4, 0, 0},
	}

	expectedMatrix := mat.NewDense(2, 2, []float64{
		2, 3,
		1, 2,
	})

	matrix := mapper.initMatrix(initialPos, targetPos)

	require.True(t, mat.Equal(expectedMatrix, matrix))
}
