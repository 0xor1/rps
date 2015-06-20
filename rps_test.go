package rps

import(
	`time`
	`strconv`
	`testing`
	`github.com/stretchr/testify/assert`
	"github.com/0xor1/oak"
)

func Test_GetJoinResp(t *testing.T){
	g := NewGame()

	json := GetJoinResp(``, g)

	assert.Equal(t, _TURN_LENGTH, json[`turnLength`], `turnLength should be _TURN_LENGTH`)
	assert.Equal(t, g.getPlayerIdx(``), json[`myIdx`], `myIdx should be -1 when just observing`)
	assert.Equal(t, g.State, json[`state`], `state should be g.State`)
	assert.Equal(t, g.PlayerChoices, json[`choices`], `state should be g.State`)
	assert.Equal(t, 4, len(json), `json should contain 4 entries`)
}

func Test_GetEntityChangeResp(t *testing.T){
	g := NewGame()

	json := GetEntityChangeResp(``, g)

	assert.Equal(t, g.State, json[`state`], `state should be g.State`)
	assert.Equal(t, g.PlayerChoices, json[`choices`], `state should be g.State`)
	assert.Equal(t, 2, len(json), `json should contain 2 entries`)
}

func Test_PerformAct_without_act_param(t *testing.T){
	json := oak.Json{}
	g := NewGame()

	err := PerformAct(json, ``, g)

	assert.Equal(t, _ACT + ` value must be included in request`, err.Error(), `error should include appropriate message`)
}

func Test_PerformAct_with_non_string_act_param(t *testing.T){
	json := oak.Json{`act`:true}
	g := NewGame()

	err := PerformAct(json, ``, g)

	assert.Equal(t,_ACT + ` must be a string value`, err.Error(), `error should include appropriate message`)
}

func Test_PerformAct_with_invalid_act_param(t *testing.T){
	json := oak.Json{`act`:`fail`}
	g := NewGame()

	err := PerformAct(json, ``, g)

	assert.Equal(t, _ACT + ` must be either ` + _RESTART + ` or ` + _CHOOSE, err.Error(), `error should include appropriate message`)
}

func Test_PerformAct_restart_when_inappropriate_time(t *testing.T){
	json := oak.Json{`act`:`restart`}
	g := NewGame()

	err := PerformAct(json, ``, g)

	assert.Equal(t, `game is not waiting for restart`, err.Error(), `error should include appropriate message`)
}

func Test_PerformAct_restart_with_invalid_user(t *testing.T){
	json := oak.Json{`act`:`restart`}

	g := NewGame()
	dur, _ := time.ParseDuration(`-` + strconv.Itoa(_TURN_LENGTH + _TURN_LENGTH_ERROR_MARGIN + 1000) + _TIME_UNIT)
	g.TurnStart = now().Add(dur)
	g.State = _WAITING_FOR_RESTART

	err := PerformAct(json, ``, g)

	assert.Equal(t, `user is not a player in this game`, err.Error(), `error should include appropriate message`)
}