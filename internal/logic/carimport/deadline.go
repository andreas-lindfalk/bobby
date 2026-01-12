package carimport

import "time"

const (
	TaxResidenceDays       = 183
	PadronTriggerDays      = 30
	TaxExemptionDays       = 60
	UrgencyRedThreshold    = 14
	UrgencyYellowThreshold = 30
)

// CalculateDeadlines computes all relevant deadlines for car matriculation.
func CalculateDeadlines(ctx UserContext) DeadlineResult {
	if ctx.CurrentDate.IsZero() {
		ctx.CurrentDate = time.Now()
	}

	result := DeadlineResult{
		TaxResidenceDeadline: ctx.ArrivalDate.AddDate(0, 0, TaxResidenceDays),
		TaxExemptionDeadline: ctx.ArrivalDate.AddDate(0, 0, TaxExemptionDays),
		PadronRegistered:     ctx.PadronDate != nil,
	}

	if ctx.PadronDate != nil {
		padronDeadline := ctx.PadronDate.AddDate(0, 0, PadronTriggerDays)
		result.PadronDeadline = &padronDeadline
	}

	result.TaxExemptionMissed = ctx.CurrentDate.After(result.TaxExemptionDeadline)
	result.EarliestDeadline = findEarliestDeadline(result, ctx.CurrentDate)
	result.DaysRemaining = daysUntil(ctx.CurrentDate, result.EarliestDeadline)
	result.Urgency = calculateUrgency(result.DaysRemaining)

	return result
}

func findEarliestDeadline(result DeadlineResult, now time.Time) time.Time {
	candidates := []time.Time{result.TaxResidenceDeadline}

	if result.PadronDeadline != nil {
		candidates = append(candidates, *result.PadronDeadline)
	}

	var earliest time.Time
	for _, d := range candidates {
		if d.Before(now) {
			continue
		}
		if earliest.IsZero() || d.Before(earliest) {
			earliest = d
		}
	}

	if earliest.IsZero() {
		earliest = result.TaxResidenceDeadline
		if result.PadronDeadline != nil && result.PadronDeadline.After(earliest) {
			earliest = *result.PadronDeadline
		}
	}

	return earliest
}

func daysUntil(from, to time.Time) int {
	from = startOfDay(from)
	to = startOfDay(to)
	duration := to.Sub(from)
	return int(duration.Hours() / 24)
}

func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func calculateUrgency(daysRemaining int) Urgency {
	switch {
	case daysRemaining < UrgencyRedThreshold:
		return UrgencyRed
	case daysRemaining < UrgencyYellowThreshold:
		return UrgencyYellow
	default:
		return UrgencyGreen
	}
}

// IsHotLead returns true if the situation warrants immediate gestoría referral.
func IsHotLead(result DeadlineResult) bool {
	return result.Urgency == UrgencyRed
}
