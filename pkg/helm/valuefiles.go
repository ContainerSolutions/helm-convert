package helm

import (
	"fmt"
	"strings"
)

// ValueFiles type
type ValueFiles []string

// String convert a pointer into a string
func (v *ValueFiles) String() string {
	return fmt.Sprint(*v)
}

// Type return the string ValueFiles
func (v *ValueFiles) Type() string {
	return "ValueFiles"
}

// Set split coma separated values into a set of values
func (v *ValueFiles) Set(value string) error {
	for _, filePath := range strings.Split(value, ",") {
		*v = append(*v, filePath)
	}
	return nil
}
