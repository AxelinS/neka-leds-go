package screen

type Point struct{ X, Y int }

type SideCount struct {
	Top, Right, Bottom, Left int
}

type PixelLine struct {
	Pixels []Point // coordenadas reales de pantalla
}

func CountSides(points []Point, w, h, padding int) SideCount {
	var c SideCount

	for _, p := range points {
		switch {
		case p.Y <= padding:
			c.Top++
		case p.X >= w-padding:
			c.Right++
		case p.Y >= h-padding:
			c.Bottom++
		case p.X <= padding:
			c.Left++
		}
	}
	return c
}

func RectanglePerimeterPoints(w, h, n, padding int) []Point {
	x0, y0 := padding, padding
	x1, y1 := w-padding, h-padding

	top := x1 - x0
	right := y1 - y0
	perimeter := 2 * (top + right)
	step := float64(perimeter) / float64(n)

	points := make([]Point, 0, n)
	dist := 0.0

	for range n {
		d := dist
		var x, y int
		switch {
		case d < float64(top):
			x, y = x0+int(d), y0
		case d < float64(top+right):
			d -= float64(top)
			x, y = x1, y0+int(d)
		case d < float64(2*top+right):
			d -= float64(top + right)
			x, y = x1-int(d), y1
		default:
			d -= float64(2*top + right)
			x, y = x0, y1-int(d)
		}
		points = append(points, Point{x, y})
		dist += step
	}
	return points
}

func RectanglePerimeterPointsCinema(
	w, h, n int,
	padding int,
	paddingCinema int,
) []Point {
	// padding base
	x0 := padding
	x1 := w - padding
	// padding cine SOLO en Y
	y0 := padding + paddingCinema
	y1 := h - padding - paddingCinema

	top := x1 - x0
	right := y1 - y0
	perimeter := 2 * (top + right)
	step := float64(perimeter) / float64(n)

	points := make([]Point, 0, n)
	dist := 0.0
	for range n {
		d := dist
		var x, y int
		switch {
		case d < float64(top):
			x, y = x0+int(d), y0
		case d < float64(top+right):
			d -= float64(top)
			x, y = x1, y0+int(d)
		case d < float64(2*top+right):
			d -= float64(top + right)
			x, y = x1-int(d), y1
		default:
			d -= float64(2*top + right)
			x, y = x0, y1-int(d)
		}
		points = append(points, Point{x, y})
		dist += step
	}

	return points
}

func ApplyCinemaPadding(
	points []Point,
	w, h int,
	padding int,
	paddingCinema int,
) []Point {
	yBottomLimit := h - paddingCinema
	out := make([]Point, len(points))
	for i, p := range points {
		pp := p
		// demasiado arriba -> bajar
		if p.Y < paddingCinema {
			pp.Y = paddingCinema + padding/4
		}
		// demasiado abajo -> subir
		if p.Y > yBottomLimit {
			pp.Y = yBottomLimit - padding/4
		}
		out[i] = pp
	}
	return out
}
