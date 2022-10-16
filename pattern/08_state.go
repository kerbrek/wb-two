package pattern

/*
	Реализовать паттерн «состояние».
Объяснить применимость паттерна, его плюсы и минусы, а также реальные примеры использования данного примера на практике.
	https://en.wikipedia.org/wiki/State_pattern
*/

import (
	"errors"
	"fmt"
	"math/rand"
	"time"
)

// based on
// https://github.com/crazybber/awesome-patterns/blob/master/behavioral/state/problem.go

type playerState interface {
	heal(p *Player) error
	hurt(p *Player, damage int)
}

type healthyState struct{}

func (h healthyState) heal(p *Player) error {
	return nil
}

func (h healthyState) hurt(p *Player, damage int) {
	if damage > 0 && damage < p.maxHealth {
		p.health -= damage
		p.state = woundedState{}
	} else if damage > p.maxHealth {
		p.health = 0
		p.state = deadState{}
	}
}

type woundedState struct{}

func (woundedState) heal(p *Player) error {
	if p.health >= p.maxHealth-5 {
		fmt.Printf("healing from %d to %d\n", p.health, p.maxHealth)
		p.state = healthyState{}
		p.health = p.maxHealth
	} else {
		fmt.Printf("healing from %d to %d\n", p.health, p.health+5)
		p.health += 5
	}
	return nil
}

func (h woundedState) hurt(p *Player, damage int) {
	if p.health > damage {
		p.health -= damage
	} else {
		p.state = deadState{}
		p.health = 0
	}
}

type deadState struct{}

func (deadState) heal(P *Player) error {
	return errors.New("you are dead")
}

func (deadState) hurt(P *Player, damage int) {}

type Player struct {
	health    int
	maxHealth int
	state     playerState
}

func (p *Player) Heal() error {
	return p.state.heal(p)
}

func (p *Player) Hurt(damage int) {
	fmt.Printf("damage %d\n", damage)
	p.state.hurt(p, damage)
}

func DoState() {
	player := &Player{
		health:    100,
		maxHealth: 100,
		state:     healthyState{},
	}

	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 10; i++ {
		if err := player.Heal(); err != nil {
			fmt.Println(err)
			break
		}
		player.Hurt(rand.Intn(30))
	}
}
