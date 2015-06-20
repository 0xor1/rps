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
	_START_TIME_BUF				= 5000
	_RESTART_TIME_LIMIT			= 10000
	_TIME_UNIT					= `ms`
	_DELETE_AFTER				= `10m`
	//STATE
	_WAITING_FOR_OPPONENT		= 0
	_GAME_IN_PROGRESS 			= 1
	_WAITING_FOR_RESTART		= 2
	_DEACTIVATED				= 3
)

var(
	validInput = regexp.MustCompile(`^(`+_RCK+`|`+_PPR+`|`+_SCR+`)$`)
)

func now() time.Time {
	return time.Now().UTC()
}

func NewGame() *game {
	g := &game{Version: sus.NewVersion(), State: _WAITING_FOR_OPPONENT}
	g.PlayerIds[0] = sid.ObjectId()
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

func (g *game) CreatedBy() string {
	return g.PlayerIds[0]
}

func (g *game) RegisterNewUser() (string, error) {
	for i := 0; i < 2 ; i++ {
		if g.PlayerIds[i] == `` {
			g.PlayerIds[i] = sid.ObjectId()
			if g.PlayerIds[0] != `` && g.PlayerIds[1] != `` {
				dur, _ := time.ParseDuration(strconv.Itoa(_START_TIME_BUF) + _TIME_UNIT)
				g.TurnStart = now().Add(dur)
				g.State = _GAME_IN_PROGRESS
				g.setDeleteAfter()
			}
			return g.PlayerIds[i], nil
		}
	}
	return ``, errors.New(`all player slots taken`)
}

func (g *game) UnregisterUser(userId string) error {
	idx := g.getPlayerIdx(userId)
	if idx == -1 {
		return errors.New(`user is not a player in this game`)
	}

	g.PlayerIds[idx] = ``
	var t time.Time
	g.TurnStart = t
	g.State = _WAITING_FOR_OPPONENT
	g.setDeleteAfter()
	return nil
}

func (g *game) Kick() bool {
	if g.State == _WAITING_FOR_OPPONENT || g.State == _DEACTIVATED {
		return false
	}

	ret := false
	if g.State == _GAME_IN_PROGRESS {
		dur, _ := time.ParseDuration(strconv.Itoa(_TURN_LENGTH + _TURN_LENGTH_ERROR_MARGIN) + _TIME_UNIT)
		if now().After(g.TurnStart.Add(dur)) {
			g.State = _WAITING_FOR_RESTART
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
				for i := 0; i < 2; i++ {
					if g.PlayerChoices[i] != `` {
						g.PlayerChoices[i] = ``
						g.UnregisterUser(g.PlayerIds[i])
						break
					}
				}
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

	idx := g.getPlayerIdx(userId)
	if idx == -1 {
		return errors.New(`user is not a player in this game`)
	}

	if validInput.MatchString(choice) == false {
		return errors.New(`choice is not a valid string, must be one of: `+_RCK+`, `+_PPR+`, `+_SCR)
	}

	if g.PlayerChoices[idx] == `` {
		if now().After(g.TurnStart) {
			g.PlayerChoices[idx] = choice
			return nil
		}
		return errors.New(`turn hasn't started yet`)
	}
	return errors.New(`user choice has already been made`)
}

func (g *game) restart(userId string) error {
	g.Kick()

	if g.State != _WAITING_FOR_RESTART {
		return errors.New(`game is not waiting for restart`)
	}

	idx := g.getPlayerIdx(userId)
	if idx == -1 {
		return errors.New(`user is not a player in this game`)
	}

	if g.PlayerChoices[idx] != `` {
		g.PlayerChoices[idx] = ``
		if g.PlayerIds[1 - idx] == `` {
			dur, _ := time.ParseDuration(strconv.Itoa(_START_TIME_BUF) + _TIME_UNIT)
			g.TurnStart = now().Add(dur)
			g.State = _GAME_IN_PROGRESS
			g.setDeleteAfter()
		}
		return nil
	}
	return errors.New(`player has already opted to restart`)
}

func (g *game) getPlayerIdx(userId string) int {
	if userId == `` {
		return -1
	}

	if userId == g.PlayerIds[0] {
		return 0
	}

	if userId == g.PlayerIds[1] {
		return 1
	}

	return -1
}

func (g *game) setDeleteAfter() {
	dur, _ := time.ParseDuration(_DELETE_AFTER)
	g.DeleteAfter = now().Add(dur)
}
