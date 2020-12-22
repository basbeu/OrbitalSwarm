package drone

type vec3d struct {
	x float32
	y float32
	z float32
}

func (v vec3d) X() float32 {
	return v.x
}

func (v vec3d) Y() float32 {
	return v.y
}

func (v vec3d) Z() float32 {
	return v.z
}

func (v vec3d) DistanceTo(v2 vec3d) float32 {
	return Math.Sqrt(v.x*v2.x + v.y*v2.y + v.z*v2.y)
}
