package main

// --- Auxiliar Types
type CardType int
const(
	OrganCardType CardType = iota
	MedicineCardType
	VirusCardType
	TreatmentCardType
	CustomCardType
	NoneCardType
)

type TreatmentType int
const(
	Transplant TreatmentType = iota
	OrganThief
	Contagion
	LatexGlove
	MedicalError
	NotTreatment
)

type Color int
const(
	Red Color = iota
	Blue
	Yellow
	Green
	Multi
	NotColor
)

type JointState int
const(
	FreeJointState JointState = iota
	VaccinatedJointState
	ImmunisedJointState
	InfectedJointState
)

type OperationResult int
const(
	Success OperationResult = iota
	Error
	Illegal
	NotImplemented
)

// Every card must have these methods
type Card interface {
	GetCardType() CardType
}

func ColorsMatch(c1, c2 Color) bool {
	return (c1 == c2 || c1 == Multi || c2 == Multi) && !(c1 == NotColor || c2 == NotColor)  
}

// We define the specific characteriistics of a Medicine in order to be considered a Card
type Medicine struct {
	Color Color 
}
func (m Medicine) GetColor() Color { return m.Color }
func (m Medicine) GetCardType() CardType { return MedicineCardType }
func (m Medicine) ApplyToJoint(g Game, j *Joint) OperationResult {
	if !(ColorsMatch(j.Base.Color, m.Color) || ColorsMatch(j.Added[0].GetColor(), m.Color)) {return Illegal}
	if j.State == ImmunisedJointState {return Illegal}
	if len(j.Added) == 0 {
		j.Added = append(j.Added, m)
		j.State = VaccinatedJointState
		return Success
	}
	switch j.Added[0].GetCardType() {
	case VirusCardType: 
		g.Discard(j.Added[0])
		g.Discard(m)
		j.Added = []ApplicableToOrgan{}
		j.State = FreeJointState
	case MedicineCardType:
		j.Added = append(j.Added, m)
		j.State = ImmunisedJointState
	}
	return Success
}


type Organ struct {
	Color Color
}
func (o Organ) GetColor() Color { return o.Color }
func (o Organ) GetCardType() CardType { return OrganCardType }

type Virus struct {
	Color Color
}
func (v Virus) GetColor() Color { return v.Color }
func (v Virus) GetCardType() CardType { return VirusCardType }
func (v Virus) ApplyToJoint(g Game, j *Joint) OperationResult {
	if !(ColorsMatch(j.Base.Color, v.Color) || ColorsMatch(j.Added[0].GetColor(), v.Color)) {return Illegal}
	if j.State == ImmunisedJointState {return Illegal}
	if len(j.Added) == 0 {
		j.Added = append(j.Added, v)
		j.State = InfectedJointState
		return Success
	}
	switch j.Added[0].GetCardType() {
	case VirusCardType: 
		g.Discard(j.Base)
		g.Discard(j.Added[0])
		g.Discard(v)
		//TODO: remove Joint from player's body
		return NotImplemented
	case MedicineCardType:
		g.Discard(j.Added[0])
		g.Discard(v)
		j.Added = []ApplicableToOrgan{}
		j.State = FreeJointState
	}
	return Success
}

type ApplicableToOrgan interface {
	Card
	GetColor() Color
	ApplyToJoint(g Game, j *Joint) OperationResult
}

type Joint struct {
	Base Organ
	Added []ApplicableToOrgan
	State JointState
}

type Player struct {
	Hand []Card
	Body []Joint
}


type Game struct {
	Players []Player
	GrabPile []Card
	DiscardPile []Card
}

func (g *Game) Discard(c Card) {
	g.DiscardPile = append(g.DiscardPile, c)
}

func main() {
	
}