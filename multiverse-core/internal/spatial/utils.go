// pkg/spatial/utils.go

package spatial

import "math"

// Buffer добавляет буфер к полигону (упрощённо — через bounding box).
func (poly *Polygon) Buffer(distance float64) Polygon {
	if len(*poly) == 0 {
		return *poly
	}

	// Находим bounding box
	minX, minY := (*poly)[0].X, (*poly)[0].Y
	maxX, maxY := (*poly)[0].X, (*poly)[0].Y
	for _, p := range *poly {
		if p.X < minX {
			minX = p.X
		}
		if p.Y < minY {
			minY = p.Y
		}
		if p.X > maxX {
			maxX = p.X
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}

	// Возвращаем новый прямоугольник с буфером
	return Polygon{
		{X: minX - distance, Y: minY - distance},
		{X: maxX + distance, Y: minY - distance},
		{X: maxX + distance, Y: maxY + distance},
		{X: minX - distance, Y: maxY + distance},
	}
}

// Contains проверяет, содержится ли точка в прямоугольнике.
func (poly *Polygon) Contains(p Point) bool {
	if len(*poly) != 4 {
		return false // поддерживаем только прямоугольники
	}
	minX := min((*poly)[0].X, (*poly)[2].X)
	maxX := max((*poly)[0].X, (*poly)[2].X)
	minY := min((*poly)[0].Y, (*poly)[2].Y)
	maxY := max((*poly)[0].Y, (*poly)[2].Y)
	return p.X >= minX && p.X <= maxX && p.Y >= minY && p.Y <= maxY
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// DistanceBetween вычисляет евклидово расстояние.
func DistanceBetween(a, b Point) float64 {
	dx := a.X - b.X
	dy := a.Y - b.Y
	return math.Sqrt(dx*dx + dy*dy)
}
