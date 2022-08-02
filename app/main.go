package main

import (
	"fmt"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/fake"
)

func Hello() string {
	return "Hello, world!"
}

func main() {
	fmt.Println(Hello())
	fmt.Println(fake.GoodBye())
}
