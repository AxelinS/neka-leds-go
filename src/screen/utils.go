package screen

func GetInnerOuterVals(width, height, argLeds, padding, lineLen, startPoint int) ([]Point, []Point) {
	inner := RectanglePerimeterPoints(width, height, argLeds, padding+lineLen, startPoint)
	outer := RectanglePerimeterPoints(width, height, argLeds, padding, startPoint)
	return inner, outer
}
