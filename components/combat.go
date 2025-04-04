package components

type Combat interface {
	Health() float64
	AttackPower() float64
	Attacking() bool
	Attack() bool
	Update()
	Damage(amount float64)
}

type BasicCombat struct {
	health      float64
	attackPower float64
	attacking   bool
}

func NewBasicCombat(health, attackPower float64) *BasicCombat {
	return &BasicCombat{
		health,
		attackPower,
		false,
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

func (b *BasicCombat) Attacking() bool {
	return b.attacking
}

func (b *BasicCombat) Attack() bool {
	b.attacking = true
	return true
}

func (b *BasicCombat) Update() {
}

var _ Combat = (*BasicCombat)(nil)

type EnemyCombat struct {
	*BasicCombat
	attackCooldown  float64
	timeSinceAttack float64
}

func NewEnemyCombat(health, attackPower, attackCooldown float64) *EnemyCombat {
	return &EnemyCombat{
		NewBasicCombat(health, attackPower),
		attackCooldown,
		0,
	}
}

func (e *EnemyCombat) Attack() bool {
	if e.timeSinceAttack >= e.attackCooldown {
		e.attacking = true
		e.timeSinceAttack = 0
		return true
	}
	return false
}

func (e *EnemyCombat) Update() {
	e.timeSinceAttack += 1
}

type PlayerCombat struct {
	*BasicCombat
	attackCooldown  float64
	timeSinceAttack float64
}

func NewPlayerCombat(health, attackPower, attackCooldown float64) *PlayerCombat {
	return &PlayerCombat{
		NewBasicCombat(health, attackPower),
		attackCooldown,
		0,
	}
}

func (p *PlayerCombat) Attack() bool {
	if p.timeSinceAttack >= p.attackCooldown {
		p.attacking = true
		p.timeSinceAttack = 0
		return true
	}
	return false
}

func (p *PlayerCombat) Update() {
	p.timeSinceAttack += 1
}

func (p *PlayerCombat) Damage(amount, tempHP float64) {
	blockedDMG := tempHP - amount
	if blockedDMG < 0 {
		p.health -= blockedDMG
	}

}

var _ Combat = (*EnemyCombat)(nil)
