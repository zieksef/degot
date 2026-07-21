package output

import (
	"fmt"
)

const (
	Prefix = "[degot]: "
)

func Sprintf(format string, args ...any) string {
	return fmt.Sprintf(Prefix+format, args...)
}
