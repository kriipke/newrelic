package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Dashboard struct {
	Name        string     `json:"name"`
	Permissions string     `json:"permissions"`
	Pages       []Page     `json:"pages"`
	Variables   []Variable `json:"variables"`
}

type Page struct {
	Name    string   `json:"name"`
	Widgets []Widget `json:"widgets"`
}

type Widget struct {
	Title               string             `json:"title"`
	Layout              Layout             `json:"layout"`
	Visualization       Visualization      `json:"visualization"`
	LinkedEntityGuids   interface{}        `json:"linkedEntityGuids"`
	RawConfiguration    RawConfiguration   `json:"rawConfiguration"`
}

type Layout struct {
	Column int `json:"column"`
	Row    int `json:"row"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

type Visualization struct {
	ID string `json:"id"`
}

type RawConfiguration struct {
	NRQLQueries     []NRQLQuery         `json:"nrqlQueries"`
	PlatformOptions map[string]bool     `json:"platformOptions"`
	Legend          *Legend             `json:"legend,omitempty"`
	Facet           *Facet              `json:"facet,omitempty"`
}

type NRQLQuery struct {
	AccountIDs []string `json:"accountIds"`
	Query      string   `json:"query"`
}

type Legend struct {
	Enabled bool `json:"enabled"`
}

type Facet struct {
	ShowOtherSeries bool `json:"showOtherSeries"`
}

type Variable struct {
	Name               string         `json:"name"`
	Type               string         `json:"type"`
	NRQLQuery          *NRQLQuery     `json:"nrqlQuery,omitempty"`
	DefaultValues      []string       `json:"defaultValues,omitempty"`
	ReplacementStrategy string        `json:"replacementStrategy"`
	IsMultiSelection   *bool          `json:"isMultiSelection,omitempty"`
}

func main() {
	accountID := "REDACTED_ACCOUNT_ID"

	dashboard := Dashboard{
		Name:        "Service A Platform Overview",
		Permissions: "PUBLIC_READ_WRITE",
		Pages: []Page{
			{
				Name: "Platform Health",
				Widgets: []Widget{
					makeWidget("Deployments Over Time", "viz.line", 1, 7, 3, 2,
						fmt.Sprintf("FROM Deployment SELECT count(*) FACET appName SINCE 1 day ago TIMESERIES"), accountID),
					makeWidget("Errors by Deployment", "viz.pie", 4, 7, 3, 2,
						fmt.Sprintf("FROM TransactionError SELECT count(*) FACET deploymentName SINCE 30 minutes ago"), accountID, true),
					makeWidget("Top URIs (1h)", "viz.pie", 1, 9, 3, 2,
						fmt.Sprintf("FROM Transaction SELECT count(*) FACET request.uri LIMIT 10 SINCE 1 hour ago"), accountID, true),
					makeWidget("Request Volume by App", "viz.stacked-bar", 4, 9, 3, 2,
						fmt.Sprintf("FROM Metric SELECT sum(apm.service.overview.web) FACET appName TIMESERIES SINCE 30 minutes ago"), accountID, false),
				},
			},
		},
		Variables: []Variable{
			{
				Name: "appName",
				Type: "NRQL",
				NRQLQuery: &NRQLQuery{
					AccountIDs: []string{accountID},
					Query:      "SELECT uniques(appName) FROM Transaction SINCE 30 minutes ago",
				},
				IsMultiSelection:   boolPtr(true),
				ReplacementStrategy: "STRING",
			},
			{
				Name:               "namespace",
				Type:               "STRING",
				DefaultValues:      []string{"REDACTED_NAMESPACE"},
				ReplacementStrategy: "STRING",
			},
			{
				Name:               "clusterName",
				Type:               "STRING",
				DefaultValues:      []string{"REDACTED_CLUSTER"},
				ReplacementStrategy: "STRING",
			},
		},
	}

	// Output to file
	file, _ := os.Create("dashboard.json")
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	_ = encoder.Encode(dashboard)
	fmt.Println("✅ Dashboard JSON written to dashboard.json")
}

func makeWidget(title, viz string, col, row, width, height int, query string, accountID string, extras ...bool) Widget {
	w := Widget{
		Title:         title,
		Layout:        Layout{Column: col, Row: row, Width: width, Height: height},
		Visualization: Visualization{ID: viz},
		RawConfiguration: RawConfiguration{
			NRQLQueries: []NRQLQuery{
				{
					AccountIDs: []string{accountID},
					Query:      query,
				},
			},
			PlatformOptions: map[string]bool{"ignoreTimeRange": false},
		},
	}

	if len(extras) > 0 && extras[0] {
		w.RawConfiguration.Legend = &Legend{Enabled: true}
		w.RawConfiguration.Facet = &Facet{ShowOtherSeries: true}
	} else if viz == "viz.stacked-bar" {
		w.RawConfiguration.Legend = &Legend{Enabled: true}
		w.RawConfiguration.Facet = &Facet{ShowOtherSeries: false}
	}

	return w
}

func boolPtr(b bool) *bool {
	return &b
}

