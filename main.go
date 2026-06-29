package main

import (
	"fmt"
	"rand"
)

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

type Result struct {
	Type OperationResult
	Message string
	
}

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
}
func (j Joint) DeriveJointState() JointState {
	if len(j.Added) == 0 { return FreeJointState}
	if len(j.Added) == 1 {
		if j.Added[0].GetCardType() == MedicineCardType { return VaccinatedJointState }
		return InfectedJointState
	}
	// len(j.Added) == 2
	return ImmunisedJointState
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
func (g *Game) ApplyAddonToJoint(addon ApplicableToOrgan, joint *Joint) Result {
	baseMatches := ColorsMatch(addon.GetColor(), joint.Base.Color)
	pureOrgan := joint.DeriveJointState() == FreeJointState
	addedMatches := !pureOrgan && ColorsMatch(addon.GetColor(), joint.Added[0].GetColor())

	if !baseMatches && !addedMatches {
		return Result{Illegal, fmt.Sprintf("ERR: Addon color %v has been tried to apply to joint color %v", addon.GetColor(), joint.Base.Color)}
	}

	if joint.DeriveJointState() == ImmunisedJointState {
		return Result{Illegal, "ERR: Tried to apply addon to immunised joint"}
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
				return Result{OrganToDestroy, "Destroy organ" + string(&joint)}
			}
		}
	}
	return Result{Success, "Success"}
}
const StartingDeck [68]Card = [68]Card{
		Organ{Color: Multi}, 5*Organ{Color: Red}, 5*Organ{Color: Green}, 5*Organ{Color: Blue}, 5*Organ{Color: Yellow},
		Virus{Color: Multi}, 4*Virus{Color: Red}, 4*Virus{Color: Green}, 4*Virus{Color: Blue}, 4*Virus{Color: Yellow},
		4*Medicine{Color: Multi}, 4*Medicine{Color: Red}, 4*Medicine{Color: Green}, 4*Medicine{Color: Blue}, 4*Medicine{Color: Yellow},
		2*Treatment{TreatmentType: Contagion}, 3*Treatment{TreatmentType: OrganThief}, 3*Treatment{TreatmentType: Transplant}, Treatment{TreatmentType: LatexGlove}, Treatment{TreatmentType: MedicalError}
	}
func(g *Game) InitDeck() {
	g.GrabPile = StartingDeck
	g.GrabPile.Shuffle()
}
func(cards *[]Card)Shuffle() {
	shuffled := make([]Card, len(cards))
	copy(shuffled, cards)
	for i := len(shuffled) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}
	cards = &shuffled
}
func(g *Game) DrawCard(player *Player) Result {
	if len(g.GrabPile) == 0 {
		// Need to flip discardpile into grabpile
		g.GrabPile = g.DiscardPile[::-1]
		g.DiscardPile = []Card{}
	}
	if len(player.Hand) >= 3 {
		return Result{Illegal, "ERR: Tried to draw card but player already has 3 or more cards"}
	}
	card := g.GrabPile[len(g.GrabPile)-1]
	g.GrabPile = g.GrabPile[:len(g.GrabPile)-1]
	player.Hand = append(player.Hand, card)
	if g.DrawCard(player) == Illegal { //Probably a bad idea, but recursively tries to draw until it can't anymore, as there should be any state in which a player shuldn't draw until 3 cards
		return Result{Success, "Success"}
	}
}
func(g *Game) DealInitialHands(numberOfPlayers int) Result {
	// Assuming all players empty hands
	if numberOfPlayers < 2 || numberOfPlayers > 6 {
		return Result{Illegal, "ERR: Invalid number of players"}
	}
	g.Players = make([]Player, numberOfPlayers)
	for i := 0; i < numberOfPlayers; i++ {
		res := g.DrawCard(i)
		if res.Type != Success {
			return Result{Error, fmt.Sprintf("Result of type %v with message %v when dealing player %v", res.Type, res.Message, i)}
		}
	}
	return Result{Success, "Success"}
} 

func main() {

}

/* TODO:
[*] InitDeck / Shuffle / DrawCard / DealInitialHands
[ ] HasWon
[ ] Action struct
[ ] Regla de exclusividad de color en Body
[ ] PlayOrgan
[ ] PlayAddon  ←  aquí vive la destrucción de órgano
[ ] DiscardAndDraw
[ ] DrawPhase / PlayPhase / game loop
[ ] PrintGameState / PrintHand / input parser
[ ] 5 tratamientos
*/
