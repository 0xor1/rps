package rps

import(
	`time`
	`errors`
	`regexp`
	`strings`
	`github.com/0xor1/oak`
	`github.com/0xor1/joak`
	`github.com/gorilla/mux`
	`golang.org/x/net/context`
)

const(
	_ACT 		= `act`
	_RESTART 	= `restart`
	_CHOOSE 	= `choose`
	_VAL 		= `val`
)

func RouteLocalTest(router *mux.Router, ops []string, millisecsPerChoice int){
	initStaticProperties(ops, millisecsPerChoice)
	joak.RouteLocalTest(router, newGame, 300, `rps`, newGame(), getJoinResp, getEntityChangeResp, performAct)
}

func RouteGaeProd(router *mux.Router, options []string, millisecsPerChoice int, ctx context.Context, newAuthKey string, newCrypKey string, oldAuthKey string, oldCrypKey string) error {
	initStaticProperties(options, millisecsPerChoice)
	deleteAfter, _ := time.ParseDuration(_DELETE_AFTER)
	clearAfter, _ := time.ParseDuration(_DELETE_AFTER)
	return joak.RouteGaeProd(router, newGame, 300, `rps`, newGame(), getJoinResp, getEntityChangeResp, performAct, deleteAfter, clearAfter, `game`, ctx, newAuthKey, newCrypKey, oldAuthKey, oldCrypKey)
}

func initStaticProperties(ops []string, millisecsPerChoice int){
	options = ops
	turnLength = millisecsPerChoice * len(options)
	validInput = regexp.MustCompile(`^(`+strings.Join(options, `|`)+`)$`)
}

func getJoinResp(userId string, e oak.Entity) oak.Json {
	resp := getEntityChangeResp(userId, e)
	g, _ := e.(*game)
	resp[`options`] = options
	resp[`turnLength`] = turnLength
	resp[`myIdx`] = g.getPlayerIdx(userId)
	return resp
}

func getEntityChangeResp(userId string, e oak.Entity) oak.Json {
	g, _ := e.(*game)
	return oak.Json{
		`state`: g.State,
		`choices`: g.PlayerChoices,
	}
}

func performAct(json oak.Json, userId string, e oak.Entity) (err error) {
	g, _ := e.(*game)
	if actParam, exists := json[_ACT]; exists {
		if act, ok := actParam.(string); ok {
			if act == _RESTART {
				return g.restart(userId)
			} else if act == _CHOOSE {
				if valParam, exists := json[_VAL]; exists {
					if val, ok := valParam.(string); ok {
						return g.makeChoice(userId, val)
					}else {
						return errors.New(_VAL + ` must be a string value`)
					}
				} else {
					return errors.New(_VAL + ` value must be included in request`)
				}
			} else {
				return errors.New(_ACT + ` must be either ` + _RESTART + ` or ` + _CHOOSE)
			}
		} else {
			return errors.New(_ACT + ` must be a string value`)
		}
	} else {
		return errors.New(_ACT + ` value must be included in request`)
	}
}