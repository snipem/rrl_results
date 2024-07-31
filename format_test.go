package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_getResults(t *testing.T) {
	r, err := getResults("results_zandvoort.csv")
	assert.NoError(t, err)
	fmt.Printf("%v\n", r)
	teamResults, err := getTeamResults(r)
	assert.NoError(t, err)

	assert.Equal(t, r.Standings[1].Driver.id, "dante_1006")
	assert.Equal(t, r.Standings[1].FastestLap, true)
	assert.Equal(t, r.Standings[1].Points, 35+2) // #+2 for fastest lap

	formattedString, err := formatWhatsApp(r, teamResults, true, true)
	assert.NoError(t, err)

	fmt.Println(formattedString)
}

func Test_getTeams(t *testing.T) {
	drivers, err := getDrivers()
	assert.NoError(t, err)
	teams, err := getTeams(drivers)
	assert.NoError(t, err)
	fmt.Printf("%v\n", teams)
}

func Test_getSeries(t *testing.T) {
	assert.Equal(t, "A", getSeries("Triechex"))
	assert.Equal(t, "A", getSeries("LukasFalk46"))
}

func Test_isEligibleToEarnPoints(t *testing.T) {
	assert.False(t, isEligibleToEarnPoints("B", "A"))
	assert.True(t, isEligibleToEarnPoints("B", "Multi"))

	assert.True(t, isEligibleToEarnPoints("A", "B"))
	assert.True(t, isEligibleToEarnPoints("", "A"))

	assert.True(t, isEligibleToEarnPoints("B", "B"))
	assert.True(t, isEligibleToEarnPoints("B", ""))

	assert.True(t, isEligibleToEarnPoints("Multi", "A"))
	assert.True(t, isEligibleToEarnPoints("Multi", ""))
	assert.True(t, isEligibleToEarnPoints("Multi", "B"))
}

func Test_getCircus(t *testing.T) {

	c := getCircus()
	assert.GreaterOrEqual(t, len(c.Drivers), 1)
	assert.GreaterOrEqual(t, len(c.Teams), 1)

}
