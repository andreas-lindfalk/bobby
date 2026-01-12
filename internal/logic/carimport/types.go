// Package carimport provides domain logic for vehicle matriculation
// from Swedish to Spanish plates, including deadline calculations
// and registration tax estimation.
package carimport

import "time"

// Urgency levels for deadline status
type Urgency string

const (
	UrgencyGreen  Urgency = "green"
	UrgencyYellow Urgency = "yellow"
	UrgencyRed    Urgency = "red"
)

// DeadlineResult contains calculated deadlines for car matriculation.
type DeadlineResult struct {
	TaxResidenceDeadline time.Time  `json:"tax_residence_deadline"`
	PadronDeadline       *time.Time `json:"padron_deadline,omitempty"`
	TaxExemptionDeadline time.Time  `json:"tax_exemption_deadline"`
	EarliestDeadline     time.Time  `json:"earliest_deadline"`
	DaysRemaining        int        `json:"days_remaining"`
	TaxExemptionMissed   bool       `json:"tax_exemption_missed"`
	Urgency              Urgency    `json:"urgency"`
	PadronRegistered     bool       `json:"padron_registered"`
}

// TaxEstimate contains the estimated registration tax.
type TaxEstimate struct {
	Amount       float64 `json:"amount"`
	Exempt       bool    `json:"exempt"`
	ExemptReason string  `json:"exempt_reason,omitempty"`
	TaxBand      float64 `json:"tax_band"`
	CO2Category  string  `json:"co2_category"`
}

// CarDetails contains information about the vehicle for tax calculation.
type CarDetails struct {
	Value        float64 `json:"value"`
	CO2Emissions int     `json:"co2_emissions"`
	IsElectric   bool    `json:"is_electric"`
	IsHybrid     bool    `json:"is_hybrid"`
}

// UserContext contains the user's situation for deadline calculation.
type UserContext struct {
	ArrivalDate time.Time  `json:"arrival_date"`
	PadronDate  *time.Time `json:"padron_date,omitempty"`
	CurrentDate time.Time  `json:"current_date"`
}
