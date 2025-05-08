package domain

type Map struct {
	UUID                    string    `json:"uuid"`
	DisplayName             string    `json:"displayName"`
	NarrativeDescription    *string   `json:"narrativeDescription"`
	TacticalDescription     string    `json:"tacticalDescription"`
	Coordinates             string    `json:"coordinates"`
	DisplayIcon             string    `json:"displayIcon"`
	ListViewIcon            string    `json:"listViewIcon"`
	ListViewIconTall        string    `json:"listViewIconTall"`
	Splash                  string    `json:"splash"`
	StylizedBackgroundImage string    `json:"stylizedBackgroundImage"`
	PremierBackgroundImage  string    `json:"premierBackgroundImage"`
	AssetPath               string    `json:"assetPath"`
	MapUrl                  string    `json:"mapUrl"`
	XMultiplier             float64   `json:"xMultiplier"`
	YMultiplier             float64   `json:"yMultiplier"`
	XScalarToAdd            float64   `json:"xScalarToAdd"`
	YScalarToAdd            float64   `json:"yScalarToAdd"`
	Callouts                []Callout `json:"callouts,omitempty"`
}

type Location struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type Callout struct {
	RegionName      string   `json:"regionName"`
	SuperRegionName string   `json:"superRegionName"`
	Location        Location `json:"location"`
}
