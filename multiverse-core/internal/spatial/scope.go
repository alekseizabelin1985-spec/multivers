// pkg/spatial/scope.go

package spatial

// VisibilityScope — динамическая область наблюдения.
type VisibilityScope struct {
	Center   Point
	Radius   float64
	Polygon  *Polygon
	Override func(eventPoint Point) bool // Кастомная логика (опционально)
}

// DefaultScope создаёт область по типу и параметрам.
func DefaultScope(entityType string, geometry *Geometry, config map[string]interface{}) VisibilityScope {
	baseRadius := 200.0
	switch entityType {
	case "player":
		if p, ok := config["perception"].(float64); ok {
			baseRadius = p * 200
		}
	case "group":
		baseRadius = 300 + geometry.MaxRadius()
	case "location", "region":
		baseRadius = geometry.MaxRadius() + 200
	default: // world, universe
		return VisibilityScope{Radius: 1e9}
	}

	center := geometry.Center()
	return VisibilityScope{
		Center: center,
		Radius: baseRadius,
	}
}

// Buffer расширяет область на заданное расстояние (метров).
func (vs *VisibilityScope) Buffer(distance float64) VisibilityScope {
	if vs.Polygon != nil {
		buffered := vs.Polygon.Buffer(distance)
		return VisibilityScope{Polygon: &buffered}
	}
	return VisibilityScope{
		Center: vs.Center,
		Radius: vs.Radius + distance,
	}
}
