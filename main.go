package main

import (
	"github.com/heguangyu1989/celo/cmd"
	"github.com/sirupsen/logrus"
)

func main() {
	if err := cmd.Execute(); err != nil {
		logrus.Fatal("command run fail : ", err)
	}
}
