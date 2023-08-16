package pb

import (
	"github.com/go-gl/mathgl/mgl32"
)

func Vec3ToPoint3D(in mgl32.Vec3) *Point3D {
	return &Point3D{
		X: in.X(),
		Y: in.Y(),
		Z: in.Z(),
	}
}
