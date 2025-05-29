package shared

import (
	"fmt"
	"os/exec"
	"strings"
)

func GetChangedFiles() ([]string, error) {
	cmd := exec.Command("git", "diff", "--cached", "--name-only", "--diff-filter=ACM", "*.cy.js")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error getting changed files: %v", err)
	}

	files := strings.Fields(string(output))
	return files, nil
}
