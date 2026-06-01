package main

import "fmt"

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
	OrganToDestroy
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
func (m Medicine) MedicineFunc() {}
func (m Medicine) AddonFunc() {}


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
func (v Virus) VirusFunc() {}
func (v Virus) AddonFunc() {}

type ApplicableToOrgan interface {
	Card
	GetColor() Color
	AddonFunc()
}

type Joint struct {
	Base Organ
	Added []ApplicableToOrgan
	Owner Player //added to be able to acces player from joint
}
func (j Joint) DeriveJointState() JointState {
	if len(j.Added) == 0 { return FreeJointState}
	if len(j.Added) == 1 {
		if j.Added[0].GetCardType() == MedicineCardType { return VaccinatedJointState }
		return InfectedJointState
	}
	// len(j.Added) == 2 
	return VaccinatedJointState 
}

type Treatment struct {
	TreatmentType TreatmentType
}
func (t Treatment) GetTreatmentType() TreatmentType { return t.TreatmentType }
func (t Treatment) GetCardType() CardType { return TreatmentCardType }

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
func (g *Game) ApplyAddonToJoint(addon ApplicableToOrgan, joint *Joint) OperationResult {
	baseMatches := ColorsMatch(addon.GetColor(), joint.Base.Color)
	pureOrgan := joint.DeriveJointState() == FreeJointState
	addedMatches := !pureOrgan && ColorsMatch(addon.GetColor(), joint.Added[0].GetColor())

	if !baseMatches && !addedMatches {
		fmt.Printf("ERR: Addon color %v has been tried to apply to joint color %v", addon.GetColor(), joint.Base.Color)
		return Illegal
	}
	
	if joint.DeriveJointState() == ImmunisedJointState {
		fmt.Println("ERR: Tried to apply addon to immunised joint")
		return Illegal
	}

	if pureOrgan {
		joint.Added = []ApplicableToOrgan{addon}	
	} else {
		if addon.GetCardType() == MedicineCardType {
			if joint.DeriveJointState() == InfectedJointState {
				joint.Added = []ApplicableToOrgan{}
			} else { //VaccinatedJointState
				joint.Added = append(joint.Added, addon)
			}
		} else {
			if joint.DeriveJointState() == VaccinatedJointState {
				joint.Added = []ApplicableToOrgan{}
			} else { //InfectedJointState
				// remove joint
				return OrganToDestroy
			}
		}
	}
	return Success
}



func main() {

}
