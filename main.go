
package main

import (
    "encoding/json"
    "fmt"
    "gopkg.in/yaml.v3"
    "os"
)

type WidgetSpec struct {
    Title  string `yaml:"title"`
    Viz    string `yaml:"viz"`
    Column int    `yaml:"column"`
    Row    int    `yaml:"row"`
    Width  int    `yaml:"width"`
    Height int    `yaml:"height"`
    Query  string `yaml:"query"`
    Legend bool   `yaml:"legend"`
}

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
    NRQLQueries     []NRQLQuery       `json:"nrqlQueries"`
    PlatformOptions map[string]bool   `json:"platformOptions"`
    Legend          *Legend           `json:"legend,omitempty"`
    Facet           *Facet            `json:"facet,omitempty"`
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
    yamlFile, err := os.ReadFile("config.yaml")
    if err != nil {
        panic(err)
    }

    var data struct {
        Widgets []WidgetSpec `yaml:"widgets"`
    }

    if err := yaml.Unmarshal(yamlFile, &data); err != nil {
        panic(err)
    }

    accountID := "REDACTED_ACCOUNT_ID"
    widgets := make([]Widget, 0, len(data.Widgets))
    for _, w := range data.Widgets {
        widget := Widget{
            Title: w.Title,
            Layout: Layout{
                Column: w.Column,
                Row:    w.Row,
                Width:  w.Width,
                Height: w.Height,
            },
            Visualization:     Visualization{ID: w.Viz},
            LinkedEntityGuids: nil,
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
        if w.Legend {
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

    out, err := os.Create("generated_dashboard.json")
    if err != nil {
        panic(err)
    }
    defer out.Close()

    enc := json.NewEncoder(out)
    enc.SetIndent("", "  ")
    enc.Encode(dashboard)

    fmt.Println("✅ Dashboard written to generated_dashboard.json")
}

func boolPtr(b bool) *bool {
    return &b
}
