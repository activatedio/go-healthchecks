package cmd

import (
	"fmt"
	"os"
)

func CheckExit(err error) {
	if err != nil {
		fmt.Println("Error: " + err.Error())
		os.Exit(1)
	}
}
