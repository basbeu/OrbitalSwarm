package pathgenerator

import (
	"log"
	"math"
	"math/rand"
	"time"

	"gonum.org/v1/gonum/spatial/r3"
)

type SimplePathGenerator struct {
	done chan [][]r3.Vec
}

func NewSimplePathGenerator() *SimplePathGenerator {
	rand.Seed(time.Now().UnixNano())
	g := &SimplePathGenerator{
		done: make(chan [][]r3.Vec),
	}
	return g
}

func (g *SimplePathGenerator) GeneratePath(from []r3.Vec, dest []r3.Vec) <-chan [][]r3.Vec {
	go func() {
		paths := generateBasicPath(from, dest)
		for {
			if validatePaths(from, dest, paths) {
				break
			}
			paths = g.mix(paths)
		}
		g.done <- paths
	}()
	return g.done
}

// generateBasicPath generate a basic path such as to have a minimal set of moves to reach the mapping target
func generateBasicPath(from []r3.Vec, dest []r3.Vec) [][]r3.Vec {
	// Unit moves
	still := r3.Vec{X: 0, Y: 0, Z: 0}
	right := r3.Vec{X: 1, Y: 0, Z: 0}
	left := r3.Vec{X: -1, Y: 0, Z: 0}
	up := r3.Vec{X: 0, Y: 1, Z: 0}
	down := r3.Vec{X: 0, Y: -1, Z: 0}
	forward := r3.Vec{X: 0, Y: 0, Z: 1}
	backward := r3.Vec{X: 0, Y: 0, Z: -1}

	// Path
	paths := make([][]r3.Vec, len(from))
	for i := range paths {
		paths[i] = make([]r3.Vec, 0)
	}

	tmpFrom := make([]r3.Vec, len(from))
	tmpDest := make([]r3.Vec, len(dest))
	copy(tmpFrom, from)
	copy(tmpDest, dest)

	// First phase - floor step
	noMoves := true
	singleMove := true
	floorMove := make([]r3.Vec, len(from))
	precisionMove := make([]r3.Vec, len(from))
	for i := range paths {
		floorMove[i] = r3.Vec{X: math.Floor(from[i].X), Y: math.Floor(from[i].Y), Z: math.Floor(from[i].Z)}.Sub(from[i])
		precisionMove[i] = dest[i].Sub(r3.Vec{X: math.Floor(dest[i].X), Y: math.Floor(dest[i].Y), Z: math.Floor(dest[i].Z)})
		tmpFrom[i] = from[i].Add(floorMove[i])

		singleMove = singleMove && r3.Norm(floorMove[i]) <= 1.0
		noMoves = noMoves && floorMove[i] == still
	}

	singleMoveAdd := func(path []r3.Vec, move r3.Vec) []r3.Vec {
		return append(path, move)
	}
	defaultMoveAdd := func(path []r3.Vec, move r3.Vec) []r3.Vec {
		return append(path, []r3.Vec{move.Scale(0.5), move.Scale(0.5)}...)
	}

	if !noMoves {
		addFunction := defaultMoveAdd
		if singleMove {
			addFunction = singleMoveAdd
		}
		for i, moves := range floorMove {
			paths[i] = addFunction(paths[i], moves)
		}
	}

	// Second phase - unit moves
	remaining := make([]r3.Vec, len(from))
	for i, f := range tmpFrom {
		remaining[i] = dest[i].Sub(f)
	}

	targetReached := 0
	for targetReached != len(paths) {
		targetReached = 0
		for i, v := range remaining {
			if v.Y >= 1 {
				paths[i] = append(paths[i], up)
				remaining[i] = remaining[i].Sub(up)
			} else if v.X >= 1 {
				paths[i] = append(paths[i], right)
				remaining[i] = remaining[i].Sub(right)
			} else if v.Z >= 1 {
				paths[i] = append(paths[i], forward)
				remaining[i] = remaining[i].Sub(forward)
			} else if v.Z <= -1 {
				paths[i] = append(paths[i], backward)
				remaining[i] = remaining[i].Sub(backward)
			} else if v.X <= -1 {
				paths[i] = append(paths[i], left)
				remaining[i] = remaining[i].Sub(left)
			} else if v.Y <= -1 {
				paths[i] = append(paths[i], down)
				remaining[i] = remaining[i].Sub(down)
			} else {
				paths[i] = append(paths[i], still)
			}
			if remaining[i].X > -1 && remaining[i].X < 1 && remaining[i].Y > -1 && remaining[i].Y < 1 && remaining[i].Z > -1 && remaining[i].Z < 1 {
				targetReached++
			}
		}
	}

	// Third phase - floor step
	noMoves = true
	singleMove = true
	for i := range paths {
		precisionMove[i] = dest[i].Sub(r3.Vec{X: math.Floor(dest[i].X), Y: math.Floor(dest[i].Y), Z: math.Floor(dest[i].Z)})

		singleMove = singleMove && r3.Norm(precisionMove[i]) <= 1.0
		noMoves = noMoves && precisionMove[i] == still
	}

	if !noMoves {
		addFunction := defaultMoveAdd
		if singleMove {
			addFunction = singleMoveAdd
		}
		for i, move := range precisionMove {
			paths[i] = addFunction(paths[i], move)
		}
	}

	return paths
}

func (p *SimplePathGenerator) mix(paths [][]r3.Vec) [][]r3.Vec {
	for _, path := range paths {
		rand.Shuffle(len(path), func(i, j int) { path[i], path[j] = path[j], path[i] })
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
		locations[i][0] = location
	}

	if len(paths) == 0 {
		return true
	}

	for round := 1; round <= len(paths[0]); round++ {
		set := make(map[r3.Vec]bool)

		// Detect position identical for 2 drones or underground
		for i := 0; i < len(paths); i++ {
			locations[i][round] = locations[i][round-1].Add(paths[i][round-1])
			if locations[i][round].Y < 0 {
				log.Printf("ERROR : drone under ground level")
				return false
			}
			if set[locations[i][round]] == true {
				log.Printf("ERROR : drone at same location at given round")
				return false
			}
			set[locations[i][round]] = true
		}

		// Detect position exchange for 2 drones
		for i := 0; i < len(locations); i++ {
			for j := i + 1; j < len(locations); j++ {
				if locations[i][round] == locations[j][round-1] &&
					locations[j][round-1] == locations[i][round] {
					log.Printf("ERROR : Drone switch location")
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
