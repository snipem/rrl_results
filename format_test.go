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

	fmt.Println(formatWhatsApp(r, true, true))
}

func Test_getTeams(t *testing.T) {
	teams, err := getTeams("teams.csv")
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
