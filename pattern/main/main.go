package main

import (
	"fmt"
	"pattern"
)

func main() {
	fmt.Println("---=== Facade ===---")
	pattern.DoFacade()
	fmt.Println()

	fmt.Println("---=== Builder ===---")
	pattern.DoBuilder()
	fmt.Println()

	fmt.Println("---=== Visitor ===---")
	pattern.DoVisitor()
	fmt.Println()

	fmt.Println("---=== Command ===---")
	pattern.DoCommand()
	fmt.Println()

	fmt.Println("---=== Chain of Responsibility ===---")
	pattern.DoChainOfResponsibility()
	fmt.Println()

	fmt.Println("---=== Factory Method ===---")
	pattern.DoFactoryMethod()
	fmt.Println()

	fmt.Println("---=== Strategy ===---")
	pattern.DoStrategy()
	fmt.Println()

	fmt.Println("---=== State ===---")
	pattern.DoState()
	fmt.Println()
}
