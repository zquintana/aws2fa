package main

import (
	"aws2fa/cmd"
	"fmt"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
	}
}
