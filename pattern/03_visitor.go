package pattern

/*
	Реализовать паттерн «посетитель».
Объяснить применимость паттерна, его плюсы и минусы, а также реальные примеры использования данного примера на практике.
	https://en.wikipedia.org/wiki/Visitor_pattern
*/

import "fmt"

// based on
// https://medium.com/@felipedutratine/visitor-design-pattern-in-golang-3c142a12945a

type Employee interface {
	FullName()
	Accept(Visitor)
}

type Developer struct {
	FirstName string
	LastName  string
	Salary    int
}

func (d Developer) FullName() {
	fmt.Println("Developer", d.FirstName, d.LastName)
}

func (d Developer) Accept(v Visitor) {
	v.VisitDeveloper(d)
}

type TeamLead struct {
	FirstName string
	LastName  string
	Salary    int
}

func (t TeamLead) FullName() {
	fmt.Println("TeamLead", t.FirstName, t.LastName)
}

func (t TeamLead) Accept(v Visitor) {
	v.VisitTeamLead(t)
}

type Visitor interface {
	VisitDeveloper(d Developer)
	VisitTeamLead(t TeamLead)
}

type IncomeCalc struct {
	bonusRate int
}

func (c IncomeCalc) VisitDeveloper(d Developer) {
	fmt.Println("Income:", d.Salary+d.Salary*c.bonusRate/100)
}

func (c IncomeCalc) VisitTeamLead(t TeamLead) {
	fmt.Println("Income:", t.Salary+t.Salary*c.bonusRate/100)
}

func DoVisitor() {
	developer := Developer{"John", "Smith", 3000}
	teamlead := TeamLead{"Alice", "Liddell", 5000}

	developer.FullName()
	developer.Accept(IncomeCalc{10})

	fmt.Println("------------")

	teamlead.FullName()
	teamlead.Accept(IncomeCalc{20})
}
