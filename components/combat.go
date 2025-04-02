package components

type Combat interface {
	Health() float64
	AttackPower() float64
	Damage(amount float64)
}

type BasicCombat struct {
	health      float64
	attackPower float64
}

func NewBasicCombat(health, attackPower float64) *BasicCombat {
	return &BasicCombat{
		health,
		attackPower,
	}
}

func (b *BasicCombat) AttackPower() float64 {
	return b.attackPower
}
func (b *BasicCombat) Health() float64 {
	return b.health
}

func (b *BasicCombat) Damage(amount float64) {
	b.health -= amount
}

var _ Combat = (*BasicCombat)(nil)
