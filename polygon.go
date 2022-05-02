package main

import (
	"github.com/fogleman/gg"
	"math"
	"math/rand"
	"sort"
)

type Vec2 struct {
	x float64
	y float64
}

type Point struct {
	x float64
	y float64
}

type rgba struct {
	r uint8
	g uint8
	b uint8
	a uint8
}

type Polygon struct {
	color    rgba
	vertices []Point
}

func (poly *Polygon) AddPoint(point Point) {
	poly.vertices = append(poly.vertices, point)
	poly.removeSelfIntersect()
}

func (poly Polygon) GetCenter() Point {
	centerX := 0.0
	centerY := 0.0
	for _, vertex := range poly.vertices {
		centerX += vertex.x
		centerY += vertex.y
	}
	centerX /= float64(len(poly.vertices))
	centerY /= float64(len(poly.vertices))
	return Point{x: centerX, y: centerY}
}

func (poly *Polygon) removeSelfIntersect() {
	center := poly.GetCenter()
	sort.SliceStable(poly.vertices, func(i, j int) bool {
		baseVec := Vec2{
			x: poly.vertices[0].x - center.x,
			y: poly.vertices[0].y - center.y,
		}
		vec1 := Vec2{
			x: poly.vertices[i].x - center.x,
			y: poly.vertices[i].y - center.y,
		}
		vec2 := Vec2{
			x: poly.vertices[j].x - center.x,
			y: poly.vertices[j].y - center.y,
		}

		return angleBetweenVectors(baseVec, vec1) < angleBetweenVectors(baseVec, vec2)

	})
}

func angleBetweenVectors(vec1 Vec2, vec2 Vec2) float64 {
	angle := math.Atan2(vec2.y, vec2.x) - math.Atan2(vec1.y, vec1.x)
	return angle
}

func drawSolution(solution []Polygon) gg.Context {
	context := gg.NewContext(200, 200)
	// Set background
	context.DrawRectangle(0, 0, 200, 200)
	context.SetRGBA255(int(BG.r), int(BG.g), int(BG.b), int(BG.a))
	context.Fill()
	for _, poly := range solution {
		for _, vertex := range poly.vertices[1:] {
			context.LineTo(vertex.x, vertex.y)
		}

		context.LineTo(poly.vertices[0].x, poly.vertices[0].y)
		context.SetRGBA255(int(poly.color.r), int(poly.color.g), int(poly.color.b), int(poly.color.a))
		context.Fill()
	}
	return *context
}

func makePolygon() Polygon {
	r := rand.Intn(255)
	g := rand.Intn(255)
	b := rand.Intn(255)
	a := rand.Intn(255-75) + 75
	polygon := Polygon{color: rgba{
		r: uint8(r),
		g: uint8(g),
		b: uint8(b),
		a: uint8(a),
	}}

	sides := 3
	for i := 0; i < sides; i++ {
		x := rand.Intn(200)
		y := rand.Intn(200)
		polygon.vertices = append(polygon.vertices, Point{x: float64(x), y: float64(y)})
	}
	polygon.removeSelfIntersect()

	return polygon
}
