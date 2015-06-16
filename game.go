package rps

import(
	`time`
	`errors`
	`regexp`
	`strconv`
	`math/rand`
	`github.com/0xor1/sid`
	`github.com/0xor1/sus`
)

const(
	_RCK 						= `rck`
	_PPR 						= `ppr`
	_SCR 						= `scr`
	_TURN_LENGTH 				= 3000
	_TURN_LENGTH_ERROR_MARGIN	= 500
	_START_TIME_BUF				= 3000
	_RESTART_TIME_LIMIT			= 5000
	_TIME_UNIT					= `ms`
	_DELETE_AFTER				= `1h`
)

var(
	validInput = regexp.MustCompile(`^(`+_RCK+`|`+_PPR+`|`+_SCR+`)$`)
)

type Game interface {
	sus.Version
	IsActive() bool
	OwnedBy() string
	RegisterNewPlayer() (string, error)
	UnregisterPlayer(userId string) error
	Kick() bool
	MakeChoice(userId string, choice string) error
}

func now() time.Time {
	return time.Now().UTC()
}

func NewGame() Game {
	g := &game{sus.NewVersion()}
	g.PlayerIds[0] = sid.ObjectId()
	dur, _ := time.ParseDuration(_DELETE_AFTER)
	g.DeleteAfter = now().Add(dur)
	return g
}

type game struct {
	sus.Version					`datastore:",noindex"`
	PlayerIds 		[2]string	`datastore:",noindex"`
	TurnStart		time.Time	`datastore:",noindex"`
	PlayerChoices	[2]string	`datastore:",noindex"`
	DeleteAfter		time.Time	`datastore:""`
}

func (g *game) IsActive() bool {
	dur, _ := time.ParseDuration(strconv.Itoa(_TURN_LENGTH + _TURN_LENGTH_ERROR_MARGIN + _RESTART_TIME_LIMIT) + _TIME_UNIT)
	return g.TurnStart.IsZero() || now().Before(g.TurnStart.Add(dur))
}

func (g *game) OwnedBy() string {
	return g.PlayerIds[0]
}

func (g *game) RegisterNewPlayer() (string, error) {
	if g.PlayerIds[1] != `` {
		return ``, errors.New(`all player slots taken`)
	}
	g.PlayerIds[1] = sid.ObjectId()
	dur, _ := time.ParseDuration(strconv.Itoa(_START_TIME_BUF) + _TIME_UNIT)
	g.TurnStart = now().Add(dur)
	return g.PlayerIds[1], nil
}

func (g *game) UnregisterPlayer(userId string) error {
	if userId == `` {
		return errors.New(`userId must can not be empty string`)
	}
	if g.PlayerIds[1] == userId {
		g.PlayerIds[1] = ``
		return nil
	}
	if g.PlayerIds[0] == userId {
		g.PlayerIds[0] = g.PlayerIds[1]
		g.PlayerIds[1] = ``
		return nil
	}
	return errors.New(userId + ` is not a player in this game`)
}

func (g *game) Kick() bool {
	// if turn is over and a player hasn't made a choice, make it for them
	ret := false
	dur, _ := time.ParseDuration(strconv.Itoa(_TURN_LENGTH + _TURN_LENGTH_ERROR_MARGIN) + _TIME_UNIT)
	if now().After(g.TurnStart.Add(dur)) {
		for i := 0; i < 2; i++ {
			if g.PlayerChoices[i] == `` {
				ret = true
				switch r := rand.Intn(3); r {
					case 0: g.PlayerChoices[i] = _RCK
					case 1: g.PlayerChoices[i] = _PPR
					case 2: g.PlayerChoices[i] = _SCR
				}
			}
		}
	}
	return ret
}

func (g *game) makeChoice(userId string, choice string) error {
	if g.TurnStart.IsZero() {
		return errors.New(`game has not started yet`)
	}

	dur, _ := time.ParseDuration(strconv.Itoa(_TURN_LENGTH + _TURN_LENGTH_ERROR_MARGIN) + _TIME_UNIT)
	if now().After(g.TurnStart.Add(dur)) {
		return errors.New(`time limit is over`)
	}

	if userId == `` {
		return errors.New(`userId must can not be empty string`)
	}

	if validInput.MatchString(choice) == false {
		return errors.New(`choice is not a valid string, must be one of: `+_RCK+`, `+_PPR+`, `+_SCR)
	}

	for i := 0; i < 2; i++ {
		if g.PlayerIds[i] == userId {
			if g.PlayerChoices[i] != `` {
				return errors.New(`user choice has already been made`)
			}
			g.PlayerChoices[i] = choice
			return nil
		}
	}

	return errors.New(userId + ` is not a player in this game`)
}

func (g *game) restart() error {
	dur, _ := time.ParseDuration(strconv.Itoa(_TURN_LENGTH + _TURN_LENGTH_ERROR_MARGIN + _RESTART_TIME_LIMIT) + _TIME_UNIT)
	if g.PlayerChoices[0] != `` && g.PlayerChoices[1] != `` && now().Before(g.TurnStart.Add(dur)) {
		dur, _ := time.ParseDuration(strconv.Itoa(_START_TIME_BUF) + _TIME_UNIT)
		g.TurnStart = now().Add(dur)
		return nil
	}
	return errors.New(`cannot restart game now`)
}