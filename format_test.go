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

	fmt.Println(formatWhatsApp(r))
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
