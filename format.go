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

var circus *Circus

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

	// fmt.Println("Results:", results)
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

type Driver struct {
	id         string
	HomeSeries string
	// Class is PRO, PROAM or AM
	Class string
	Team  *Team
}

type Position struct {
	Position   int
	Driver     *Driver
	DNF        bool
	FastestLap bool
	Points     int
	Team       *Team
}

type Result struct {
	Standings []Position
	URL       string
	EventName string
	Series    string
}

type Team struct {
	Name       string
	Drivers    []*Driver
	HomeSeries string
}

type Circus struct {
	Drivers []*Driver
	Teams   []*Team
}

func formatWhatsApp(individualResults Result, teamResults Result, showPoints bool, showSeries bool) (string, error) {
	rs := fmt.Sprintf("*%s*\nInoffizielles Ergebnis!\n", individualResults.EventName)
	rs += "```\n"

	rs += getAsciiTable(individualResults, true, showPoints, showSeries, false)
	rs += "```"
	rs += "\n"
	rs += "Teams nach Punkten:\n"
	rs += "```\n"
	rs += getAsciiTable(teamResults, true, showPoints, false, true)

	rs += "```"

	return rs, nil
}

func getAsciiTable(r Result, showPosition bool, showPoints bool, showSeries bool, isTeamTable bool) string {
	buffer := bytes.Buffer{}
	w := tabwriter.NewWriter(&buffer, 0, 1, 1, ' ', tabwriter.DiscardEmptyColumns)

	// position
	if showPosition {
		fmt.Fprintf(w, "P\t")
	}
	fmt.Fprintf(w, "Driver\t")
	fmt.Fprintf(w, "\t") // remarks
	if showPoints {
		fmt.Fprintf(w, "Pts\t")
	}
	if showSeries {
		fmt.Fprintf(w, "S\t")
	}

	fmt.Fprintf(w, "\n")

	for _, standing := range r.Standings {
		//rs = rs + fmt.Sprintf("%2d. %s\n", standing.Position, standing.Driver)
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
		if !isTeamTable {
			fmt.Fprintf(w, "%v\t", standing.Driver.id)
		} else {
			fmt.Fprintf(w, "%v\t", standing.Team.Name)
		}
		fmt.Fprintf(w, "%s\t", strings.Join(remarks, ","))
		if showPoints {
			fmt.Fprintf(w, "%s\t", formatPoints(standing.Points))
		}
		if showSeries {
			fmt.Fprintf(w, "%v\t", standing.Driver.HomeSeries)
		}

		fmt.Fprintf(w, "\n")

		//if showPoints {
		//	fmt.Fprintf(w, "%d\t%v\t%v\t%s\t\n", standing.Position, standing.Driver, getSeries(standing.Driver), points, strings.Join(remarks, ","))
		//} else {
		//	fmt.Fprintf(w, "%d\t%v\t%s\t\n", standing.Position, standing.Driver, getSeries(standing.Driver), strings.Join(remarks, ","))
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

func getTeams(drivers []*Driver) ([]*Team, error) {
	var teams []*Team

	teamCSV, err := readCSVFile("teams.csv")
	if err != nil {
		return teams, err
	}
	for i := 0; i < len(teamCSV); i++ {
		t := &Team{
			Name:       teamCSV[i][1],
			HomeSeries: teamCSV[i][0],
		}
		for _, memberId := range teamCSV[i][2:] {
			d, found := getDriverById(drivers, memberId)
			if !found {
				return teams, fmt.Errorf("Driver with id %s for team %s not found", memberId, t.Name)
			}
			t.Drivers = append(t.Drivers, d)
			d.Team = t
		}

		teams = append(teams, t)
	}
	return teams, nil
}

func getDriverById(drivers []*Driver, id string) (d *Driver, found bool) {

	for _, driver := range drivers {

		if driver.id == id {
			return driver, true
		}
	}
	return nil, false
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

		driverId := strings.TrimSpace(resultsCSV[i][0])
		driver, found := getDriverById(getCircus().Drivers, driverId)
		if !found {
			log.Printf("Driver with id %s not found, creating new one with no teams and no series associated\n", driverId)
			driver = &Driver{
				id:         driverId,
				HomeSeries: "",
				Class:      "",
				Team:       nil,
			}
		}
		homeSeries := getSeries(driverId)

		didNotFinish := dnfPosition != 0 && position >= dnfPosition
		earnedPoints := 0
		hasFastestLap := driverId == fastestLap

		if isEligibleToEarnPoints(r.Series, homeSeries) && !didNotFinish {
			earnedPoints = pointScale[pointsPosition]
			pointsPosition += 1

			if hasFastestLap {
				earnedPoints += 2
			}
		}

		r.Standings = append(r.Standings, Position{
			Position:   position,
			Driver:     driver,
			DNF:        didNotFinish,
			Points:     earnedPoints,
			FastestLap: hasFastestLap,
		})
		position += 1
	}

	return r, nil
}

func getCircus() *Circus {

	if circus == nil {

		drivers, err := getDrivers()

		if err != nil {
			log.Fatal(err)
		}

		teams, err := getTeams(drivers)

		if err != nil {
			log.Fatal(err)
		}

		circus = &Circus{
			Drivers: drivers,
			Teams:   teams,
		}

	}
	return circus
}

func getDrivers() ([]*Driver, error) {

	var drivers []*Driver

	teamCSV, err := readCSVFile("einteilung.csv")
	if err != nil {
		return drivers, err
	}
	for i := 0; i < len(teamCSV); i++ {

		class := ""
		if len(teamCSV[i]) >= 3 {
			class = teamCSV[i][2]
		}

		d := &Driver{
			HomeSeries: teamCSV[i][0],
			id:         teamCSV[i][1],
			Class:      class,
		}
		drivers = append(drivers, d)
	}
	return drivers, nil

}

func getTeamResults(r Result) (Result, error) {

	teamResult := Result{
		EventName: r.EventName,
		Series:    r.Series,
		Standings: []Position{},
	}

	teams := circus.Teams

	for _, team := range teams {

		// only get those teams with matching series
		if r.Series != "" && team.HomeSeries != r.Series {
			continue
		}

		// TODO support more than one driver
		teamPoints := getPoints(r.Standings, team.Drivers[0].id) + getPoints(r.Standings, team.Drivers[1].id)

		// do not add teams that have no points
		if teamPoints == 0 {
			continue
		}

		teamResult.Standings = append(teamResult.Standings, Position{
			//Driver:     team.Name,
			Team:       team,
			DNF:        false,
			Points:     teamPoints,
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
		if standing.Driver.id == name {
			return standing.Points
		}
	}
	return 0
}
