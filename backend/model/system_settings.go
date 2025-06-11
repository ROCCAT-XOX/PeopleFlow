package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GermanState definiert die deutschen Bundesländer
type GermanState string

const (
	StateBadenWuerttemberg     GermanState = "baden_wuerttemberg"
	StateBayern                GermanState = "bayern"
	StateBerlin                GermanState = "berlin"
	StateBrandenburg           GermanState = "brandenburg"
	StateBremen                GermanState = "bremen"
	StateHamburg               GermanState = "hamburg"
	StateHessen                GermanState = "hessen"
	StateMecklenburgVorpommern GermanState = "mecklenburg_vorpommern"
	StateNiedersachsen         GermanState = "niedersachsen"
	StateNordrheinWestfalen    GermanState = "nordrhein_westfalen"
	StateRheinlandPfalz        GermanState = "rheinland_pfalz"
	StateSaarland              GermanState = "saarland"
	StateSachsen               GermanState = "sachsen"
	StateSachsenAnhalt         GermanState = "sachsen_anhalt"
	StateSchleswigHolstein     GermanState = "schleswig_holstein"
	StateThueringen            GermanState = "thueringen"
)

// SystemSettings enthält die globalen Systemeinstellungen
type SystemSettings struct {
	ID                  primitive.ObjectID         `bson:"_id,omitempty" json:"id"`
	CompanyName         string                     `bson:"companyName" json:"companyName"`
	CompanyAddress      string                     `bson:"companyAddress" json:"companyAddress"`
	Language            string                     `bson:"language" json:"language"` // Interface language
	State               string                     `bson:"state" json:"state"` // German state for holiday calculation
	DefaultWorkingHours float64                    `bson:"defaultWorkingHours" json:"defaultWorkingHours"`
	DefaultVacationDays int                        `bson:"defaultVacationDays" json:"defaultVacationDays"`
	EmailNotifications  *EmailNotificationSettings `bson:"emailNotifications,omitempty" json:"emailNotifications,omitempty"`
	CreatedAt           time.Time                  `bson:"createdAt" json:"createdAt"`
	UpdatedAt           time.Time                  `bson:"updatedAt" json:"updatedAt"`
}

// DefaultSystemSettings erstellt Standardeinstellungen
func DefaultSystemSettings() *SystemSettings {
	return &SystemSettings{
		Language:            "de", // Default to German
		State:               string(StateNordrheinWestfalen),
		DefaultWorkingHours: 40,
		DefaultVacationDays: 30,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}
}

// IsValid prüft, ob die GermanState gültig ist
func (gs GermanState) IsValid() bool {
	switch gs {
	case StateBadenWuerttemberg, StateBayern, StateBerlin, StateBrandenburg,
		StateBremen, StateHamburg, StateHessen, StateMecklenburgVorpommern,
		StateNiedersachsen, StateNordrheinWestfalen, StateRheinlandPfalz,
		StateSaarland, StateSachsen, StateSachsenAnhalt, StateSchleswigHolstein,
		StateThueringen:
		return true
	default:
		return false
	}
}

// GetLabel gibt das benutzerfreundliche Label für das Bundesland zurück
func (gs GermanState) GetLabel() string {
	switch gs {
	case StateBadenWuerttemberg:
		return "Baden-Württemberg"
	case StateBayern:
		return "Bayern"
	case StateBerlin:
		return "Berlin"
	case StateBrandenburg:
		return "Brandenburg"
	case StateBremen:
		return "Bremen"
	case StateHamburg:
		return "Hamburg"
	case StateHessen:
		return "Hessen"
	case StateMecklenburgVorpommern:
		return "Mecklenburg-Vorpommern"
	case StateNiedersachsen:
		return "Niedersachsen"
	case StateNordrheinWestfalen:
		return "Nordrhein-Westfalen"
	case StateRheinlandPfalz:
		return "Rheinland-Pfalz"
	case StateSaarland:
		return "Saarland"
	case StateSachsen:
		return "Sachsen"
	case StateSachsenAnhalt:
		return "Sachsen-Anhalt"
	case StateSchleswigHolstein:
		return "Schleswig-Holstein"
	case StateThueringen:
		return "Thüringen"
	default:
		return string(gs)
	}
}

// HasEmailNotifications prüft, ob E-Mail-Benachrichtigungen konfiguriert sind
func (ss *SystemSettings) HasEmailNotifications() bool {
	return ss.EmailNotifications != nil && ss.EmailNotifications.Enabled
}

// IsEmailConfigured prüft, ob die E-Mail-Konfiguration vollständig ist
func (ss *SystemSettings) IsEmailConfigured() bool {
	if !ss.HasEmailNotifications() {
		return false
	}

	en := ss.EmailNotifications
	return en.SMTPHost != "" && en.SMTPPort > 0 &&
		en.SMTPUser != "" && en.SMTPPass != "" &&
		en.FromEmail != "" && en.FromName != ""
}

// GetGermanStates returns all German states with their labels
func GetGermanStates() []map[string]string {
	return []map[string]string{
		{"value": string(StateBadenWuerttemberg), "label": StateBadenWuerttemberg.GetLabel()},
		{"value": string(StateBayern), "label": StateBayern.GetLabel()},
		{"value": string(StateBerlin), "label": StateBerlin.GetLabel()},
		{"value": string(StateBrandenburg), "label": StateBrandenburg.GetLabel()},
		{"value": string(StateBremen), "label": StateBremen.GetLabel()},
		{"value": string(StateHamburg), "label": StateHamburg.GetLabel()},
		{"value": string(StateHessen), "label": StateHessen.GetLabel()},
		{"value": string(StateMecklenburgVorpommern), "label": StateMecklenburgVorpommern.GetLabel()},
		{"value": string(StateNiedersachsen), "label": StateNiedersachsen.GetLabel()},
		{"value": string(StateNordrheinWestfalen), "label": StateNordrheinWestfalen.GetLabel()},
		{"value": string(StateRheinlandPfalz), "label": StateRheinlandPfalz.GetLabel()},
		{"value": string(StateSaarland), "label": StateSaarland.GetLabel()},
		{"value": string(StateSachsen), "label": StateSachsen.GetLabel()},
		{"value": string(StateSachsenAnhalt), "label": StateSachsenAnhalt.GetLabel()},
		{"value": string(StateSchleswigHolstein), "label": StateSchleswigHolstein.GetLabel()},
		{"value": string(StateThueringen), "label": StateThueringen.GetLabel()},
	}
}
