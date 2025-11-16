// pkg/spatial/geometry.go

package spatial

import "math"

// Point — 2D точка.
type Point struct {
	X, Y float64
}

// Polygon — замкнутый полигон (первый == последний не обязателен).
type Polygon []Point

// BoundingBox — ограничивающий прямоугольник.
type BoundingBox struct {
	Min, Max Point
}

// Circle — круговая область.
type Circle struct {
	Center Point
	Radius float64
}

// Geometry — универсальная геометрия сущности.
type Geometry struct {
	Point       *Point
	Circle      *Circle
	Polygon     *Polygon
	BoundingBox *BoundingBox
	Custom      interface{} // Для сложных форм (в будущем)
}

// Center возвращает центр масс геометрии.
func (g *Geometry) Center() Point {
	switch {
	case g.Point != nil:
		return *g.Point
	case g.Circle != nil:
		return g.Circle.Center
	case g.Polygon != nil && len(*g.Polygon) > 0:
		return centroid(*g.Polygon)
	case g.BoundingBox != nil:
		return Point{
			X: (g.BoundingBox.Min.X + g.BoundingBox.Max.X) / 2,
			Y: (g.BoundingBox.Min.Y + g.BoundingBox.Max.Y) / 2,
		}
	default:
		return Point{0, 0}
	}
}

// MaxRadius возвращает максимальное расстояние от центра до границы.
func (g *Geometry) MaxRadius() float64 {
	center := g.Center()
	switch {
	case g.Point != nil:
		return 0
	case g.Circle != nil:
		return g.Circle.Radius
	case g.Polygon != nil && len(*g.Polygon) > 0:
		maxR := 0.0
		for _, p := range *g.Polygon {
			dx := p.X - center.X
			dy := p.Y - center.Y
			r := math.Sqrt(dx*dx + dy*dy)
			if r > maxR {
				maxR = r
			}
		}
		return maxR
	case g.BoundingBox != nil:
		dx := g.BoundingBox.Max.X - center.X
		dy := g.BoundingBox.Max.Y - center.Y
		return math.Sqrt(dx*dx + dy*dy)
	default:
		return 0
	}
}

// centroid вычисляет центр масс полигона (упрощённо — среднее).
func centroid(poly Polygon) Point {
	if len(poly) == 0 {
		return Point{0, 0}
	}
	sumX, sumY := 0.0, 0.0
	for _, p := range poly {
		sumX += p.X
		sumY += p.Y
	}
	n := float64(len(poly))
	return Point{X: sumX / n, Y: sumY / n}
}
