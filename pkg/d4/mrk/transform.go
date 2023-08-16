package mrk

import (
	"github.com/Dakota628/d4parse/pkg/d4"
	"github.com/go-gl/mathgl/mgl32"
	"sync"
)

type Transform struct {
	position mgl32.Vec3
	rotation mgl32.Quat
	scale    mgl32.Vec3
	parent   *Transform
	//children mapset.Set[*Transform]

	lwMatrix *mgl32.Mat4
	lwMu     sync.RWMutex
}

func NewRootTransform() *Transform {
	return NewTransform(0, 0, 0, 0, 0, 0, 1, 1, 1, 1)
}

func NewTransform(px, py, pz, qx, qy, qz, qw, sx, sy, sz float32) *Transform {
	return &Transform{
		position: mgl32.Vec3{px, py, pz},
		rotation: mgl32.Quat{
			W: qw,
			V: mgl32.Vec3{qx, qy, qz},
		},
		scale:  mgl32.Vec3{sx, sy, sz},
		parent: nil,
		//children: mapset.NewSet[*Transform](),
	}
}

func NewTranslateTransform(v *d4.DT_VECTOR3D) *Transform {
	return NewTransform(v.X, v.Y, v.Z, 0, 0, 0, 1, 1, 1, 1)
}

func NewPRTransform(pr *d4.PRTransform) *Transform {
	return NewTransform(
		pr.Wp.X,
		pr.Wp.Y,
		pr.Wp.Z,
		pr.Q.X.Value,
		pr.Q.Y.Value,
		pr.Q.Z.Value,
		pr.Q.W.Value,
		1,
		1,
		1,
	)
}

func NewPRAndScaleTransform(pr *d4.PRTransform, scale *d4.DT_VECTOR3D) *Transform {
	return NewTransform(
		pr.Wp.X,
		pr.Wp.Y,
		pr.Wp.Z,
		pr.Q.X.Value,
		pr.Q.Y.Value,
		pr.Q.Z.Value,
		pr.Q.W.Value,
		scale.X,
		scale.Y,
		scale.Z,
	)
}

func NewPRSTransform(prs *d4.PRSTransform) *Transform {
	return NewTransform(
		prs.Wp.X,
		prs.Wp.Y,
		prs.Wp.Z,
		prs.Q.X.Value,
		prs.Q.Y.Value,
		prs.Q.Z.Value,
		prs.Q.W.Value,
		prs.VScale.X,
		prs.VScale.Y,
		prs.VScale.Z,
	)
}

// Add is technically not thread-safe.
func (t *Transform) Add(child *Transform) {
	//t.children.Add(child)
	child.parent = t
}

func (t *Transform) AddTranslate(v *d4.DT_VECTOR3D) (child *Transform) {
	child = NewTranslateTransform(v)
	t.Add(child)
	return child
}

func (t *Transform) AddPR(pr *d4.PRTransform) (child *Transform) {
	child = NewPRTransform(pr)
	t.Add(child)
	return child
}

func (t *Transform) AddPRAndScale(pr *d4.PRTransform, scale *d4.DT_VECTOR3D) (child *Transform) {
	child = NewPRAndScaleTransform(pr, scale)
	t.Add(child)
	return child
}

func (t *Transform) AddPRS(prs *d4.PRSTransform) (child *Transform) {
	child = NewPRSTransform(prs)
	t.Add(child)
	return child
}

func (t *Transform) Path(f func(t *Transform)) {
	curr := t
	for curr != nil {
		defer f(curr)
		curr = curr.parent
	}
}

func (t *Transform) LocalTranslationMatrix() mgl32.Mat4 {
	return mgl32.Mat4FromRows(
		mgl32.Vec4{1, 0, 0, t.position.X()},
		mgl32.Vec4{0, 1, 0, t.position.Y()},
		mgl32.Vec4{0, 0, 1, t.position.Z()},
		mgl32.Vec4{0, 0, 0, 1},
	)
}

func (t *Transform) LocalRotationMatrix() mgl32.Mat4 {
	return t.rotation.Mat4()
}

func (t *Transform) LocalScalingMatrix() mgl32.Mat4 {
	return mgl32.Mat4FromRows(
		mgl32.Vec4{t.scale.X(), 0, 0, 0},
		mgl32.Vec4{0, t.scale.Y(), 0, 0},
		mgl32.Vec4{0, 0, t.scale.Z(), 0},
		mgl32.Vec4{0, 0, 0, 1},
	)
}

func (t *Transform) LocalToWorldMatrix() mgl32.Mat4 {
	t.lwMu.RLock()

	if t.lwMatrix != nil {
		defer t.lwMu.RUnlock()
		return *t.lwMatrix
	}

	t.lwMu.RUnlock()
	t.lwMu.Lock()
	defer t.lwMu.Unlock()

	m := t.LocalTranslationMatrix().
		Mul4(t.LocalRotationMatrix()).
		Mul4(t.LocalScalingMatrix())

	if t.parent != nil {
		m = t.parent.LocalToWorldMatrix().Mul4(m)
	}

	t.lwMatrix = &m
	return m
}

func (t *Transform) GetWorldPosition() mgl32.Vec3 {
	if t.parent == nil {
		return t.position
	}

	m := t.LocalToWorldMatrix()
	return m.Col(3).Vec3()
}

func (t *Transform) GetRelWorldPos(v *d4.DT_VECTOR3D) mgl32.Vec3 {
	m := t.LocalToWorldMatrix()
	return mgl32.Vec3{
		v.X*m.At(0, 0) + v.Y*m.At(0, 1) + v.Z*m.At(0, 2),
		v.X*m.At(1, 0) + v.Y*m.At(1, 1) + v.Z*m.At(1, 2),
		v.X*m.At(2, 0) + v.Y*m.At(2, 1) + v.Z*m.At(2, 2),
	}
}
