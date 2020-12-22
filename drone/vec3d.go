package drone

import "math"

type vec3d struct {
	x float64
	y float64
	z float64
}

func (v vec3d) X() float64 {
	return v.x
}

func (v vec3d) Y() float64 {
	return v.y
}

func (v vec3d) Z() float64 {
	return v.z
}

func (v vec3d) DistanceTo(v2 vec3d) float64 {
	return math.Sqrt(v.x*v2.x + v.y*v2.y + v.z*v2.y)
}
