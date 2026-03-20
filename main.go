package main

import (
	"fmt"
	"os"
	"zcfgcli/cmd"
)

func main() {
	rootCmd := cmd.NewRoot()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
