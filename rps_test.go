package rps

import(
	`time`
	`strconv`
	`testing`
	`github.com/0xor1/oak`
	`github.com/gorilla/mux`
	`golang.org/x/net/context`
	`github.com/stretchr/testify/assert`
)

func Test_RouteLocal(t *testing.T){
	RouteLocalTest(mux.NewRouter())
}

func Test_RouteGae(t *testing.T){
	RouteGaeProd(mux.NewRouter(), context.Background(), ``, ``, ``, ``)
}

func Test_getJoinResp(t *testing.T){
	g := newGame().(*game)

	json := getJoinResp(``, g)

	assert.Equal(t, _TURN_LENGTH, json[`turnLength`], `turnLength should be _TURN_LENGTH`)
	assert.Equal(t, g.getPlayerIdx(``), json[`myIdx`], `myIdx should be -1 when just observing`)
	assert.Equal(t, g.State, json[`state`], `state should be g.State`)
	assert.Equal(t, g.PlayerChoices, json[`choices`], `state should be g.State`)
	assert.Equal(t, 4, len(json), `json should contain 4 entries`)
}

func Test_getEntityChangeResp(t *testing.T){
	g := newGame().(*game)

	json := getEntityChangeResp(``, g)

	assert.Equal(t, g.State, json[`state`], `state should be g.State`)
	assert.Equal(t, g.PlayerChoices, json[`choices`], `state should be g.State`)
	assert.Equal(t, 2, len(json), `json should contain 2 entries`)
}

func Test_performAct_without_act_param(t *testing.T){
	json := oak.Json{}
	g := newGame()

	err := performAct(json, ``, g)

	assert.Equal(t, _ACT + ` value must be included in request`, err.Error(), `error should include appropriate message`)
}

func Test_performAct_with_non_string_act_param(t *testing.T){
	json := oak.Json{`act`:true}
	g := newGame()

	err := performAct(json, ``, g)

	assert.Equal(t,_ACT + ` must be a string value`, err.Error(), `error should include appropriate message`)
}

func Test_performAct_with_invalid_act_param(t *testing.T){
	json := oak.Json{`act`:`fail`}
	g := newGame()

	err := performAct(json, ``, g)

	assert.Equal(t, _ACT + ` must be either ` + _RESTART + ` or ` + _CHOOSE, err.Error(), `error should include appropriate message`)
}

func Test_performAct_restart_when_inappropriate_time(t *testing.T){
	json := oak.Json{`act`:`restart`}
	g := newGame()

	err := performAct(json, ``, g)

	assert.Equal(t, `game is not waiting for restart`, err.Error(), `error should include appropriate message`)
}

func Test_performAct_restart_with_invalid_user(t *testing.T){
	json := oak.Json{`act`:`restart`}

	g := newGame().(*game)
	dur, _ := time.ParseDuration(`-` + strconv.Itoa(_TURN_LENGTH + _TURN_LENGTH_ERROR_MARGIN + 1000) + _TIME_UNIT)
	g.TurnStart = now().Add(dur)
	g.State = _WAITING_FOR_RESTART

	err := performAct(json, ``, g)

	assert.Equal(t, `user is not a player in this game`, err.Error(), `error should include appropriate message`)
}

func Test_performAct_restart_success(t *testing.T){
	json := oak.Json{`act`:`restart`}

	g := newGame().(*game)
	dur, _ := time.ParseDuration(`-` + strconv.Itoa(_TURN_LENGTH + _TURN_LENGTH_ERROR_MARGIN + 1000) + _TIME_UNIT)
	g.TurnStart = now().Add(dur)
	g.State = _WAITING_FOR_RESTART
	g.PlayerIds = [2]string{`0`, `1`}
	g.PlayerChoices = [2]string{`0`, `1`}

	err := performAct(json, `0`, g)

	assert.Nil(t, err, `err should be nil`)
	assert.Equal(t, ``, g.PlayerChoices[0], `PlayerChoices[0] should be set to empty string`)
	assert.Equal(t, _WAITING_FOR_RESTART, g.State, `State should still be _WAITING_FOR_RESTART`)

	err = performAct(json, `0`, g)

	assert.Equal(t, `player has already opted to restart`, err.Error(), `err should contain appropriate message`)

	err = performAct(json, `1`, g)

	assert.Nil(t, err, `err should be nil`)
	assert.Equal(t, ``, g.PlayerChoices[1], `PlayerChoices[1] should be set to empty string`)
	assert.Equal(t, _GAME_IN_PROGRESS, g.State, `State should be set to _GAME_IN_PROGRESS`)
	dur, _ = time.ParseDuration(strconv.Itoa(_START_TIME_BUF) + _TIME_UNIT)
	assert.Equal(t, now().Add(dur), g.TurnStart, `TurnStart should have been updated`)
	dur, _ = time.ParseDuration(_DELETE_AFTER)
	assert.Equal(t, now().Add(dur), g.DeleteAfter, `DeleteAfter should have been updated`)
}