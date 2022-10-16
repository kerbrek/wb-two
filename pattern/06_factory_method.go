package pattern

/*
	Реализовать паттерн «фабричный метод».
Объяснить применимость паттерна, его плюсы и минусы, а также реальные примеры использования данного примера на практике.
	https://en.wikipedia.org/wiki/Factory_method_pattern
*/

import (
	"fmt"
	"log"
)

// based on
// https://github.com/shubhamzanwar/design-patterns/tree/master/1-factory

type Pet interface {
	GetName() string
	GetSound() string
	GetAge() int
}

type pet struct {
	name  string
	age   int
	sound string
}

func (p *pet) GetName() string {
	return p.name
}

func (p *pet) GetSound() string {
	return p.sound
}

func (p *pet) GetAge() int {
	return p.age
}

type dog struct {
	pet
}

type cat struct {
	pet
}

type PetType int

const (
	Cat PetType = iota
	Dog
)

func GetPet(petType PetType, name string, age int) Pet {
	switch petType {
	case Dog:
		return &dog{
			pet{
				name:  name,
				age:   age,
				sound: "bark",
			},
		}
	case Cat:
		return &cat{
			pet{
				name:  name,
				age:   age,
				sound: "meow",
			},
		}
	default:
		log.Fatalln("Unknown Pet Type")
		return nil
	}
}

func describePet(pet Pet) string {
	return fmt.Sprintf("%s is %d years old. It's sound is %s", pet.GetName(), pet.GetAge(), pet.GetSound())
}

func DoFactoryMethod() {
	dog := GetPet(Dog, "Chester", 2)
	fmt.Println(describePet(dog))

	fmt.Println("-------------")

	cat := GetPet(Cat, "Mr. Buttons", 3)
	fmt.Println(describePet(cat))
}
