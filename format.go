package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func readCSVFile(filename string) ([][]string, error) {

	var records [][]string

	file, err := os.Open(filename)
	if err != nil {
		return records, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err = reader.ReadAll()
	if err != nil {
		fmt.Println("Error reading file:", err)
		return records, err
	}
	return records, nil
}

func extractMetaContent(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	fmt.Println("Response Body:", string(body))

	content := ""
	doc.Find("meta[property=\"og:title\"]").Each(func(i int, s *goquery.Selection) {
		content = strings.TrimSpace(s.Text())
	})
	return content, nil
}

func main() {

	var results string
	var teams string
	var einteilung string

	flag.StringVar(&results, "results", "", "Results parameter")
	flag.StringVar(&teams, "teams", "", "Teams parameter")
	flag.StringVar(&einteilung, "einteilung", "", "Einteilung parameter")

	flag.Parse()

	fmt.Println("Results:", results)
	fmt.Println("Teams:", teams)
	fmt.Println("Einteilung:", einteilung)

	_, _ = getResults(results)

}

type Position struct {
	Position   int
	Name       string
	DNF        bool
	FastestLap bool
}

type Result struct {
	Standings []Position
	URL       string
	EventName string
}

func formatWhatsApp(r Result) (string, error) {
	rs := "```\n"
	rs += "Nicht offizielles Ergebnis!\n\n"

	for _, standing := range r.Standings {
		rs = rs + fmt.Sprintf("%2d. %s\n", standing.Position, standing.Name)
	}

	rs += "```"

	return rs, nil
}

func getResults(results string) (Result, error) {
	resultsCSV, err := readCSVFile(results)
	r := Result{}

	if err != nil {
		return r, err
	}

	// url is in first line
	url := resultsCSV[0][0]
	r.URL = url

	// meta content is in second line
	metaContent, err := extractMetaContent(url)
	if err != nil {
		return r, err
	}
	r.EventName = metaContent

	// fastestLap is in second line
	fastestLap := resultsCSV[1][0]

	// dnf position is in last line
	dnfPosition, err := strconv.Atoi(resultsCSV[len(resultsCSV)-1][0])
	if err != nil {
		return r, err
	}

	position := 1
	for i := 2; i < len(resultsCSV)-1; i++ {
		name := resultsCSV[i][0]
		r.Standings = append(r.Standings, Position{
			Position:   position,
			Name:       name,
			DNF:        position == dnfPosition,
			FastestLap: name == fastestLap,
		})
		position += 1
	}

	return r, nil
}
