package screen

func GetInnerOuterVals(width, height, argLeds, padding, lineLen int) ([]Point, []Point) {
	inner := RectanglePerimeterPoints(width, height, argLeds, padding+lineLen)
	outer := RectanglePerimeterPoints(width, height, argLeds, padding)
	return inner, outer
}
