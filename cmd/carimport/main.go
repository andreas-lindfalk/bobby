// Command carimport is a CLI tool for calculating car matriculation deadlines.
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/andreas-lindfalk/bobby/internal/logic/carimport"
)

const dateLayout = "2006-01-02"

func main() {
	arrivalStr := flag.String("arrival", "", "Arrival date in Spain (YYYY-MM-DD) [required]")
	padronStr := flag.String("padron", "", "Padrón registration date (YYYY-MM-DD) [optional]")
	carValue := flag.Float64("car-value", 0, "Car value in EUR [optional]")
	co2 := flag.Int("co2", 0, "CO2 emissions in g/km [optional]")
	isElectric := flag.Bool("electric", false, "Is the car electric?")
	isHybrid := flag.Bool("hybrid", false, "Is the car a plug-in hybrid?")
	refDateStr := flag.String("date", "", "Reference date (YYYY-MM-DD) [default: today]")

	flag.Parse()

	if *arrivalStr == "" {
		fmt.Fprintln(os.Stderr, "Error: --arrival is required")
		flag.Usage()
		os.Exit(1)
	}

	arrivalDate, err := time.Parse(dateLayout, *arrivalStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing arrival date: %v\n", err)
		os.Exit(1)
	}

	var padronDate *time.Time
	if *padronStr != "" {
		pd, err := time.Parse(dateLayout, *padronStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing padrón date: %v\n", err)
			os.Exit(1)
		}
		padronDate = &pd
	}

	refDate := time.Now()
	if *refDateStr != "" {
		refDate, err = time.Parse(dateLayout, *refDateStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing reference date: %v\n", err)
			os.Exit(1)
		}
	}

	ctx := carimport.UserContext{
		ArrivalDate: arrivalDate,
		PadronDate:  padronDate,
		CurrentDate: refDate,
	}
	deadlines := carimport.CalculateDeadlines(ctx)

	printReport(deadlines, refDate)

	if *carValue > 0 {
		car := carimport.CarDetails{
			Value:        *carValue,
			CO2Emissions: *co2,
			IsElectric:   *isElectric,
			IsHybrid:     *isHybrid,
		}
		tax := carimport.CalculateTax(car, arrivalDate, refDate)
		printTaxReport(tax, deadlines.TaxExemptionMissed)
	}

	if carimport.IsHotLead(deadlines) {
		fmt.Println()
		fmt.Println("🔥 HOT LEAD: Immediate action required!")
		fmt.Println("   Consider connecting with a verified gestoría in Orihuela Costa.")
	}
}

func printReport(d carimport.DeadlineResult, refDate time.Time) {
	urgencyIcon := map[carimport.Urgency]string{
		carimport.UrgencyGreen:  "🟢",
		carimport.UrgencyYellow: "🟡",
		carimport.UrgencyRed:    "🔴",
	}

	fmt.Println()
	fmt.Println("┌─────────────────────────────────────────────────────────────┐")
	fmt.Println("│                    DEADLINE REPORT                          │")
	fmt.Println("├─────────────────────────────────────────────────────────────┤")
	fmt.Printf("│  Reference Date:        %-36s│\n", refDate.Format(dateLayout))
	fmt.Printf("│  183-Day Tax Residency: %-20s (%3d days)    │\n",
		d.TaxResidenceDeadline.Format(dateLayout), daysFrom(refDate, d.TaxResidenceDeadline))

	if d.PadronDeadline != nil {
		fmt.Printf("│  Padrón 30-Day Limit:   %-20s (%3d days)    │\n",
			d.PadronDeadline.Format(dateLayout), daysFrom(refDate, *d.PadronDeadline))
	} else {
		fmt.Printf("│  Padrón Status:         %-36s│\n", "Not registered")
	}

	exemptionStatus := d.TaxExemptionDeadline.Format(dateLayout)
	if d.TaxExemptionMissed {
		exemptionStatus += " (MISSED)"
	}
	fmt.Printf("│  Tax Exemption Window:  %-36s│\n", exemptionStatus)

	fmt.Println("├─────────────────────────────────────────────────────────────┤")
	fmt.Printf("│  Earliest Deadline:     %-36s│\n", d.EarliestDeadline.Format(dateLayout))
	fmt.Printf("│  Days Remaining:        %-36d│\n", d.DaysRemaining)
	fmt.Printf("│  Urgency:               %s %-34s│\n", urgencyIcon[d.Urgency], string(d.Urgency))
	fmt.Println("└─────────────────────────────────────────────────────────────┘")
}

func printTaxReport(t carimport.TaxEstimate, exemptionMissed bool) {
	fmt.Println()
	fmt.Println("┌─────────────────────────────────────────────────────────────┐")
	fmt.Println("│                    TAX ESTIMATE                             │")
	fmt.Println("├─────────────────────────────────────────────────────────────┤")
	fmt.Printf("│  CO2 Category:          %-36s│\n", truncate(t.CO2Category, 36))
	fmt.Printf("│  Tax Band:              %-35.2f%%│\n", t.TaxBand*100)

	if t.Exempt {
		fmt.Printf("│  Status:                %-36s│\n", "EXEMPT ✓")
		fmt.Printf("│  Reason:                %-36s│\n", truncate(t.ExemptReason, 36))
	} else {
		fmt.Printf("│  Estimated Tax:         €%-35.2f│\n", t.Amount)
	}
	fmt.Println("└─────────────────────────────────────────────────────────────┘")
}

func daysFrom(from, to time.Time) int {
	return int(to.Sub(from).Hours() / 24)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
