package pattern

/*
	Реализовать паттерн «стратегия».
Объяснить применимость паттерна, его плюсы и минусы, а также реальные примеры использования данного примера на практике.
	https://en.wikipedia.org/wiki/Strategy_pattern
*/

import "fmt"

// based on
// https://levelup.gitconnected.com/the-strategy-pattern-in-go-2072d2b9d6ae

type weapon interface {
	useWeapon(opponent *Character)
}

type sword struct {
	damage int
	name   string
}

func NewSword(name string, damage int) *sword {
	return &sword{
		name:   name,
		damage: damage,
	}
}

func (s *sword) useWeapon(c *Character) {
	fmt.Printf("slashes the %s with a %s!\n", c.name, s.name)
	c.health -= s.damage
}

type bow struct {
	damage int
	name   string
}

func NewBow(name string, damage int) *bow {
	return &bow{
		name:   name,
		damage: damage,
	}
}

func (b *bow) useWeapon(c *Character) {
	fmt.Printf("shoots the %s with a %s!\n", c.name, b.name)
	c.health -= b.damage
}

type Character struct {
	health int
	weapon weapon
	name   string
}

func NewCharacter(name string) *Character {
	return &Character{
		name:   name,
		health: 100,
	}
}

func (c *Character) EquipWeapon(w weapon) {
	c.weapon = w
}

func (c *Character) Attack(opponent *Character) {
	fmt.Printf("The %s ", c.name)
	c.weapon.useWeapon(opponent)
}

func printCharacterStats(c *Character) {
	fmt.Printf("The %s has %d health left.\n", c.name, c.health)
}

func DoStrategy() {
	godSword := NewSword("Armadyl Godsword", 45)
	darkBow := NewBow("Dark Bow", 35)
	giantSword := NewSword("Giant's Sword", 55)

	champion := NewCharacter("Champion")
	champion.EquipWeapon(godSword)
	troll := NewCharacter("Cave Troll")
	troll.EquipWeapon(giantSword)

	printCharacterStats(champion)
	printCharacterStats(troll)

	champion.Attack(troll)
	printCharacterStats(troll)

	troll.Attack(champion)
	printCharacterStats(champion)

	champion.EquipWeapon(darkBow)
	champion.Attack(troll)
	printCharacterStats(troll)
}
