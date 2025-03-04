package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

func getChangedFiles() ([]string, error) {
	cmd := exec.Command("git", "diff", "--cached", "--name-only", "--diff-filter=ACM", "*.cy.js")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error getting changed files: %v", err)
	}

	files := strings.Fields(string(output))
	return files, nil
}

func checkCypressUnsafeChaining(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading file %s: %v", filePath, err)
	}

	unsafeChainRegex := regexp.MustCompile(`\.(clear|click|check|type)\(.*\)\.\w+`)
	safeChainRegex := regexp.MustCompile(`\.(clear|click|check|type)\(.*\)\.(then|and)`)

	matches := unsafeChainRegex.FindAll(content, -1)
	for _, match := range matches {
		if !safeChainRegex.Match(match) {
			return errors.New("unsafe chain detected")
		}
	}

	return nil
}

func main() {
	changedFiles, err := getChangedFiles()
	if err != nil {
		log.Fatal(err)
	}

	errFiles := ""
	for _, file := range changedFiles {
		err := checkCypressUnsafeChaining(file)
		if err != nil {
			errFiles += file + "\n"
		}
	}

	if errFiles != "" {
		fmt.Fprintln(os.Stderr, fmt.Sprintf(`Unsafe chained command found in:

%s
It is unsafe to chain further commands after clear(), click(), check() or type(), 

Bad:
cy.get('#selector').clear().type('value')

Good:
cy.get('#selector').clear();
cy.get('#selector').type('value')`, errFiles))
		os.Exit(1)
	}
	os.Exit(0)
}
