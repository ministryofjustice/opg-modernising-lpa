package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/ministryofjustice/opg-modernising-lpa/scripts/pre-commit/shared"
)

func checkItOnly(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading file %s: %v", filePath, err)
	}

	regex := regexp.MustCompile(`it\.only`)
	if regex.Match(content) {
		return errors.New("spec contains it.only")
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
		err := checkItOnly(file)
		if err != nil {
			errFiles += file + "\n"
		}
	}

	if errFiles != "" {
		fmt.Fprintln(os.Stderr, fmt.Sprintf(`it.only found in:

%s
Remove all instances so the full test suites can run`, errFiles))
		os.Exit(1)
	}
	os.Exit(0)
}
