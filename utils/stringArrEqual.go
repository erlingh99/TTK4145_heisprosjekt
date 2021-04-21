package utils


func StringArrEqual(a []string, b []string) bool {
	for _,v := range a {
		if !Contains(b, v) {
			return false
		}
	}
	return true
}