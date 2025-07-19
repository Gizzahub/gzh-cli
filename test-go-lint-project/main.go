package main

import "fmt"

func main() {
	message := "Hello, World!"
	fmt.Println(message)

	// Intentional issues for linter to catch
	unusedVar := "This is unused"

	if true {
		fmt.Println("This is always true")
	}
}

// Add a function with no error handling
func riskyFunction() {
	fmt.Printf("This might fail")
}
