package main

import (
	"os"

	"github.com/ContainerSolutions/helm-convert/cmd"
	"github.com/golang/glog"
)

func main() {
	defer glog.Flush()

	if err := cmd.NewConvertCommand().Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
