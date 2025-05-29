package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/ministryofjustice/opg-modernising-lpa/scripts/pre-commit/shared"
)

func checkCypressUnsafeChaining(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading file %s: %v", filePath, err)
	}

	unsafeChainRegex := regexp.MustCompile(`\.(clear|click|check|type|focus)\(.*\)\.\w+`)
	safeChainRegex := regexp.MustCompile(`\.(clear|click|check|type|focus)\(.*\)\.(then|and)`)

	matches := unsafeChainRegex.FindAll(content, -1)
	for _, match := range matches {
		if !safeChainRegex.Match(match) {
			return errors.New("unsafe chain detected")
		}
	}

	return nil
}

func main() {
	changedFiles, err := shared.GetChangedFiles()
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
It is unsafe to chain further commands after clear(), click(), check(), type() or focus(),

Bad:
cy.get('#selector').clear().type('value')

Good:
cy.get('#selector').clear();
cy.get('#selector').type('value')`, errFiles))
		os.Exit(1)
	}
	os.Exit(0)
}
