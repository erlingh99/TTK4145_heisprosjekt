package utils

func StringArrDiff(a []string, b []string) []string {
	c := make([]string, 0)
	for _, val := range a {
		if !Contains(b, val) {
			c = append(c, val)
		}
	}
	return c
}