package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_getResults(t *testing.T) {
	r, err := getResults("results_zandvoort.txt")
	assert.NoError(t, err)
	fmt.Printf("%v\n", r)

	fmt.Println(formatWhatsApp(r))
}

func Test_getTeams(t *testing.T) {
	teams, err := getTeams("teams_a_serie.txt")
	assert.NoError(t, err)
	fmt.Printf("%v\n", teams)
}
