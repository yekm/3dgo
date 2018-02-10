package main

import (
	"github.com/andlabs/ui"
	"fmt"
	"time"
	"math"
)

type Renderer struct {
	a         *ui.Area
	dp        *ui.AreaDrawParams
	model     Model
	projector Projector
	rotTime   float64 // seconds per rotation
}

func (r *Renderer) mainLoop() {
	for {
		r.a.QueueRedrawAll()
		time.Sleep(time.Duration(20) * time.Millisecond)
	}
}

func (r *Renderer) drawModel(a *ui.Area, dp *ui.AreaDrawParams, model *Model) {

	path := ui.NewPath(ui.Winding)
	for _, t := range model.triangles {
		if Angle(t.v1, t.Normal()) > math.Pi / 2 {	// back-face culling
			point := r.projector.project(t.v1)
			path.NewFigure(point.x, point.y)

			point = r.projector.project(t.v2)
			path.LineTo(point.x, point.y)

			point = r.projector.project(t.v3)
			path.LineTo(point.x, point.y)
			path.CloseFigure()
		}
	}
	path.End()

	dp.Context.Stroke(path,
		&ui.Brush{A:1, Type:ui.Solid},
		&ui.StrokeParams{ui.FlatCap, ui.MiterJoin, 1, 2, nil, 1})
	path.Free()
}

func (r *Renderer) Draw(a *ui.Area, dp *ui.AreaDrawParams) {

	angle := (float64(time.Now().UnixNano() % (int64(r.rotTime * 1e9))) / 1e9) *
				((2 * math.Pi) / r.rotTime)

	r.drawModel(a, dp, r.model.Clone().Rot(angle, angle, angle).Move(0, 0, -2.5))
	r.drawModel(a, dp, r.model.Clone().Apply(ScaleM(.2, .2, .2)).Rot(-angle * 2, angle, -angle).Move(.8, -.8, -2.5))
}

func (r Renderer) MouseEvent(a *ui.Area, me *ui.AreaMouseEvent) {
	return
}

func (r Renderer) MouseCrossed(a *ui.Area, left bool) {
	return
}

func (r Renderer) DragBroken(a *ui.Area) {
	return
}

func (r Renderer) KeyEvent(a *ui.Area, ke *ui.AreaKeyEvent) (handled bool) {
	return
}

func rad(degrees float64) float64 {
	return degrees * math.Pi / 180
}

func Cube() (*Model) {
	top := (&Model{[]Triangle{
		// counter-clockwise vertex winding
		*NewTriangle(.5, .5, 0,  -.5, .5, 0,  -.5, -.5, 0),
		*NewTriangle(-.5, -.5, 0,  .5, -.5, 0,  .5, .5, 0),
	}}).Move(0, 0, .5)

	cube := top.Merge(
		*top.Clone().Rot(rad(180), 0, 0),	// bottom
		*top.Clone().Rot(rad(90), 0, 0),	// north
		*top.Clone().Rot(rad(-90), 0, 0),	// south
		*top.Clone().Rot(0, rad(90), 0),	// west
		*top.Clone().Rot(0, rad(-90), 0),	// east
	)
	fmt.Printf("%+v\n", cube)
	return cube
}

func main() {
	err := ui.Main(func() {

		renderer := Renderer{
			a:    nil,
			dp:        nil,
			projector: *NewProjector(600, 52),
			model: *Cube().Rot(math.Pi / 4, math.Pi / 4, math.Pi / 4),
			rotTime: 18,	// seconds per full rotation
		}
		canvas := ui.NewArea(&renderer)
		renderer.a = canvas

		box := ui.NewVerticalBox()
		box.Append(canvas, true)
		window := ui.NewWindow("Perspective Projection", renderer.projector.size, renderer.projector.size, false)
		window.SetMargined(false)
		window.SetChild(box)
		window.OnClosing(func(*ui.Window) bool {
			ui.Quit()
			return true
		})
		window.Show()

		go func () {
			renderer.mainLoop()
		}()
	})
	if err != nil {
		panic(err)
	}
}