package utils

import "math"

// Vec3d represents an immutable minimal tri-dimensional vector
type Vec3d struct {
	x float64
	y float64
	z float64
}

// NewVec3d creates and returns a tri-dimensional vector
func NewVec3d(x, y, z float64) Vec3d {
	return Vec3d{x, y, z}
}

// X returns the first component  of the vector
func (v Vec3d) X() float64 {
	return v.x
}

// Y returns the second component of the vector
func (v Vec3d) Y() float64 {
	return v.y
}

// Z returns the third component of the vector
func (v Vec3d) Z() float64 {
	return v.z
}

// DistanceTo computes the Euclidean distance with v2
func (v Vec3d) DistanceTo(v2 Vec3d) float64 {
	return math.Sqrt(v.x*v2.x + v.y*v2.y + v.z*v2.y)
}
