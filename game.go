package rps

import(
	`time`
	`errors`
	`regexp`
	`strconv`
	`math/rand`
	`github.com/0xor1/sid`
	`github.com/0xor1/sus`
	"github.com/0xor1/oak"
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
	_DELETE_AFTER				= `10m`
	//STATE
	_WAITING_FOR_OPPONENT		= 0
	_GAME_IN_PROGRESS			= 1
	_WAITING_FOR_RESTART		= 2
	_DEACTIVATED				= 3
)

var(
	validInput = regexp.MustCompile(`^(`+_RCK+`|`+_PPR+`|`+_SCR+`)$`)
)

func now() time.Time {
	return time.Now().UTC()
}

func NewGame() oak.Entity {
	g := &game{sus.NewVersion()}
	g.PlayerIds[0] = sid.ObjectId()
	g.State = _WAITING_FOR_OPPONENT
	g.setDeleteAfter()
	return g
}

type game struct {
	sus.Version					`datastore:",noindex"`
	PlayerIds 		[2]string	`datastore:",noindex"`
	State	 		int			`datastore:",noindex"`
	TurnStart		time.Time	`datastore:",noindex"`
	PlayerChoices	[2]string	`datastore:",noindex"`
	DeleteAfter		time.Time	`datastore:""`
}

func (g *game) IsActive() bool {
	return g.State != _DEACTIVATED
}

func (g *game) OwnedBy() string {
	return g.PlayerIds[0]
}

func (g *game) RegisterNewUser() (string, error) {
	for i := 0; i < 2 ; i++ {
		if g.PlayerIds[i] == `` {
			g.PlayerIds[i] = sid.ObjectId()
			if i == 1 {
				dur, _ := time.ParseDuration(strconv.Itoa(_START_TIME_BUF) + _TIME_UNIT)
				g.TurnStart = now().Add(dur)
				g.State = _GAME_IN_PROGRESS
			}
			return g.PlayerIds[1], nil
		}
		g.PlayerIds[i] = sid.ObjectId()
	}
	return ``, errors.New(`all player slots taken`)
}

func (g *game) UnregisterUser(userId string) error {
	if userId == `` {
		return errors.New(`userId can not be empty string`)
	}
	if g.PlayerIds[1] == userId {
		g.PlayerIds[1] = ``
		g.State = _WAITING_FOR_OPPONENT
		return nil
	}
	if g.PlayerIds[0] == userId {
		g.PlayerIds[0] = g.PlayerIds[1]
		g.PlayerIds[1] = ``
		g.State = _WAITING_FOR_OPPONENT
		return nil
	}
	return errors.New(userId + ` is not a player in this game`)
}

func (g *game) Kick() bool {
	if g.State == _WAITING_FOR_OPPONENT || g.State == _DEACTIVATED {
		return false
	}

	ret := false
	if g.State == _GAME_IN_PROGRESS {
		dur, _ := time.ParseDuration(strconv.Itoa(_TURN_LENGTH + _TURN_LENGTH_ERROR_MARGIN) + _TIME_UNIT)
		if now().After(g.TurnStart.Add(dur)) {
			g.State == _WAITING_FOR_RESTART
			ret = true
			for i := 0; i < 2; i++ {
				if g.PlayerChoices[i] == `` {
					switch r := rand.Intn(3); r {
						case 0: g.PlayerChoices[i] = _RCK
						case 1: g.PlayerChoices[i] = _PPR
						case 2: g.PlayerChoices[i] = _SCR
					}
				}
			}
		}
	}

	if g.State == _WAITING_FOR_RESTART {
		dur, _ := time.ParseDuration(strconv.Itoa(_TURN_LENGTH + _TURN_LENGTH_ERROR_MARGIN + _RESTART_TIME_LIMIT) + _TIME_UNIT)
		if now().After(g.TurnStart.Add(dur)) {
			ret = true
			if (g.PlayerChoices[0] == `` || g.PlayerChoices[1] == ``) && !(g.PlayerChoices[0] == `` && g.PlayerChoices[1] == ``) {
				g.State = _WAITING_FOR_OPPONENT
				g.setDeleteAfter()
			} else {
				g.State = _DEACTIVATED
			}
		}
	}

	return ret
}

func (g *game) makeChoice(userId string, choice string) error {
	g.Kick()

	if g.State != _GAME_IN_PROGRESS {
		return errors.New(`game is not in progress`)
	}

	if userId == `` {
		return errors.New(`userId can not be empty string`)
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

func (g *game) restart(userId string) error {
	g.Kick()

	if g.State != _WAITING_FOR_RESTART {
		return errors.New(`game is not waiting for restart`)
	}

	if userId == `` {
		return errors.New(`userId can not be empty string`)
	}

	if g.PlayerIds[0] != userId && g.PlayerIds[1] != userId {
		return errors.New(`user is not a player in this game`)
	}

	dur, _ := time.ParseDuration(strconv.Itoa(_TURN_LENGTH + _TURN_LENGTH_ERROR_MARGIN + _RESTART_TIME_LIMIT) + _TIME_UNIT)
	if g.PlayerChoices[0] != `` && g.PlayerChoices[1] != `` && now().Before(g.TurnStart.Add(dur)) {
		dur, _ := time.ParseDuration(strconv.Itoa(_START_TIME_BUF) + _TIME_UNIT)
		g.TurnStart = now().Add(dur)
		return nil
	}
	return errors.New(`cannot restart game now`)
}

func (g *game) setDeleteAfter() {
	dur, _ := time.ParseDuration(_DELETE_AFTER)
	g.DeleteAfter = now().Add(dur)
}
