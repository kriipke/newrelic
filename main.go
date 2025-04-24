package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

type DashboardConfig struct {
	Widgets []WidgetV3 `yaml:"widgets"`
}

type WidgetV3 struct {
	Title      string     `yaml:"title"`
	Visual     VisualV3   `yaml:"visual"`
	Location   LocationV3 `yaml:"location"`
	Dimensions SizeV3     `yaml:"dimensions"`
	Query      string     `yaml:"query"`
}

type VisualV3 struct {
	Type   string `yaml:"type"`
	Legend bool   `yaml:"legend"`
}

type LocationV3 struct {
	Col int `yaml:"col"`
	Row int `yaml:"row"`
}

type SizeV3 struct {
	Width  int `yaml:"width"`
	Height int `yaml:"height"`
}

// JSON output types (unchanged)
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
	Title             string           `json:"title"`
	Layout            Layout           `json:"layout"`
	Visualization     Visualization    `json:"visualization"`
	LinkedEntityGuids interface{}      `json:"linkedEntityGuids"`
	RawConfiguration  RawConfiguration `json:"rawConfiguration"`
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
	NRQLQueries     []NRQLQuery     `json:"nrqlQueries"`
	PlatformOptions map[string]bool `json:"platformOptions"`
	Legend          *Legend         `json:"legend,omitempty"`
	Facet           *Facet          `json:"facet,omitempty"`
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
	Name                string     `json:"name"`
	Type                string     `json:"type"`
	NRQLQuery           *NRQLQuery `json:"nrqlQuery,omitempty"`
	DefaultValues       []string   `json:"defaultValues,omitempty"`
	ReplacementStrategy string     `json:"replacementStrategy"`
	IsMultiSelection    *bool      `json:"isMultiSelection,omitempty"`
}

func main() {
	var accountID string
	flag.StringVar(&accountID, "account", "", "New Relic account ID to use in queries")
	flag.Parse()

	if accountID == "" {
		fmt.Println("❌ Error: --account argument is required")
		os.Exit(1)
	}

	configPath := "config.yaml"
	configFile, err := os.ReadFile(configPath)
	if err != nil {
		panic(err)
	}

	var cfg DashboardConfig
	if err := yaml.Unmarshal(configFile, &cfg); err != nil {
		panic(err)
	}

	widgets := []Widget{}
	for _, w := range cfg.Widgets {
		widget := Widget{
			Title:         w.Title,
			Layout:        Layout{Column: w.Location.Col, Row: w.Location.Row, Width: w.Dimensions.Width, Height: w.Dimensions.Height},
			Visualization: Visualization{ID: w.Visual.Type},
			RawConfiguration: RawConfiguration{
				NRQLQueries: []NRQLQuery{
					{
						AccountIDs: []string{accountID},
						Query:      w.Query,
					},
				},
				PlatformOptions: map[string]bool{"ignoreTimeRange": false},
			},
		}

		if w.Visual.Legend {
			widget.RawConfiguration.Legend = &Legend{Enabled: true}
			widget.RawConfiguration.Facet = &Facet{ShowOtherSeries: true}
		}

		widgets = append(widgets, widget)
	}

	dashboard := Dashboard{
		Name:        "Service A Platform Overview",
		Permissions: "PUBLIC_READ_WRITE",
		Pages: []Page{
			{
				Name:    "Platform Health",
				Widgets: widgets,
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
				IsMultiSelection:    boolPtr(true),
				ReplacementStrategy: "STRING",
			},
			{
				Name:                "namespace",
				Type:                "STRING",
				DefaultValues:       []string{"REDACTED_NAMESPACE"},
				ReplacementStrategy: "STRING",
			},
			{
				Name:                "clusterName",
				Type:                "STRING",
				DefaultValues:       []string{"REDACTED_CLUSTER"},
				ReplacementStrategy: "STRING",
			},
		},
	}

	outFile := "generated_dashboard_v2.json"
	out, err := os.Create(outFile)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	enc.Encode(dashboard)

	fmt.Printf("✅ Dashboard written to %s\n", outFile)
}

func boolPtr(b bool) *bool {
	return &b
}

