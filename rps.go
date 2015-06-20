package rps

import(
	`errors`
	"github.com/0xor1/oak"
)

const(
	_ACT = `act`
	_RESTART = `restart`
	_CHOOSE = `choose`
	_VAL = `val`
)

func GetJoinResp (userId string, e oak.Entity) oak.Json {
	resp := GetEntityChangeResp(userId, e)
	g, _ := e.(*game)
	resp[`turnLength`] = _TURN_LENGTH
	resp[`myIdx`] = g.getPlayerIdx(userId)
	return resp
}

func GetEntityChangeResp (userId string, e oak.Entity) oak.Json {
	g, _ := e.(*game)
	return oak.Json{
		`state`: g.State,
		`choices`: g.PlayerChoices,
	}
}

func PerformAct (json oak.Json, userId string, e oak.Entity) (err error) {
	g, _ := e.(*game)
	if actParam, exists := json[_ACT]; exists {
		if act, ok := actParam.(string); ok {
			if act == _RESTART {
				return g.restart(userId)
			}
			if act == _CHOOSE {
				if valParam, exists := json[_VAL]; exists {
					if val, ok := valParam.(string); ok {
						return g.makeChoice(userId, val)
					}else {
						return errors.New(_VAL + ` must be a string value`)
					}
				} else {
					return errors.New(_VAL + ` value must be included in request`)
				}
			}
			return errors.New(_ACT + ` must be either ` + _RESTART + ` or ` + _CHOOSE)
		} else {
			return errors.New(_ACT + ` must be a string value`)
		}
	} else {
		return errors.New(_ACT + ` value must be included in request`)
	}
}