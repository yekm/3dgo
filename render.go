// Copyright 2018 Erik van Zijst -- erik.van.zijst@gmail.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

import (
	"image"
	"math"
	"os"
	"time"

	"github.com/fogleman/gg"
)

type Renderer struct {
	a            *gg.Context
	model        Model
	cameraMatrix M4
	projector    Projector
	rotTime      float64 // seconds per rotation
}

func (r *Renderer) frame() image.Image {
	r.Draw(r.a)
	return r.a.Image()
}

func (r *Renderer) drawModel(dc *gg.Context, model *Model) {

	for _, t := range model.triangles {
		if Dot(t.v1, t.Normal()) < 0. &&
			// frustum near-plane clipping:
			t.v1.z <= r.projector.clipping && t.v2.z <= r.projector.clipping && t.v3.z <= r.projector.clipping {

			point := r.projector.project(t.v1)
			dc.LineTo(point.x, point.y)

			point = r.projector.project(t.v2)
			dc.LineTo(point.x, point.y)

			point = r.projector.project(t.v3)
			dc.LineTo(point.x, point.y)

			dc.SetLineWidth(1)
			dc.Stroke()
		}
	}
}

func (r *Renderer) Draw(dc *gg.Context) {
	angleZ := (float64(time.Now().UnixNano()%(int64(r.rotTime*1e9))) / 1e9) *
		((2 * math.Pi) / r.rotTime)
	angleY := (float64(time.Now().UnixNano()%(int64(r.rotTime*1e9/2))) / 1e9) *
		((2 * math.Pi) / (r.rotTime / 2))
	mat := RotX(math.Pi / 2.).Mul(RotY(angleY)).Mul(RotZ(angleZ))
	model := r.model.Clone().Apply(mat)

	dc.SetHexColor("000")
	dc.Clear()
	dc.SetRGBA(1, 1, 1, 1)

	r.drawModel(dc, model.Apply(r.cameraMatrix.Inverse()))
}

/*
func (r *Renderer) KeyEvent(a *ui.Area, ke *ui.AreaKeyEvent) (handled bool) {
	step := .25

	if !ke.Up {
		tm := new(M4).SetIdentity()

		switch ke.Key {
		case int32('w'):
			tm = TransM(NewV4(0, 0, -step))
		case int32('s'):
			tm = TransM(NewV4(0, 0, step))
		case int32('a'):
			tm = TransM(NewV4(-step, 0, 0))
		case int32('d'):
			tm = TransM(NewV4(step, 0, 0))
		}

		switch ke.ExtKey {
		case ui.Left:
			tm = RotY(rad(step * 4))
		case ui.Right:
			tm = RotY(rad(-step * 4))
		}
		r.cameraMatrix.Mul(tm)
		return true
	}
	return
}
*/
func rad(degrees float64) float64 {
	return degrees * math.Pi / 180
}

func Cube() *Model {
	top := &Model{[]Triangle{
		// counter-clockwise vertex winding
		*NewTriangle(.5, .5, .5, -.5, .5, .5, -.5, -.5, .5),
		*NewTriangle(-.5, -.5, .5, .5, -.5, .5, .5, .5, .5),
	}}

	cube := top.Merge(
		*top.Clone().Rot(rad(180), 0, 0), // bottom
		*top.Clone().Rot(rad(90), 0, 0),  // north
		*top.Clone().Rot(rad(-90), 0, 0), // south
		*top.Clone().Rot(0, rad(90), 0),  // west
		*top.Clone().Rot(0, rad(-90), 0), // east
	)
	return cube
}

func get_renderer(filename string, w, h int) Renderer {
	var model Model
	if filename != "" {
		f, err := os.Open(filename)
		if err != nil {
			panic(err)
		}
		model = *NewSTLReader(f).ReadModel(true)
		f.Close()
	} else {
		model = *Cube().Rot(math.Pi/4, math.Pi/4, math.Pi/4)
	}

	renderer := Renderer{
		a:            gg.NewContext(w, h),
		projector:    *NewProjector(h, 52),
		model:        model,
		cameraMatrix: *TransM(NewV4(0, 0, 2)),
		rotTime:      5, // seconds per full rotation
	}

	return renderer
}
