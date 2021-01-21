package utils

import "math"

// Vec3d represents an mutable minimal tri-dimensional vector
type Vec3d struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

// NewVec3d creates and returns a tri-dimensional vector
func NewVec3d(x, y, z float64) Vec3d {
	return Vec3d{X: x, Y: y, Z: z}
}

// DistanceTo computes the Euclidean distance with v2
func (v Vec3d) DistanceTo(v2 Vec3d) float64 {
	return math.Sqrt(v.X*v2.X + v.Y*v2.Y + v.Z*v2.Z)
}

// Clone create a deep copy of a Vec3d
func (v Vec3d) Clone() Vec3d {
	return Vec3d{
		X: v.X,
		Y: v.Y,
		Z: v.Z,
	}
}

// Add the given vector to the current vector
func (v Vec3d) Add(vec Vec3d) {
	v.X += vec.X
	v.Y += vec.Y
	v.Z += vec.Z
}
