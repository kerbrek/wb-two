package pattern

/*
	Реализовать паттерн «строитель».
Объяснить применимость паттерна, его плюсы и минусы, а также реальные примеры использования данного примера на практике.
	https://en.wikipedia.org/wiki/Builder_pattern
*/

import (
	"fmt"
	"strings"
)

type sub struct {
	bread     string
	hasCheese bool
	toppings  []string
	sauces    []string
}

type iSubBuilder interface {
	setBread()
	setCheese()
	setToppings()
	setSauces()
	getSub() sub
}

type veggieDelightBuilder struct {
	sub
}

func (v *veggieDelightBuilder) setBread() {
	v.sub.bread = "parmesan oregano"
}

func (v *veggieDelightBuilder) setCheese() {
	v.sub.hasCheese = false
}

func (v *veggieDelightBuilder) setToppings() {
	v.sub.toppings = []string{"olives", "tomatoes", "onions", "jalapeños"}
}

func (v *veggieDelightBuilder) setSauces() {
	v.sub.sauces = []string{"south west"}
}

func (v *veggieDelightBuilder) getSub() sub {
	return v.sub
}

type chickenTeriyakiBuilder struct {
	sub
}

func (c *chickenTeriyakiBuilder) setBread() {
	c.sub.bread = "italian"
}

func (c *chickenTeriyakiBuilder) setCheese() {
	c.sub.hasCheese = true
}

func (c *chickenTeriyakiBuilder) setToppings() {
	c.sub.toppings = []string{"roasted chicken", "olives", "onions", "jalapeños"}
}

func (c *chickenTeriyakiBuilder) setSauces() {
	c.sub.sauces = []string{"chilli", "bbq"}
}

func (c *chickenTeriyakiBuilder) getSub() sub {
	return c.sub
}

type director struct {
	builder iSubBuilder
}

func (d *director) setBuilder(builder iSubBuilder) {
	d.builder = builder
}

func (d *director) buildSub() sub {
	d.builder.setBread()
	d.builder.setCheese()
	d.builder.setToppings()
	d.builder.setSauces()

	return d.builder.getSub()
}

func describeSub(sub sub) {
	fmt.Printf(
		"bread: %s, cheese: %t, toppings: %s, sauces: %s\n",
		sub.bread,
		sub.hasCheese,
		strings.Join(sub.toppings, ", "),
		strings.Join(sub.sauces, ", "),
	)
}

func DoBuilder() {
	veggieDelight := &veggieDelightBuilder{}
	director := &director{
		builder: veggieDelight,
	}
	veggieDelightSub := director.buildSub()
	describeSub(veggieDelightSub)

	fmt.Println("------------")

	director.setBuilder(&chickenTeriyakiBuilder{})
	chickenTeriyakiSub := director.buildSub()
	describeSub(chickenTeriyakiSub)
}
