// pkg/spatial/filter.go

package spatial

// IsInScope проверяет, попадает ли точка в область.
func (vs *VisibilityScope) IsInScope(p Point) bool {
	if vs.Override != nil {
		return vs.Override(p)
	}
	if vs.Radius > 1e8 { // бесконечность
		return true
	}
	if vs.Polygon != nil {
		return vs.Polygon.Contains(p)
	}
	dx := p.X - vs.Center.X
	dy := p.Y - vs.Center.Y
	return dx*dx+dy*dy <= vs.Radius*vs.Radius
}

// FilterByScope фильтрует точки по области.
func FilterByScope(points []Point, vs VisibilityScope) []Point {
	var result []Point
	for _, p := range points {
		if vs.IsInScope(p) {
			result = append(result, p)
		}
	}
	return result
}
