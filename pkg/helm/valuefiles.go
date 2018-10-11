package helm

import (
	"fmt"
	"strings"
)

type ValueFiles []string

func (v *ValueFiles) String() string {
	return fmt.Sprint(*v)
}

func (v *ValueFiles) Type() string {
	return "ValueFiles"
}

func (v *ValueFiles) Set(value string) error {
	for _, filePath := range strings.Split(value, ",") {
		*v = append(*v, filePath)
	}
	return nil
}
