package pattern

/*
	Реализовать паттерн «цепочка вызовов».
Объяснить применимость паттерна, его плюсы и минусы, а также реальные примеры использования данного примера на практике.
	https://en.wikipedia.org/wiki/Chain-of-responsibility_pattern
*/

import "fmt"

type Step interface {
	Run(*Customer)
	SetNextStep(Step)
}

type Customer struct {
	name           string
	isHighPriority bool
}

type VoiceAssistant struct {
	next Step
}

func (v *VoiceAssistant) Run(cust *Customer) {
	fmt.Println("[Voice Assistant] Serving the customer:", cust.name)
	v.next.Run(cust)
}

func (v *VoiceAssistant) SetNextStep(next Step) {
	v.next = next
}

type Associate struct {
	next Step
}

func (a *Associate) Run(cust *Customer) {
	if cust.isHighPriority {
		fmt.Println("Redirecting customer directly to manager")
		a.next.Run(cust)
		return
	}
	fmt.Println("[Associate] Serving the customer:", cust.name)
	a.next.Run(cust)
}

func (a *Associate) SetNextStep(next Step) {
	a.next = next
}

type Manager struct {
	next Step
}

func (m *Manager) Run(cust *Customer) {
	fmt.Println("[Manager] Serving the customer:", cust.name)
}

func (m *Manager) SetNextStep(next Step) {
	m.next = next
}

func DoChainOfResponsibility() {
	m := &Manager{}

	assoc := &Associate{}
	assoc.SetNextStep(m)

	va := &VoiceAssistant{}
	va.SetNextStep(assoc)

	regular := &Customer{
		name: "Bob",
	}

	va.Run(regular)

	fmt.Println("------------")

	highPriority := &Customer{
		name:           "Alice",
		isHighPriority: true,
	}

	va.Run(highPriority)
}
