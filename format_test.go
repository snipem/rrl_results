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
