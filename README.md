# Virus! en Go

Implementación en Go del juego de cartas **Virus!**: un juego de 2 a 6 jugadores donde cada uno arma un cuerpo con órganos de colores y lo defiende de los virus de los demás. Gana quien primero tiene **4 órganos sanos**.

El proyecto está pensado como un **núcleo de reglas** independiente de la interfaz, para poder conectarle después una CLI, una GUI y un motor (IA) que juegue contra personas.

> **Estado:** el núcleo de reglas está escrito pero todavía no hay forma de jugar: la entrada del usuario (`GetPlayerNum`, `GetActionFromInput`) aún no está implementada. Ver [Roadmap](#roadmap).

---

## Requisitos y ejecución

- Go **1.23 o superior** (usa `for range n` y `slices.Backward`).

```sh
go vet ./...
go run main.go
```

Por ahora `main` termina en `panic("not implemented")` en `GetPlayerNum`.

---

## Reglas implementadas

### Mazo (68 cartas)

| Tipo | Cartas |
|---|---|
| Órganos | 5 por color (rojo, azul, amarillo, verde) + 1 multicolor = 21 |
| Virus | 4 por color + 1 multicolor = 17 |
| Medicinas | 4 por color + 4 multicolor = 20 |
| Tratamientos | 3 Transplant, 3 Organ Thief, 2 Contagion, 1 Latex Glove, 1 Medical Error = 10 |

### Turno

1. Cada jugador empieza con 3 cartas; empieza un jugador al azar.
2. En su turno, el jugador ejecuta **una acción**: jugar una carta o descartar de 1 a 3.
3. Al terminar, roba hasta volver a tener 3 cartas.
4. Si se acaba el mazo, el descarte se da vuelta y pasa a ser el nuevo mazo.

### Estados de un órgano (`Joint`)

Un órgano en la mesa puede tener 0, 1 o 2 cartas encima (`Added`). El estado se **deriva** de esas cartas (`DeriveJointState`), no se guarda aparte.

| Estado | Cartas encima | + Virus (color compatible) | + Medicina (color compatible) |
|---|---|---|---|
| Libre | ninguna | → Infectado | → Vacunado |
| Vacunado | 1 medicina | se retira la medicina | → Inmunizado |
| Inmunizado | 2 medicinas | ilegal | ilegal |
| Infectado | 1 virus | se destruye el órgano | se retira el virus |

Compatibilidad de colores (`ColorsMatch`): dos colores son compatibles si son iguales o si alguno es multicolor. Una carta añadida se compara siempre con el color del **órgano base**.

### Victoria

Gana quien tenga 4 órganos sanos (es decir, no infectados) de colores distintos. Un jugador no puede tener dos órganos del mismo color.

---

## Arquitectura

Todo vive hoy en `main.go`, organizado por capas.

```
Card (interfaz)                    GetCardType()
├── Organ{Color}
├── Medicine{Color}  ┐
├── Virus{Color}     ┴─ ApplicableToOrgan: Card + GetColor() + AddonFunc()
└── Treatment{TreatmentType}

Joint   = Organ base + []ApplicableToOrgan encima   (un órgano en la mesa)
Player  = Hand []Card + Body []Joint
Game    = Players, GrabPile (mazo), DiscardPile, CurrentTurn
```

- **`ApplicableToOrgan`** es la interfaz que separa las cartas que se colocan sobre un órgano (virus y medicinas) de las que no (órganos y tratamientos). Se usa con una aserción de tipo: `card.(ApplicableToOrgan)`.
- **`Result{Type, Message}`** es el resultado de toda operación (`Success`, `OrganToDestroy`, `Error`, `Illegal`, `NotImplemented`). No transporta estado, solo informa.
- **`Action`** describe una jugada solo con **índices** (jugador emisor, jugadores objetivo, cartas, órganos) más `ActionFunc func(*Game, Action) Result`, la función que la ejecuta. El bucle de juego hace `action.ActionFunc(g, action)` sin saber de qué acción se trata.

### Funciones principales

| Función | Qué hace |
|---|---|
| `BuildStartingDeck`, `InitDeck`, `Shuffle` | Construye y baraja el mazo |
| `DrawCard`, `DealInitialHands`, `DrawPhase` | Robo, reparto inicial y primer jugador |
| `PlayOrgan` | Coloca un órgano en un cuerpo |
| `playAddon` + `ApplyAddonToJoint` | Aplica un virus o una medicina a un órgano |
| `discardAndDraw` | Descarta 1 a 3 cartas y roba |
| `HasWon` | Comprueba la condición de victoria |
| `PlayPhase` | Bucle de juego: acción, robo, cambio de turno |

---

## Modelo de memoria y convenciones

Go copia los structs por valor al asignarlos. Para que una acción modifique el juego real hay que trabajar con **punteros a los elementos**, no con copias.

**Regla del proyecto:** las acciones mutan el `*Game` in situ.

```go
emisor := &g.Players[action.EmisorIdx]   // puntero: modifica el jugador real
emisor := g.Players[action.EmisorIdx]    // COPIA: los cambios se pierden
```

Lo mismo para órganos: `joint := &target.Body[i]`.

Detalles que conviene tener presentes:

- **Un slice es un header `(puntero, len, cap)`.** Copiar un `Player` copia los headers de `Hand` y `Body` pero comparte los arrays. Cambiar un *elemento* a través de la copia se ve en el original; cambiar el *largo* (`append`, resegmentar) solo afecta al header de la copia.
- **Las cartas son valores dentro de una interfaz.** No tienen identidad: dos cartas iguales son indistinguibles. Por eso una carta duplicada o perdida no da ningún error; se detecta contando (el total debe ser siempre 68, repartido entre mazo, descarte, manos y cuerpos).
- **Los punteros a elementos de un slice se invalidan si ese slice crece** (`append` puede reubicar el array). No conservar `&Body[i]` a través de un `append` sobre el mismo `Body`.
- **Atomicidad:** una acción valida todo antes de mutar. Si falla, el estado queda intacto.
- Todos los métodos de `Game` usan receptor puntero.

---

## Roadmap

**Base del núcleo**
- [ ] Ejecutar el código y cubrirlo con tests (la lógica fue escrita sin ejecutarse)
- [ ] Invariante de conservación: el total de cartas siempre es 68
- [ ] Tests de tabla para los estados de un órgano (virus y medicina, colores incluyendo multicolor)
- [ ] Validar índices al inicio de cada acción, antes de mutar
- [ ] Una jugada ilegal debe pedir otra acción, no terminar el programa

**Tratamientos**
- [ ] Transplant, Organ Thief, Contagion, Latex Glove, Medical Error

**Interfaz**
- [ ] `GetPlayerNum` y `GetActionFromInput`
- [ ] `PrintGameState`, `PrintHand` y parser de comandos de texto

**Visión a largo plazo**
- [ ] GUI donde las cartas se vean moverse y se deslicen para ejecutar acciones, con entrada por comandos en paralelo
- [ ] Motor para jugar contra la máquina, con el núcleo sin distinguir entre humano y motor
- [ ] IA de self-play (MCTS / estilo AlphaZero)

---

## Licencia

Sin definir.