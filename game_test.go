package rps

import(
	`time`
	`strconv`
	`testing`
	`github.com/stretchr/testify/assert`
)

func Test_NewGame(t *testing.T){
	standardSetup()
	g := newGame().(*game)

	assert.Equal(t, 0, g.GetVersion(), `game should have initialised version to 0`)
	assert.True(t, !g.DeleteAfter.IsZero(), `game should have initialised DeleteAfter`)
	assert.NotEqual(t, ``, g.PlayerIds[0], `game should have initialised PlayerIds[0]`)
	assert.Equal(t, ``, g.PlayerIds[1], `game should not have initialised PlayerIds[1]`)
	assert.Equal(t, ``, g.CurrentChoices[0], `game should not have initialised CurrentChoices[0]`)
	assert.Equal(t, ``, g.CurrentChoices[1], `game should not have initialised CurrentChoices[1]`)
	assert.Equal(t, _WAITING_FOR_OPPONENT, g.State, `game should have set State to _WAITING_FOR_OPPONENT`)
	assert.True(t, g.TurnStart.IsZero(), `game should not have initialised TurnStart`)
}

func Test_Version(t *testing.T){
	standardSetup()
	g := newGame().(*game)

	assert.Equal(t, 0, g.GetVersion(), `game should start with Version 0`)

	g.IncrementVersion()

	assert.Equal(t, 1, g.GetVersion(), `game should have Version 1`)

	g.DecrementVersion()

	assert.Equal(t, 0, g.GetVersion(), `game should have Version 0`)
}

func Test_DeleteAfter(t *testing.T){
	standardSetup()
	g := newGame().(*game)

	assert.True(t, !g.DeleteAfter.IsZero(), `DeleteAfter should not be zero value`)

	now := time.Now().UTC()
	g.SetDeleteAfter(now)

	assert.Equal(t, now, g.DeleteAfter, `DeleteAfter should be set to now`)
}

func Test_IsActive(t *testing.T){
	standardSetup()
	g := newGame().(*game)

	assert.True(t, g.IsActive(), `game should start as active`)

	g.State = _DEACTIVATED

	assert.False(t, g.IsActive(), `game should be in active when State is set to _DEACTIVATED`)
}

func Test_CreatedBy(t *testing.T){
	standardSetup()
	g := newGame().(*game)

	assert.NotEqual(t, ``, g.CreatedBy(), `game should start with a non empty CreatedBy value`)

	g.PlayerIds[0] = ``

	assert.Equal(t, ``, g.CreatedBy(), `game should return PlayerIds[0]`)
}

func Test_RegisterNewUser(t *testing.T){
	standardSetup()
	g := newGame().(*game)
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
	standardSetup()
	g := newGame().(*game)
	err := g.UnregisterUser(``)

	assert.Equal(t, `leaving the game is not permitted, simply choose not to restart after the next turn is over instead`, err.Error(), `err should be appropriate`)
}

func Test_Kick(t *testing.T){
	standardSetup()
	g := newGame().(*game)

	assert.False(t, g.Kick(), `Kick should return false when _WAITING_FOR_OPPONENT`)

	g.State = _DEACTIVATED

	assert.False(t, g.Kick(), `Kick should return false when _DEACTIVATED`)

	g.RegisterNewUser()

	assert.False(t, g.Kick(), `Kick should return false when _GAME_IN_PROGRESS`)

	dur, _ := time.ParseDuration(`-` + strconv.Itoa(turnLength + 1000) + _TIME_UNIT)
	g.TurnStart = now().Add(dur)

	assert.True(t, g.Kick(), `Kick should return true when Turn is over`)
	assert.NotEqual(t, ``, g.CurrentChoices[0], `CurrentChoices[0] should have been set`)
	assert.NotEqual(t, ``, g.CurrentChoices[1], `CurrentChoices[1] should have been set`)
	assert.Equal(t, _WAITING_FOR_REMATCH, g.State, `State should have been set to _WAITING_FOR_RESTART`)

	for i := 0; i < _MAX_TURNS; i++ {
		g.PastChoices = append(g.PastChoices, [2]string{})
	}
	g.State = _GAME_IN_PROGRESS

	assert.True(t, g.Kick(), `Kick should return true when _MAX_TURNS reached`)
	assert.Equal(t, _DEACTIVATED, g.State, `State should have been set to _DEACTIVATED`)


	g.State = _WAITING_FOR_REMATCH
	dur, _ = time.ParseDuration(`-` + strconv.Itoa(turnLength + _REMATCH_TIME_LIMIT + 1000) + _TIME_UNIT)
	g.TurnStart = now().Add(dur)

	assert.True(t, g.Kick(), `Kick should return true when Restart time out is over`)
	assert.Equal(t, _DEACTIVATED, g.State, `State should have been set to _DEACTIVATED`)
}

func Test_makeChoice(t *testing.T){
	standardSetup()
	g := newGame().(*game)

	dur, _ := time.ParseDuration(`-` + strconv.Itoa(1000) + _TIME_UNIT)
	g.TurnStart = now().Add(dur)
	g.State = _GAME_IN_PROGRESS
	g.PlayerIds[0] = `0`
	g.PlayerIds[1] = `1`
	g.CurrentChoices[0] = `ppr`
	for i := 0; i < _MAX_TURNS; i++ {
		g.PastChoices = append(g.PastChoices, [2]string{})
	}

	g.makeChoice(`1`, `rck`)

	assert.Equal(t, _DEACTIVATED, g.State, `State should have been set to _DEACTIVATED`)
}
