package main

import (
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/jedib0t/go-pretty/v6/table"
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

const AM = "AM"
const PROAM = "PROAM"
const PRO = "PRO"

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
	showPoints := true
	whatsAppMessage, err := formatWhatsApp(individualResults, teamResults, showPoints, true, true)

	if err != nil {
		return
	}

	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(whatsAppMessage)

	forumString, err := formatForum(individualResults)
	if err != nil {
		log.Fatal(err)
	}
	f, err := os.Create("test_formatWeb.html")
	defer f.Close()
	_, err = f.WriteString(forumString)

}

func formatForum(results Result) (out string, err error) {

	out += "<h2>Inofizielles Ergebnis " + results.EventName + "</h2>"
	//out += "<p>Abgetippt von snimat und erstellt mittels <a href=\"https://github.com/snipem/rrl_results\">https://github.com/snipem/rrl_results</a></p>"

	out += "<h3>Endergebnis</h3>"

	webIndividual, err := formatWebIndividual(results)
	if err != nil {
		return "", err
	}
	out += webIndividual

	out += "<h3>Teamergebnis</h3>"
	teamresult, err := getTeamResults_new(results)
	webTeams, err := formatWebTeams(teamresult)
	if err != nil {
		return "", err
	}

	out += webTeams

	out += "<h3>Klassenergebnisse</h3>"

	for _, s := range []string{"PRO", "PROAM", "AM"} {
		out += fmt.Sprintf("<h4>%s</h4>", s)
		classResult := getClassResult(results, s)
		formatedIndividual, err := formatWebIndividual(classResult)
		if err != nil {
			return "", err
		}
		out += formatedIndividual
	}
	htmlPage := generateHTMLPage(out)
	return htmlPage, nil

}

// generateHTMLPage generates an HTML page with the given body content.
func generateHTMLPage(bodyContent string) string {
	htmlTemplate := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Document</title>
</head>
<body>
    %s
</body>
</html>`
	return fmt.Sprintf(htmlTemplate, bodyContent)
}

func getTeamResults_new(result Result) (teamResult Teamresult, err error) {
	teamResult = Teamresult{
		EventName: result.EventName,
		Series:    result.Series,
		Standings: []TeamPosition{},
	}

	teams := circus.Teams

	for _, team := range teams {

		// only get those teams with matching series
		if result.Series != "" && team.HomeSeries != result.Series {
			continue
		}

		// TODO support more than one driver
		teamPoints := 0
		if len(team.Drivers) == 2 {
			teamPoints = getPoints(result.Standings, team.Drivers[0].Id) + getPoints(result.Standings, team.Drivers[1].Id)
		} else if len(team.Drivers) == 1 {
			teamPoints = getPoints(result.Standings, team.Drivers[0].Id)
		}

		// do not add teams that have no points
		if teamPoints == 0 {
			continue
		}

		teamResult.Standings = append(teamResult.Standings, TeamPosition{
			Team:   team,
			Points: teamPoints,
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

func formatWebTeams(teamresult Teamresult) (out string, err error) {
	t := getTableWriter()

	t.AppendHeader(table.Row{"Pos.", "Teamname", "Serie", "Klasse", "Punkte"})

	for _, standing := range teamresult.Standings {
		pointsRendered := ""
		if standing.Points > 0 {
			pointsRendered = fmt.Sprintf("+%d", standing.Points)
		}

		t.AppendRow(table.Row{
			standing.Position, standing.Team.Name, standing.Team.HomeSeries, standing.Team.Class, pointsRendered,
		})

	}
	return t.RenderHTML(), nil

}

type Driver struct {
	Id         string
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
	Class     string
}

type TeamPosition struct {
	Position int
	Team     *Team
	Points   int
}

type Teamresult struct {
	Standings []TeamPosition
	URL       string
	EventName string
	Series    string
	Class     string
}

type Team struct {
	Name       string
	Drivers    []*Driver
	HomeSeries string
	Class      string
}

func (t *Team) String() string {
	return fmt.Sprintf("%s [%s, %s]", t.Name, t.Drivers[0].Id, t.Drivers[1].Id)
}

type Circus struct {
	Drivers []*Driver
	Teams   []*Team
}

func formatWebIndividual(result Result) (string, error) {

	t := getTableWriter()

	t.AppendHeader(table.Row{"Pos.", "Name", "Team", "Info", "Serie", "Klasse", "Punkte"})

	for _, standing := range result.Standings {
		remarks := getRemarks(result, standing)
		pointsRendered := ""
		if standing.Points > 0 {
			pointsRendered = fmt.Sprintf("+%d", standing.Points)
		}

		teamName := ""
		if standing.Driver.Team != nil {
			teamName = standing.Driver.Team.Name
		}

		t.AppendRow(table.Row{
			standing.Position, standing.Driver.Id, teamName, strings.Join(remarks, ", "), standing.Driver.HomeSeries, standing.Driver.Class, pointsRendered,
		})

	}
	return t.RenderHTML(), nil
}

func getTableWriter() table.Writer {
	t := table.NewWriter()

	t.Style().HTML = table.HTMLOptions{
		CSSClass:    "rrl_results",
		EmptyColumn: "&nbsp;",
		EscapeText:  true,
		Newline:     "<br/>",
	}
	return t
}

func formatWhatsApp(individualResults Result, teamResults Result, showPoints bool, showSeries bool, showClass bool) (string, error) {

	showPosition := true

	rs := fmt.Sprintf("*%s*\nInoffizielles Ergebnis!\n", individualResults.EventName)
	rs += "```\n"

	rs += getAsciiTable(individualResults, showPosition, showPoints, showSeries, showClass)
	rs += "```"
	rs += "\n"
	rs += "Teams nach Punkten:\n"
	rs += "```\n"
	rs += getAsciiTable(teamResults, showPosition, showPoints, false, false)
	rs += "```\n"

	for i, s := range []string{"PRO", "PROAM", "AM"} {
		rs += fmt.Sprintf("Ergebnis %s-Klasse:\n", s)
		rs += "```\n"
		rs += getAsciiTable(getClassResult(individualResults, s), showPosition, showPoints, false, false)
		if i < 2 {
			rs += "\n"
		}
		rs += "```\n"

	}

	return rs, nil
}

func getAsciiTable(r Result, showPosition bool, showPoints bool, showSeries bool, showClass bool) string {
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
	if showClass {
		fmt.Fprintf(w, "C\t")
	}

	fmt.Fprintf(w, "\n")

	for _, standing := range r.Standings {
		//rs = rs + fmt.Sprintf("%2d. %s\n", standing.Position, standing.Driver)
		remarks := getRemarks(r, standing)

		if showPosition {
			fmt.Fprintf(w, "%d\t", standing.Position)
		}
		if standing.Driver != nil {
			fmt.Fprintf(w, "%v\t", standing.Driver.Id)
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
		if standing.Driver != nil && showClass {
			if standing.Driver.HomeSeries == r.Series {
				fmt.Fprintf(w, "%v\t", standing.Driver.Class)
			} else {
				fmt.Fprintf(w, "%v\t", "")
			}
		} else if standing.Team == nil && showClass {
			fmt.Fprintf(w, "%v\t", standing.Team.Class)
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

func getRemarks(r Result, standing Position) []string {
	var remarks []string
	if standing.FastestLap {
		remarks = append(remarks, "SR")
	}
	if standing.DNF {
		remarks = append(remarks, "DNF")
	}

	// This is for the individual table
	if standing.Driver != nil && standing.Driver.HomeSeries != r.Series {
		remarks = append(remarks, "EF")
	}
	return remarks
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
		return teams, fmt.Errorf("failed to read teams.csv: %w", err)
	}
	for i := 0; i < len(teamCSV); i++ {
		t := &Team{
			HomeSeries: teamCSV[i][0],
			Name:       teamCSV[i][1],
			Class:      teamCSV[i][4],
		}
		for _, memberId := range teamCSV[i][2:4] {
			if memberId == "" {
				continue
			}

			d, found := getDriverById(drivers, memberId)
			if !found {
				return teams, fmt.Errorf("driver with Id %s for team %s not found", memberId, t.Name)
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

		if driver.Id == id {
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
			log.Printf("Driver with Id %s not found, creating new one with no teams and no series associated\n", driverId)
			driver = &Driver{
				Id:         driverId,
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
			log.Fatal(fmt.Sprintf("getCircus -> getTeams: %s", err))
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
		return drivers, fmt.Errorf("getDrivers -> readCSVFile: %s", err)
	}
	for i := 0; i < len(teamCSV); i++ {

		if len(teamCSV[i]) != 3 {
			return drivers, fmt.Errorf("incomplete driver line: %s", teamCSV[i])
		}

		homeSeries := teamCSV[i][0]
		id := teamCSV[i][1]
		class := teamCSV[i][2]

		d := &Driver{
			HomeSeries: homeSeries,
			Id:         id,
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
		teamPoints := 0
		if len(team.Drivers) == 2 {
			teamPoints = getPoints(r.Standings, team.Drivers[0].Id) + getPoints(r.Standings, team.Drivers[1].Id)
		} else if len(team.Drivers) == 1 {
			teamPoints = getPoints(r.Standings, team.Drivers[0].Id)
		}

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
		if standing.Driver.Id == name {
			return standing.Points
		}
	}
	return 0
}

// getClassResult returns the result for the given class and the series of the event
func getClassResult(r Result, class string) (classresult Result) {

	classresult = Result{
		EventName: r.EventName,
		Class:     class,
		Series:    r.Series,
		Standings: []Position{},
	}

	for i := 0; i < len(r.Standings); i++ {
		if r.Standings[i].Driver.Class == class && r.Standings[i].Driver.HomeSeries == r.Series {
			classresult.Standings = append(classresult.Standings, r.Standings[i])
		}
	}
	return classresult

}
