package rps

import(
	`regexp`
	`testing`
	`github.com/stretchr/testify/assert`
)

func Test_Uuid(t *testing.T){
	re := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	assert.True(t, , `Uuid should return a valid uuid string`)
}
