package main

import (
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

func checkCypressClearChaining(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading file %s: %v", filePath, err)
	}

	clearChainRegex := regexp.MustCompile(`\.clear\(\)\s*\.`)

	if clearChainRegex.Match(content) {
		return fmt.Errorf(`Chained .clear() command found in %s
Cypress .clear() should not be chained with other commands.
Example of invalid code:
cy.get('#selector').clear().type('value')

Correct usage:
cy.get('#selector').clear();
cy.get('#selector').type('value')`, filePath)
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
		err := checkCypressClearChaining(file)
		if err != nil {
			errFiles += file + "\n"
		}
	}

	if errFiles != "" {
		fmt.Fprintln(os.Stderr, fmt.Sprintf(`Chained .clear() command found in:

%s

Cypress .clear() is unsafe to chain with other commands.

Bad:
cy.get('#selector').clear().type('value')

Good:
cy.get('#selector').clear();
cy.get('#selector').type('value')`, errFiles))
		os.Exit(1)
	}
	os.Exit(0)
}
