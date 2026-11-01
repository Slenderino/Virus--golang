package main

import (
	"fmt"
	"math/rand"
	"slices"
)

// --- Tipos auxiliares
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

// Toda carta debe tener este método
type Card interface {
	GetCardType() CardType
}

func ColorsMatch(c1, c2 Color) bool {
	return (c1 == c2 || c1 == Multi || c2 == Multi) && !(c1 == NotColor || c2 == NotColor)
}

// Que parámetros tiene toda medicina
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

func PlayerHasOrganColor(p Player, color Color) bool {
	for _, o := range p.Body {
		if o.Base.Color == color {
			return true
		}
	}
	return false
}

type Virus struct {
	Color Color
}
func (v Virus) GetColor() Color { return v.Color }
func (v Virus) GetCardType() CardType { return VirusCardType }
func (v Virus) VirusFunc() {}
func (v Virus) AddonFunc() {}

// Si una carta tiene color, y tiene la función etiqueta AddonFunc, entonces aplica como ApplicableToOrgan
type ApplicableToOrgan interface {
	Card
	GetColor() Color
	AddonFunc()
}

// Un Joint es el conjunto formado por el órgano base y cualquier medicina o virus (ApplicableToOrgan) que pueda tener encima
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

// El struct Game es el estado, alojando a los jugadores (manos y cuerpos), las pilas de cartas, y el turno actual
type Game struct {
	Players []Player
	GrabPile []Card
	DiscardPile []Card
	CurrentTurn int
}

func (g *Game) Discard(c Card) {
	g.DiscardPile = append(g.DiscardPile, c)
}

// Función atomizada, solo se muta el estado si la acción va a ser exitosa, si falla el estado queda intacto
func (g *Game) ApplyAddonToJoint(addon ApplicableToOrgan, joint *Joint) Result {
	baseMatches := ColorsMatch(addon.GetColor(), joint.Base.Color)
	pureOrgan := joint.DeriveJointState() == FreeJointState

	if !baseMatches && (pureOrgan || !(joint.Added[0].GetColor() == Multi)) {
		// Si el addon no es multicolor entonces es del mismo color de la base, y si la base no es compatible entonces ninguna de las cartas es compatible
		return Result{Illegal, fmt.Sprintf("ERR: Addon color %v has been tried to apply to joint color %v", addon.GetColor(), joint.Base.Color)}
	}

	if joint.DeriveJointState() == ImmunisedJointState {
		return Result{Illegal, "ERR: Tried to apply addon to immunised joint"}
	}

	// baseMatches O added es multicolor

	if pureOrgan {
		joint.Added = []ApplicableToOrgan{addon} // El órgano sí o sí tiene un color compatible con el addon
	} else {
		if addon.GetCardType() == MedicineCardType {
			if joint.DeriveJointState() == InfectedJointState {
				g.Discard(joint.Added[0])
				g.Discard(addon)
				joint.Added = []ApplicableToOrgan{}

			} else { //VaccinatedJointState
				if baseMatches {
					joint.Added = append(joint.Added, addon)
				} else {
					return Result{Illegal, "ERR: Tried to vaccinate incompatible organ"}
				}
			}
		} else { // VirusCardType
			if joint.DeriveJointState() == VaccinatedJointState {
				g.Discard(joint.Added[0])
				g.Discard(addon)
				joint.Added = []ApplicableToOrgan{}
			} else { //InfectedJointState
				if baseMatches {
					// hay que quitar el joint del cuerpo del jugador
					return Result{OrganToDestroy, fmt.Sprintf("Destroy organ %v", joint)}
				} else {
					return Result{Illegal, "ERR: Cannot infect incompatible organ"}
				}
			}
		}
	}
	return Result{Success, "Successfully placed addon onto joint"}
}

func BuildStartingDeck() []Card {
    deck := []Card{}

    // Cartas únicas
    deck = append(deck, Organ{Multi}, Virus{Multi}, Treatment{LatexGlove}, Treatment{MedicalError})
    // Dos cartas
    for range 2 {
        deck = append(deck, Treatment{Contagion})
    }
    // Tres cartas
    for range 3 {
        deck = append(deck, Treatment{OrganThief}, Treatment{Transplant})
    }

    for range 4 {
        deck = append(deck, Virus{Red}, Virus{Green}, Virus{Blue}, Virus{Yellow},
        Medicine{Multi}, Medicine{Red}, Medicine{Green}, Medicine{Blue}, Medicine{Yellow})
    }

    for range 5 {
        deck = append(deck, Organ{Color: Red}, Organ{Color: Green},
                            Organ{Color: Blue}, Organ{Color: Yellow})
    }
    return deck
}
func(g *Game) InitDeck() {
	g.GrabPile = BuildStartingDeck()
	Shuffle(&g.GrabPile)
}
func Shuffle(cards *[]Card) {
    s := *cards
    for i := len(s) - 1; i > 0; i-- {
        j := rand.Intn(i + 1)
        s[i], s[j] = s[j], s[i]
    }
}

/*
 * DrawCard le da una carta de g.GrabPile al jugador, si nos quedamos sin GrabPile le da vuelta a DiscardPile y la vuelve la nueva GrabPile
 *
 * Nos aseguramos primero de que GrabPile no esté vacío
 *
 */
func (g *Game) DrawCard(player *Player) Result {
    if len(g.GrabPile) == 0 {
        if len(g.DiscardPile) == 0 {
            return Result{Error, "ERR: Both piles are empty"}
        }
        for _, v := range slices.Backward(g.DiscardPile) {
            g.GrabPile = append(g.GrabPile, v)
        }
        g.DiscardPile = []Card{}
        // GrabPile no se mezcla, una interpretación mas literal del manual, en el que no menciona ningún barajeo
    }
    card := g.GrabPile[len(g.GrabPile)-1]
    g.GrabPile = g.GrabPile[:len(g.GrabPile)-1]
    player.Hand = append(player.Hand, card)
    return Result{Success, "Success"}
}

/*
 * DealInitialHands deals the initial hands to all players, drawing 3 cards each.
 * It can do this at a *Game level, not needing to access the individual player's hand, DrawCard
 * manages that responsability.
 *
 * It checks the number of players to ensure it's within the valid range (2-6).
 * Then instantiates them inside g.Players, drawing three cards for each, and propagating any error
 * occurred during the dealing.
 */
func(g *Game) DealInitialHands(numberOfPlayers int) Result {
	// Assuming all players empty hands
	if numberOfPlayers < 2 || numberOfPlayers > 6 {
		return Result{Illegal, "ERR: Invalid number of players"}
	}
	g.Players = make([]Player, numberOfPlayers)
	for i := range numberOfPlayers {
		for range 3 {
			res := g.DrawCard(&g.Players[i])
			if res.Type != Success {
				return Result{Error, fmt.Sprintf("Result of type %v with message %v when dealing player %v", res.Type, res.Message, i)}
		}
		}
	}
	return Result{Success, "Success"}
}

/*
 * HasWon checks if a player has won based on his number of joints and their states.
 * Assuming no colors are repeated, having 4 or more healthy joints would indicate a victory.
 *
 * We check the lenght of the body, and then, for each joint, we count the healthy ones, if healthy_joints >= 4 then true, else false.
 */
func HasWon(player Player) bool {
	if len(player.Body) >= 4 {
		var healthyJoints int
		for _, joint := range player.Body {
			if joint.DeriveJointState() != InfectedJointState {
				healthyJoints++
			}
		}
		if healthyJoints >= 4 { return true }
	}
	return false
}


type Action struct {
	EmisorIdx        int
	TargetPlayerIdx  []int
	EmisorJointIdx   []int
	TargetJointIdx   []int
	ActuatingCardIdx []int
	ActionFunc func(*Game, Action) Result
}

func PlayOrgan(g *Game, action Action) Result {
	emisor := &g.Players[action.EmisorIdx]

	if len(action.ActuatingCardIdx) != 1 {
		return Result{Error, "Can only play 1 organ."}
	}

	card := emisor.Hand[action.ActuatingCardIdx[0]]

	organ, ok := card.(Organ)
	if !ok {
    	return Result{Type: Error, Message: "Card is not an organ."}
	}

	if len(action.TargetPlayerIdx) != 1 {
		return Result{Error, "Invalid number of target players."}
	}

	target := &g.Players[action.TargetPlayerIdx[0]]

	if PlayerHasOrganColor(*target, organ.Color) {
		return Result{Type: Error, Message: fmt.Sprintf("Player has already organ of color %v.", organ.Color)}
	}

	target.Body = append(target.Body, Joint{Base: organ})
	deleteIndexFromSlice(&emisor.Hand, action.ActuatingCardIdx[0])
	return Result{Success, "Succesfully placed organ"}


 }

func deleteIndexFromSlice[T any](s *[]T, index int) {
	for i := index; i < len(*s)-1; i++ {
		(*s)[i] = (*s)[i+1]
	}
	*s = (*s)[:len(*s)-1]
 }

func playAddon(g *Game, action Action) Result {
	emisor := &g.Players[action.EmisorIdx]

	if len(action.ActuatingCardIdx) != 1 {
		return Result{Error, "Can only play 1 addon at a time."}
	}

	addonToPlay := emisor.Hand[action.ActuatingCardIdx[0]]

	addon, ok := addonToPlay.(ApplicableToOrgan)
	if !ok {
		return Result{Error, "Sent card is not applicable to organs."}
	}

	if len(action.TargetPlayerIdx) != 1 {
		return Result{Error, "Cannot have more or less than 1 target player."}
	}

	target := &g.Players[action.TargetPlayerIdx[0]]

	if len(action.TargetJointIdx) != 1 {
		return Result{Error, "Cannot have more or less than 1 target joint."}
	}

	targetJoint := target.Body[action.TargetJointIdx[0]]

	result :=  g.ApplyAddonToJoint(addon, &targetJoint)

	if result.Type == Success || result.Type == OrganToDestroy {
		deleteIndexFromSlice(&emisor.Hand, action.ActuatingCardIdx[0])
		if result.Type == OrganToDestroy {
			g.DiscardPile = append(g.DiscardPile, targetJoint.Base, targetJoint.Added[0], addonToPlay)
			deleteIndexFromSlice(&target.Body, action.TargetJointIdx[0])
		}
	}

	return result
}

func discardAndDraw(g *Game, action Action) Result {
	emisor := &g.Players[action.EmisorIdx]
	cardIdxsToDiscard := action.ActuatingCardIdx

	if len(cardIdxsToDiscard) > 3 || len(cardIdxsToDiscard) < 1 {
		return Result{Error, "Cards to discard must be between 1 and 3"}
	}

	for i := range cardIdxsToDiscard {
		toDiscard := emisor.Hand[i]
		g.Discard(toDiscard)
		deleteIndexFromSlice(&emisor.Hand, i)
	}

	results := []Result{}

	for range len(cardIdxsToDiscard) {
		results = append(results, g.DrawCard(emisor))
	}

	for _, result := range results {
		if result.Type != Success {
			return result
		}
	}

	return Result{Success, fmt.Sprintf("Succesfully discarded and re-drawed %v cards", len(cardIdxsToDiscard))}

}

func(g *Game) DrawPhase(playerNum int) {
	g.CurrentTurn = rand.Intn(playerNum)
	g.DealInitialHands(playerNum)
}

func(g *Game) PlayPhase() (winnerIndex int) {
	for {
		for i, player := range g.Players {
			if HasWon(player) {
				winnerIndex = i
				return
			}
		}
		action := GetActionFromInput(g.CurrentTurn)
		result := action.ActionFunc(g, action) //execute action
		if result.Type != Success {
			panic(result.Message) //propagate error
		}
		for range 3 - len(g.Players[g.CurrentTurn].Hand) {
			g.DrawCard(&g.Players[g.CurrentTurn]) //draw until 3 cards
		}
		g.CurrentTurn = (g.CurrentTurn + 1) % len(g.Players) //pass the turn
	}
}

func GetActionFromInput(playerIdx int) Action { panic("not implemented") }

func GetPlayerNum() int { panic("not implemented") }

func main() {
	playerNum := GetPlayerNum()

	game := Game{}
	game.InitDeck()
	game.DiscardPile = []Card{}
	game.DrawPhase(playerNum)

	winnerIndex := game.PlayPhase()

	fmt.Printf("Player number %v has won!!! \n", winnerIndex)
}

/* TODO:
[*] DrawPhase / PlayPhase / game loop
[ ] PrintGameState / PrintHand / input parser
[ ] 5 tratamientos
*/
