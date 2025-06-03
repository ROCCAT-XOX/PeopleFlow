// backend/model/system_settings.go
package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// SystemSettings repräsentiert die systemweiten Einstellungen
type SystemSettings struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CompanyName string             `bson:"companyName" json:"companyName"`
	Language    string             `bson:"language" json:"language"`
	State       string             `bson:"state" json:"state"` // Bundesland für Feiertage
	Timezone    string             `bson:"timezone" json:"timezone"`
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// GermanState repräsentiert die deutschen Bundesländer
type GermanState string

const (
	StateDefault               GermanState = ""
	StateBadenWuerttemberg     GermanState = "BW" // Baden-Württemberg
	StateBayern                GermanState = "BY" // Bayern
	StateBerlin                GermanState = "BE" // Berlin
	StateBrandenburg           GermanState = "BB" // Brandenburg
	StateBremen                GermanState = "HB" // Bremen
	StateHamburg               GermanState = "HH" // Hamburg
	StateHessen                GermanState = "HE" // Hessen
	StateMecklenburgVorpommern GermanState = "MV" // Mecklenburg-Vorpommern
	StateNiedersachsen         GermanState = "NI" // Niedersachsen
	StateNordrheinWestfalen    GermanState = "NW" // Nordrhein-Westfalen
	StateRheinlandPfalz        GermanState = "RP" // Rheinland-Pfalz
	StateSaarland              GermanState = "SL" // Saarland
	StateSachsen               GermanState = "SN" // Sachsen
	StateSachsenAnhalt         GermanState = "ST" // Sachsen-Anhalt
	StateSchleswig             GermanState = "SH" // Schleswig-Holstein
	StateThueringen            GermanState = "TH" // Thüringen
)

// GetGermanStates gibt alle deutschen Bundesländer zurück
func GetGermanStates() map[GermanState]string {
	return map[GermanState]string{
		StateBadenWuerttemberg:     "Baden-Württemberg",
		StateBayern:                "Bayern",
		StateBerlin:                "Berlin",
		StateBrandenburg:           "Brandenburg",
		StateBremen:                "Bremen",
		StateHamburg:               "Hamburg",
		StateHessen:                "Hessen",
		StateMecklenburgVorpommern: "Mecklenburg-Vorpommern",
		StateNiedersachsen:         "Niedersachsen",
		StateNordrheinWestfalen:    "Nordrhein-Westfalen",
		StateRheinlandPfalz:        "Rheinland-Pfalz",
		StateSaarland:              "Saarland",
		StateSachsen:               "Sachsen",
		StateSachsenAnhalt:         "Sachsen-Anhalt",
		StateSchleswig:             "Schleswig-Holstein",
		StateThueringen:            "Thüringen",
	}
}

// GetDisplayName gibt den deutschen Namen des Bundeslandes zurück
func (s GermanState) GetDisplayName() string {
	states := GetGermanStates()
	if name, exists := states[s]; exists {
		return name
	}
	return string(s)
}

// DefaultSystemSettings gibt die Standard-Systemeinstellungen zurück
func DefaultSystemSettings() *SystemSettings {
	return &SystemSettings{
		CompanyName: "PeopleFlow GmbH",
		Language:    "de",
		State:       string(StateNordrheinWestfalen), // NRW als Standard
		Timezone:    "Europe/Berlin",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}
