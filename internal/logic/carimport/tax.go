package carimport

import "time"

const (
	CO2Band0Max = 120
	CO2Band1Max = 159
	CO2Band1Pct = 0.0475
	CO2Band2Max = 199
	CO2Band2Pct = 0.0975
	CO2Band3Pct = 0.1475
)

// CalculateTax estimates the registration tax (Impuesto de Matriculación).
func CalculateTax(car CarDetails, arrivalDate time.Time, currentDate time.Time) TaxEstimate {
	if car.IsElectric {
		return TaxEstimate{
			Amount:       0,
			Exempt:       true,
			ExemptReason: "Electric vehicles are exempt from registration tax",
			TaxBand:      0,
			CO2Category:  "Electric (0 g/km)",
		}
	}

	taxBand, co2Category := getCO2Band(car.CO2Emissions)

	exemptionDeadline := arrivalDate.AddDate(0, 0, TaxExemptionDays)
	if currentDate.Before(exemptionDeadline) || currentDate.Equal(exemptionDeadline) {
		return TaxEstimate{
			Amount:       0,
			Exempt:       true,
			ExemptReason: "Cambio de residencia: process initiated within 60 days of arrival",
			TaxBand:      taxBand,
			CO2Category:  co2Category,
		}
	}

	amount := car.Value * taxBand

	if car.IsHybrid && taxBand > 0 {
		amount = amount * 0.25
	}

	return TaxEstimate{
		Amount:      amount,
		Exempt:      false,
		TaxBand:     taxBand,
		CO2Category: co2Category,
	}
}

func getCO2Band(co2 int) (float64, string) {
	switch {
	case co2 <= CO2Band0Max:
		return 0, "Low emissions (≤120 g/km) - 0% tax"
	case co2 <= CO2Band1Max:
		return CO2Band1Pct, "Medium-low emissions (121-159 g/km) - 4.75% tax"
	case co2 <= CO2Band2Max:
		return CO2Band2Pct, "Medium-high emissions (160-199 g/km) - 9.75% tax"
	default:
		return CO2Band3Pct, "High emissions (≥200 g/km) - 14.75% tax"
	}
}

// EstimateTaxWithoutExemption calculates what the tax would be if exemption is lost.
func EstimateTaxWithoutExemption(car CarDetails) TaxEstimate {
	if car.IsElectric {
		return TaxEstimate{
			Amount:       0,
			Exempt:       true,
			ExemptReason: "Electric vehicles are always exempt",
			TaxBand:      0,
			CO2Category:  "Electric (0 g/km)",
		}
	}

	taxBand, co2Category := getCO2Band(car.CO2Emissions)
	amount := car.Value * taxBand

	if car.IsHybrid && taxBand > 0 {
		amount = amount * 0.25
	}

	return TaxEstimate{
		Amount:      amount,
		Exempt:      false,
		TaxBand:     taxBand,
		CO2Category: co2Category,
	}
}
