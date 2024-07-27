package main

import (
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"os"
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

func extractMetaContent(url string) (string, error) {
	// Create a new request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return "", nil
	}

	// Set the User-Agent header
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.0.0 Safari/537.36")

	// Create a client and send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return "", nil
	}
	defer resp.Body.Close()
	defer resp.Body.Close()
	//body, err := io.ReadAll(resp.Body)

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	//body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	//fmt.Println("Response Body:", string(body))

	content := ""
	doc.Find("meta[property=\"og:title\"]").Each(func(i int, s *goquery.Selection) {
		con, _ := s.Attr("content")
		content = strings.TrimSpace(con)
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
	Series    string
}

type Team struct {
	Name    string
	Members []string
	Series  string
}

func formatWhatsApp(r Result) (string, error) {
	rs := fmt.Sprintf("*%s*\nNicht offizielles Ergebnis!\n", r.EventName)
	rs += "```\n"

	buffer := bytes.Buffer{}
	w := tabwriter.NewWriter(&buffer, 10, 1, 1, ' ', tabwriter.Debug)

	fmt.Fprintf(w, "Pos.\tName\tPunkte\tBemerkungen\t\n")
	for _, standing := range r.Standings {
		//rs = rs + fmt.Sprintf("%2d. %s\n", standing.Position, standing.Name)
		points := 0
		var remarks []string
		if standing.FastestLap {
			remarks = append(remarks, "SR")
		}
		if standing.DNF {
			remarks = append(remarks, "DNF")
		}

		fmt.Fprintf(w, "%d\t%v\t%v\t%s\t\n", standing.Position, standing.Name, points, strings.Join(remarks, ","))
	}
	w.Flush()
	rs += buffer.String()

	rs += "```"

	return rs, nil
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
			Series: teamCSV[i][0],
		})
	}
	return teams, nil
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
			DNF:        position >= dnfPosition,
			FastestLap: name == fastestLap,
		})
		position += 1
	}

	return r, nil
}
