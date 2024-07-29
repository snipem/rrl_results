package main

import (
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
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

var pointScale = []int{
	40,
	35,
	32,
	30,
	28,
	26,
	24,
	23,
	22,
	21,
	20,
	19,
	18,
	17,
	16,
	15,
	14,
	13,
	12,
	11,
	10,
	9,
	8,
	7,
	6,
	5,
	4,
	3,
	2,
	1,
}

func extractMetaContent(url string) (title string, series string, err error) {

	series = ""

	if strings.Contains(url, "b-serie") {
		series = "B"
	}

	if strings.Contains(url, "a-serie") {
		series = "A"
	}

	if strings.Contains(url, "multiclass") {
		series = "Multi"
	}

	// Create a new request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return "", "", nil
	}

	// Set the User-Agent header
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.0.0 Safari/537.36")

	// Create a client and send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return "", "", nil
	}
	defer resp.Body.Close()
	defer resp.Body.Close()
	//body, err := io.ReadAll(resp.Body)

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	//body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	//fmt.Println("Response Body:", string(body))

	content := ""
	doc.Find("meta[property=\"og:title\"]").Each(func(i int, s *goquery.Selection) {
		con, _ := s.Attr("content")
		content = strings.TrimSpace(con)
		content = strings.Replace(content, " - Rookie Racing League - ACC / F1 24 / GT7 / WRC 23", "", -1)
	})
	return content, series, nil
}

func main() {

	var results string
	//var teams string
	//var einteilung string

	flag.StringVar(&results, "results", "", "Results parameter")
	//flag.StringVar(&teams, "teams", "", "Teams parameter")
	//flag.StringVar(&einteilung, "einteilung", "", "Einteilung parameter")

	flag.Parse()

	fmt.Println("Results:", results)
	//fmt.Println("Teams:", teams)
	//fmt.Println("Einteilung:", einteilung)

	individualResults, err := getResults(results)
	if err != nil {
		log.Fatal(err)
	}
	teamResults, err := getTeamResults(individualResults)
	if err != nil {
		log.Fatal(err)
	}
	whatsAppMessage, err := formatWhatsApp(individualResults, teamResults, false, true)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(whatsAppMessage)

}

type Position struct {
	Position   int
	Name       string
	DNF        bool
	FastestLap bool
	Points     int
	HomeSeries string
}

type Result struct {
	Standings []Position
	URL       string
	EventName string
	Series    string
}

type Team struct {
	Name       string
	Members    []string
	HomeSeries string
}

func formatWhatsApp(individualResults Result, teamResults Result, showPoints bool, showSeries bool) (string, error) {
	rs := fmt.Sprintf("*%s*\nInoffizielles Ergebnis!\n", individualResults.EventName)
	rs += "```\n"

	rs += getAsciiTable(individualResults, true, showPoints, showSeries)
	rs += "```"
	rs += "\n"
	rs += "Teams nach Punkten:\n"
	rs += "```\n"
	rs += getAsciiTable(teamResults, false, showPoints, false)

	rs += "```"

	return rs, nil
}

func getAsciiTable(r Result, showPosition bool, showPoints bool, showSeries bool) string {
	buffer := bytes.Buffer{}
	w := tabwriter.NewWriter(&buffer, 0, 1, 1, ' ', tabwriter.DiscardEmptyColumns)

	// position
	if showPosition {
		fmt.Fprintf(w, "P\t")
	}
	fmt.Fprintf(w, "Name\t")
	fmt.Fprintf(w, "\t") // remarks
	if showPoints {
		fmt.Fprintf(w, "Pts\t")
	}
	if showSeries {
		fmt.Fprintf(w, "S\t")
	}

	fmt.Fprintf(w, "\n")

	for _, standing := range r.Standings {
		//rs = rs + fmt.Sprintf("%2d. %s\n", standing.Position, standing.Name)
		var remarks []string
		if standing.FastestLap {
			remarks = append(remarks, "SR")
		}
		if standing.DNF {
			remarks = append(remarks, "DNF")
		}

		if showPosition {
			fmt.Fprintf(w, "%d\t", standing.Position)
		}
		fmt.Fprintf(w, "%v\t", standing.Name)
		fmt.Fprintf(w, "%s\t", strings.Join(remarks, ","))
		if showPoints {
			fmt.Fprintf(w, "%s\t", formatPoints(standing.Points))
		}
		if showSeries {
			fmt.Fprintf(w, "%v\t", standing.HomeSeries)
		}

		fmt.Fprintf(w, "\n")

		//if showPoints {
		//	fmt.Fprintf(w, "%d\t%v\t%v\t%s\t\n", standing.Position, standing.Name, getSeries(standing.Name), points, strings.Join(remarks, ","))
		//} else {
		//	fmt.Fprintf(w, "%d\t%v\t%s\t\n", standing.Position, standing.Name, getSeries(standing.Name), strings.Join(remarks, ","))
		//}
	}
	w.Flush()
	return buffer.String()
}

func formatPoints(points int) any {
	if points > 0 {
		return fmt.Sprintf("+%d", points)
	}
	return ""
}

func getSeries(name string) string {

	// TODO This is a hack
	content, err := readCSVFile("einteilung.csv")
	if err != nil {
		log.Fatal(err)
	}
	for _, line := range content {
		if line[1] == name {
			return strings.Replace(line[0], "-Serie", "", 1)
		}
	}
	return ""
}

func getTeams(teamsfile string) ([]Team, error) {
	var teams []Team

	teamCSV, err := readCSVFile(teamsfile)
	if err != nil {
		return teams, err
	}
	for i := 0; i < len(teamCSV); i++ {
		teams = append(teams, Team{
			Name: teamCSV[i][1],
			Members: []string{
				teamCSV[i][2],
				teamCSV[i][3],
			},
			HomeSeries: teamCSV[i][0],
		})
	}
	return teams, nil
}

func isEligibleToEarnPoints(series string, homeSeries string) bool {

	if series == "Multi" {
		return true
	}

	// Anyone can earn points in a no series
	if series == "" {
		return true
	}

	if series == homeSeries {
		return true
	}

	// A drives may only earn points in their home series
	if homeSeries != "A" {
		return true
	}

	return false
}

func getResults(resultsfile string) (Result, error) {
	resultsCSV, err := readCSVFile(resultsfile)
	r := Result{}

	if err != nil {
		return r, err
	}

	// url is in first line
	url := resultsCSV[0][0]
	r.URL = url

	// dnf position is second line
	dnfPosition, err := strconv.Atoi(resultsCSV[1][0])
	if err != nil {
		return r, err
	}

	// meta content is in second line
	eventTitle, series, err := extractMetaContent(url)
	if err != nil {
		return r, err
	}
	r.EventName = eventTitle
	r.Series = series

	// fastestLap is in third line
	fastestLap := resultsCSV[2][0]

	position := 1
	pointsPosition := 0

	for i := 3; i < len(resultsCSV); i++ {
		name := strings.TrimSpace(resultsCSV[i][0])
		homeSeries := getSeries(name)

		didNotFinish := dnfPosition != 0 && position >= dnfPosition
		earnedPoints := 0
		hasFastestLap := name == fastestLap

		if isEligibleToEarnPoints(r.Series, homeSeries) && !didNotFinish {
			earnedPoints = pointScale[pointsPosition]
			pointsPosition += 1

			if hasFastestLap {
				earnedPoints += 2
			}
		}

		r.Standings = append(r.Standings, Position{
			Position:   position,
			Name:       name,
			DNF:        didNotFinish,
			Points:     earnedPoints,
			HomeSeries: homeSeries,
			FastestLap: hasFastestLap,
		})
		position += 1
	}

	return r, nil
}

func getTeamResults(r Result) (Result, error) {

	teamResult := Result{
		EventName: r.EventName,
		Series:    r.Series,
		Standings: []Position{},
	}

	teams, err := getTeams("teams.csv")

	if err != nil {
		return teamResult, err
	}

	for _, team := range teams {

		// only get those teams with matching series
		if r.Series != "" && team.HomeSeries != r.Series {
			continue
		}

		teamPoints := getPoints(r.Standings, team.Members[0]) + getPoints(r.Standings, team.Members[1])

		// do not add teams that have no points
		if teamPoints == 0 {
			continue
		}

		teamResult.Standings = append(teamResult.Standings, Position{
			Name:       team.Name,
			DNF:        false,
			Points:     teamPoints,
			HomeSeries: team.HomeSeries,
			FastestLap: false,
		})

	}

	// sot by points
	sort.Slice(teamResult.Standings, func(i, j int) bool {
		return teamResult.Standings[i].Points > teamResult.Standings[j].Points
	})

	for i, _ := range teamResult.Standings {
		teamResult.Standings[i].Position = i + 1
	}
	return teamResult, nil

}

func getPoints(standings []Position, name string) int {
	for _, standing := range standings {
		if standing.Name == name {
			return standing.Points
		}
	}
	return 0
}
