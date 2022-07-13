package pattern

/*
	Реализовать паттерн «комманда».
Объяснить применимость паттерна, его плюсы и минусы, а также реальные примеры использования данного примера на практике.
	https://en.wikipedia.org/wiki/Command_pattern
*/

import "fmt"

type Command interface {
	execute()
}

type Restaurant struct {
	TotalDishes   int
	CleanedDishes int
}

func NewRestaurant() *Restaurant {
	const totalDishes = 10
	return &Restaurant{
		TotalDishes:   totalDishes,
		CleanedDishes: totalDishes,
	}
}

func (r *Restaurant) MakePizza(n int) Command {
	return &MakePizzaCommand{
		restaurant: r,
		n:          n,
	}
}

func (r *Restaurant) MakeSalad(n int) Command {
	return &MakeSaladCommand{
		restaurant: r,
		n:          n,
	}
}

func (r *Restaurant) CleanDishes() Command {
	return &CleanDishesCommand{
		restaurant: r,
	}
}

type MakePizzaCommand struct {
	n          int
	restaurant *Restaurant
}

func (c *MakePizzaCommand) execute() {
	c.restaurant.CleanedDishes -= c.n
	fmt.Println("made", c.n, "pizzas")
}

type MakeSaladCommand struct {
	n          int
	restaurant *Restaurant
}

func (c *MakeSaladCommand) execute() {
	c.restaurant.CleanedDishes -= c.n
	fmt.Println("made", c.n, "salads")
}

type CleanDishesCommand struct {
	restaurant *Restaurant
}

func (c *CleanDishesCommand) execute() {
	c.restaurant.CleanedDishes = c.restaurant.TotalDishes
	fmt.Println("dishes cleaned")
}

type Cook struct {
	Commands []Command
}

func (c *Cook) executeCommands() {
	for _, c := range c.Commands {
		c.execute()
	}
}

func DoCommand() {
	r := NewRestaurant()

	tasks := []Command{
		r.MakePizza(2),
		r.MakeSalad(1),
		r.MakePizza(3),
		r.CleanDishes(),
		r.MakePizza(4),
		r.CleanDishes(),
	}

	cooks := []*Cook{{}, {}}

	for i, task := range tasks {
		cook := cooks[i%len(cooks)]
		cook.Commands = append(cook.Commands, task)
	}

	for i, c := range cooks {
		fmt.Printf("Cook %d:\n", i)
		c.executeCommands()
	}
}
