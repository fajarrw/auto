package terminal

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// runCommand executes a command with arguments and prints its output
func RunCommand(folderPath string, name string, arg ...string) {
	cmd := exec.Command(name, arg...)
	cmd.Dir = folderPath
	output, err := cmd.CombinedOutput()
	fmt.Printf("Running command: %s %s\n", name, strings.Join(arg, " "))
	fmt.Println("Output:", string(output))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running command %s: %v\n", name, err)
		return
	}
}
