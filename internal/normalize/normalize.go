package normalize

import "strings"

func SQL(input string) string {
	return strings.Join(strings.Fields(input), " ")
}
