package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	cmd := exec.Command("golangci-lint", "run", "-c", ".golang-ci.yml")
	cmd.Dir = "/home/archmagece/myopen/Gizzahub/gzh-manager-go"

	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	fmt.Printf("Output:\n%s", output)
}
