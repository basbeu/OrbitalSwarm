package utils

import "math"

type Vec3d struct {
	x float64
	y float64
	z float64
}

func NewVec3d(x, y, z float64) Vec3d {
	return Vec3d{x, y, z}
}

func (v Vec3d) X() float64 {
	return v.x
}

func (v Vec3d) Y() float64 {
	return v.y
}

func (v Vec3d) Z() float64 {
	return v.z
}

func (v Vec3d) DistanceTo(v2 Vec3d) float64 {
	return math.Sqrt(v.x*v2.x + v.y*v2.y + v.z*v2.y)
}
