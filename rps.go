package rps

import(
	`time`
	`errors`
	`regexp`
	`strings`
	`github.com/0xor1/oak`
	`github.com/0xor1/joak`
	`github.com/gorilla/mux`
)

const(
	_ACT 		= `act`
	_RESTART 	= `restart`
	_CHOOSE 	= `choose`
	_VAL 		= `val`
)

func RouteLocalTest(router *mux.Router, options []string, resultHalfMatrix [][]int, millisecsPerChoice int, newAuthKey string, newCrypKey string, oldAuthKey string, oldCrypKey string){
	initStaticProperties(options, resultHalfMatrix, millisecsPerChoice)
	deleteAfter, _ := time.ParseDuration(_DELETE_AFTER)
	joak.RouteLocalTest(router, newGame, initGame, 300, `rps`, newAuthKey, newCrypKey, oldAuthKey, oldCrypKey, newGame(), getJoinResp, getEntityChangeResp, performAct, deleteAfter)
}

func RouteGaeProd(router *mux.Router, options []string, resultHalfMatrix [][]int, millisecsPerChoice int, newAuthKey string, newCrypKey string, oldAuthKey string, oldCrypKey string, ctxFactory joak.ContextFactory) error {
	initStaticProperties(options, resultHalfMatrix, millisecsPerChoice)
	deleteAfter, _ := time.ParseDuration(_DELETE_AFTER)
	clearAfter, _ := time.ParseDuration(_DELETE_AFTER)
	return joak.RouteGaeProd(router, newGame, initGame, 300, `rps`, newAuthKey, newCrypKey, oldAuthKey, oldCrypKey, newGame(), getJoinResp, getEntityChangeResp, performAct, deleteAfter, clearAfter, `game`, ctxFactory)
}

func initStaticProperties(ops []string, rhm [][]int, millisecsPerChoice int){
	options = ops
	resultHalfMatrix = rhm
	turnLength = millisecsPerChoice * len(options)
	validInput = regexp.MustCompile(`^(`+strings.Join(options, `|`)+`)$`)
}

func getJoinResp(userId string, e oak.Entity) oak.Json {
	resp := getEntityChangeResp(userId, e)
	g, _ := e.(*game)
	resp[`options`] = options
	resp[`pastChoices`] = g.PastChoices
	resp[`resultHalfMatrix`] = resultHalfMatrix
	resp[`turnLength`] = turnLength
	resp[`rematchTimeLimit`] = _REMATCH_TIME_LIMIT
	resp[`maxTurns`] = _DOUBLE_MAX_TURNS / 2
	resp[`myIdx`] = g.getPlayerIdx(userId)
	return resp
}

func getEntityChangeResp(userId string, e oak.Entity) oak.Json {
	g, _ := e.(*game)
	pastChoicesCount := len(g.PastChoices)
	json := oak.Json{
		`turnStart`: g.TurnStart,
		`state`: g.State,
		`currentChoices`: g.CurrentChoices,
		`pastChoicesCount`: pastChoicesCount,
	}
	if pastChoicesCount > 0 {
		json[`penultimateChoices`] = g.PastChoices[pastChoicesCount - 2:]
	}
	if g.State == _GAME_IN_PROGRESS {
		idx := g.getPlayerIdx(userId)
		if idx == -1 || g.CurrentChoices[idx] == `` {
			json[`currentChoices`] = []string{``, ``}
		}
	}
	return json
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