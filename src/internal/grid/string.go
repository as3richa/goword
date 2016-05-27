package grid

func (g Grid) String() string {
	result := ""
	for _, row := range g {
		for i, face := range row {
			if i != 0 {
				result += " "
			}
			if len(face) < 2 {
				result += " "
			}
			result += face
		}
		result += "\n"
	}
	return result
}
