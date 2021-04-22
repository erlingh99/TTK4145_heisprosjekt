package utils

func Remove(s []string, e string) []string {
	for i, a := range s {
        if a == e {
			s[i] = s[len(s)-1]
			return s[:len(s)-1]
        }
    }
    return s
}