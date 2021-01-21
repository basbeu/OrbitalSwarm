package pathgenerator

import (
	"gonum.org/v1/gonum/spatial/r3"
)

type GeneticPathGenerator struct {
	done chan [][]r3.Vec
}

func NewGeneticPathGenerator() *GeneticPathGenerator {
	g := &GeneticPathGenerator{
		done: make(chan [][]r3.Vec),
	}
	return g
}

func (g *GeneticPathGenerator) GeneratePath(from []r3.Vec, dest []r3.Vec) <-chan [][]r3.Vec {
	go func() {
		paths := generateBasicPath(from, dest)
		if validatePaths(from, dest, paths) {
			g.done <- paths
		} else {
			println("Basic path not correct")
		}
	}()
	return g.done
}

func (g *GeneticPathGenerator) Stop() {

}

// v1 Using a simple shortest path
func v1(from []r3.Vec, dest []r3.Vec) {

}

// generateBasicPath generate a basic path such as to have a minimal set of moves to reach the mapping target
func generateBasicPath(from []r3.Vec, dest []r3.Vec) [][]r3.Vec {
	still := r3.Vec{X: 0, Y: 0, Z: 0}
	right := r3.Vec{X: 1, Y: 0, Z: 0}
	left := r3.Vec{X: -1, Y: 0, Z: 0}
	up := r3.Vec{X: 0, Y: 1, Z: 0}
	down := r3.Vec{X: 0, Y: -1, Z: 0}
	forward := r3.Vec{X: 0, Y: 0, Z: 1}
	backward := r3.Vec{X: 0, Y: 0, Z: -1}

	remaining := make([]r3.Vec, len(from))
	for i, f := range from {
		remaining[i] = dest[i].Sub(f)
	}

	paths := make([][]r3.Vec, len(from))
	for i := range paths {
		paths[i] = make([]r3.Vec, 0)
	}

	targetReached := 0
	for targetReached != len(paths) {
		targetReached = 0
		for i, v := range remaining {
			if v.X > 0 {
				paths[i] = append(paths[i], right)
				remaining[i] = remaining[i].Sub(right)
			} else if v.X < 0 {
				paths[i] = append(paths[i], left)
				remaining[i] = remaining[i].Sub(left)
			} else if v.Y > 0 {
				paths[i] = append(paths[i], up)
				remaining[i] = remaining[i].Sub(up)
			} else if v.Y < 0 {
				paths[i] = append(paths[i], down)
				remaining[i] = remaining[i].Sub(down)
			} else if v.Z > 0 {
				paths[i] = append(paths[i], forward)
				remaining[i] = remaining[i].Sub(forward)
			} else if v.Z < 0 {
				paths[i] = append(paths[i], backward)
				remaining[i] = remaining[i].Sub(backward)
			} else {
				paths[i] = append(paths[i], still)
				targetReached++
			}
		}
	}
	return paths
}

// validatePaths check that no path intersect at a time t
// paths : [pathId][...steps]
func validatePaths(from []r3.Vec, dest []r3.Vec, paths [][]r3.Vec) bool {
	locations := make([][]r3.Vec, len(from))
	for i := range locations {
		locations[i] = make([]r3.Vec, len(paths[0])+1)
	}
	for i, location := range from {
		locations[0][i] = location
	}

	if len(paths) == 0 {
		return true
	}

	for round := 1; round < len(paths[0]); round++ {
		set := make(map[r3.Vec]bool)
		// Detect position identical for 2 drones or underground
		for i := 0; i < len(paths); i++ {
			locations[i][round] = locations[i][round-1].Add(paths[i][round])

			if locations[i][round].Y < 0 || set[locations[i][round]] == true {
				println("ERROR : drone at same location at given round")
				return false
			}
			set[locations[i][round]] = true
		}

		// Detect position exchange for 2 drones
		for i := 0; i < len(locations); i++ {
			for j := i + 1; j < len(locations); j++ {
				if locations[i][round] == locations[j][round-1] &&
					locations[j][round-1] == locations[i][round] {
					println("ERROR : Drone switch location")
					return false
				}
			}
		}
	}

	// Validate final location
	for i := range locations {
		final := locations[i][len(paths[0])].Sub(dest[i])
		if final != (r3.Vec{X: 0, Y: 0, Z: 0}) {
			println("ERROR : Destination not reached")
			return false
		}
	}

	return true
}
