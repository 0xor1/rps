package rps

import(
	`time`
	`strconv`
	`testing`
	`github.com/stretchr/testify/assert`
)

func Test_NewGame(t *testing.T){
	dur, _ := time.ParseDuration(_DELETE_AFTER)
	delAft := now().Add(dur)

	g := NewGame()

	assert.Equal(t, 0, g.Version.GetVersion(), `game should have initialised version to 0`)
	assert.Equal(t, delAft, g.DeleteAfter, `game should have initialised DeleteAfter`)
	assert.NotEqual(t, ``, g.PlayerIds[0], `game should have initialised PlayerIds[0]`)
	assert.Equal(t, ``, g.PlayerIds[1], `game should not have initialised PlayerIds[1]`)
	assert.Equal(t, ``, g.PlayerChoices[0], `game should not have initialised PlayerChoices[0]`)
	assert.Equal(t, ``, g.PlayerChoices[1], `game should not have initialised PlayerChoices[1]`)
	assert.Equal(t, _WAITING_FOR_OPPONENT, g.State, `game should have set State to _WAITING_FOR_OPPONENT`)
	assert.True(t, g.TurnStart.IsZero(), `game should not have initialised TurnStart`)
}

func Test_IsActive(t *testing.T){
	g := NewGame()

	assert.True(t, g.IsActive(), `game should start as active`)

	g.State = _DEACTIVATED

	assert.False(t, g.IsActive(), `game should be in active when State is set to _DEACTIVATED`)
}

func Test_CreatedBy(t *testing.T){
	g := NewGame()

	assert.NotEqual(t, ``, g.CreatedBy(), `game should start with a non empty CreatedBy value`)

	g.PlayerIds[0] = ``

	assert.Equal(t, ``, g.CreatedBy(), `game should return PlayerIds[0]`)
}

func Test_RegisterNewUser(t *testing.T){
	g := NewGame()
	userId, err := g.RegisterNewUser()

	assert.NotEqual(t, ``, userId, `userId should be a non empty string`)
	assert.Nil(t, err, `err should be nil`)
	assert.Equal(t, _GAME_IN_PROGRESS, g.State, `State should be set to _GAME_IN_PROGRESS`)
	assert.False(t, g.TurnStart.IsZero(), `TurnStart should not be zero`)

	userId, err = g.RegisterNewUser()

	assert.Equal(t, ``, userId, `userId should be an empty string`)
	assert.Equal(t, `all player slots taken`, err.Error(), `err should be appropriate`)
}

func Test_UnregisterUser(t *testing.T){
	g := NewGame()
	err := g.UnregisterUser(``)

	assert.Equal(t, `user is not a player in this game`, err.Error(), `err should be appropriate`)

	userId := g.CreatedBy()
	err = g.UnregisterUser(userId)

	assert.Nil(t, err, `err should be nil`)
	assert.Equal(t, ``, g.PlayerIds[0], `PlayerIds[0] should be empty string`)

	err = g.UnregisterUser(userId)
	assert.Equal(t, `user is not a player in this game`, err.Error(), `err should be appropriate`)
}

func Test_Kick(t *testing.T){
	g := NewGame()

	assert.False(t, g.Kick(), `Kick should return false when _WAITING_FOR_OPPONENT`)

	g.State = _DEACTIVATED

	assert.False(t, g.Kick(), `Kick should return false when _DEACTIVATED`)

	g.RegisterNewUser()

	assert.False(t, g.Kick(), `Kick should return false when _GAME_IN_PROGRESS`)

	dur, _ := time.ParseDuration(`-` + strconv.Itoa(_TURN_LENGTH + _TURN_LENGTH_ERROR_MARGIN + 1000) + _TIME_UNIT)
	g.TurnStart = now().Add(dur)

	assert.True(t, g.Kick(), `Kick should return true when Turn is over`)
	assert.NotEqual(t, ``, g.PlayerChoices[0], `PlayerChoices[0] should have been set`)
	assert.NotEqual(t, ``, g.PlayerChoices[1], `PlayerChoices[1] should have been set`)
	assert.Equal(t, _WAITING_FOR_RESTART, g.State, `State should have been set to _WAITING_FOR_RESTART`)

	dur, _ = time.ParseDuration(`-` + strconv.Itoa(_TURN_LENGTH + _TURN_LENGTH_ERROR_MARGIN + _RESTART_TIME_LIMIT + 1000) + _TIME_UNIT)
	g.TurnStart = now().Add(dur)

	assert.True(t, g.Kick(), `Kick should return true when Restart time out is over`)
	assert.Equal(t, _DEACTIVATED, g.State, `State should have been set to _DEACTIVATED`)

	g.State = _WAITING_FOR_RESTART
	g.PlayerChoices[0] = ``

	assert.True(t, g.Kick(), `Kick should return true when Restart time out is over`)
	assert.Equal(t, _WAITING_FOR_OPPONENT, g.State, `State should have been set to _WAITING_FOR_OPPONENT`)
	assert.NotEqual(t, ``, g.PlayerIds[0], `PlayerIds[0] should not have been cleared from the game`)
	assert.Equal(t, ``, g.PlayerIds[1], `PlayerIds[1] should have been cleared from the game`)
	assert.Equal(t, ``, g.PlayerChoices[1], `PlayerChoices[1] should have been cleared from the game`)

	g.State = _WAITING_FOR_RESTART
	g.PlayerIds = [2]string{`1`,`2`}
	g.PlayerChoices = [2]string{`1`,``}

	assert.True(t, g.Kick(), `Kick should return true when Restart time out is over`)
	assert.Equal(t, _WAITING_FOR_OPPONENT, g.State, `State should have been set to _WAITING_FOR_OPPONENT`)
	assert.NotEqual(t, ``, g.PlayerIds[1], `PlayerIds[0] should not have been cleared from the game`)
	assert.Equal(t, ``, g.PlayerIds[0], `PlayerIds[1] should have been cleared from the game`)
	assert.Equal(t, ``, g.PlayerChoices[0], `PlayerChoices[0] should have been cleared from the game`)
}
