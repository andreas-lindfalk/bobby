// Command carimport is a CLI tool for calculating car matriculation deadlines.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/andreas-lindfalk/bobby/internal/flows"
	"github.com/andreas-lindfalk/bobby/internal/logic/carimport"
	"github.com/andreas-lindfalk/bobby/internal/store"
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
	dbURL := flag.String("db", "", "PostgreSQL connection string [optional, enables recommendations]")
	lang := flag.String("lang", "sv,en", "Preferred languages, comma-separated [default: sv,en]")

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

	// Parse languages
	languages := parseLanguages(*lang)

	// Connect to store if DB URL provided
	ctx := context.Background()
	var deps *flows.FlowDependencies
	if *dbURL != "" {
		s, err := store.New(ctx, *dbURL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not connect to database: %v\n", err)
		} else {
			defer s.Close()
			// Run migrations and seed data
			if err := s.MigrateWithSeeds(ctx); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Could not run migrations: %v\n", err)
			}
			deps = &flows.FlowDependencies{Store: s}
			fmt.Println("✓ Connected to database (migrated)")
		}
	}

	// Run the full flow
	input := flows.VehicleImportInput{
		ArrivalDate:  arrivalDate,
		PadronDate:   padronDate,
		CarValue:     *carValue,
		CO2Emissions: *co2,
		IsElectric:   *isElectric,
		IsHybrid:     *isHybrid,
		Languages:    languages,
	}

	output, err := flows.RunVehicleImportFlow(ctx, input, deps)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running flow: %v\n", err)
		os.Exit(1)
	}

	// Print reports
	printDeadlineReport(output.Deadlines, refDate)

	if output.TaxEstimate != nil {
		printTaxReport(*output.TaxEstimate, output.Deadlines.TaxExemptionMissed)
	}

	printActionPlan(output.ActionPlan)

	if output.IsHotLead {
		printHotLeadAlert(output.RecommendedPartner)
	}
}

func parseLanguages(s string) []string {
	var langs []string
	for _, l := range splitAndTrim(s, ',') {
		if l != "" {
			langs = append(langs, l)
		}
	}
	if len(langs) == 0 {
		return []string{"en"}
	}
	return langs
}

func splitAndTrim(s string, sep rune) []string {
	var parts []string
	current := ""
	for _, c := range s {
		if c == sep {
			parts = append(parts, current)
			current = ""
		} else if c != ' ' {
			current += string(c)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

func printDeadlineReport(d carimport.DeadlineResult, refDate time.Time) {
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

func printActionPlan(plan []flows.ActionItem) {
	if len(plan) == 0 {
		return
	}

	fmt.Println()
	fmt.Println("┌─────────────────────────────────────────────────────────────┐")
	fmt.Println("│                    ACTION PLAN                              │")
	fmt.Println("└─────────────────────────────────────────────────────────────┘")

	for _, item := range plan {
		fmt.Printf("\n  %d. %s\n", item.Order, item.Title)
		fmt.Printf("     %s\n", item.Description)
		if item.Location != "" {
			fmt.Printf("     📍 %s\n", item.Location)
		}
		if item.Address != "" {
			fmt.Printf("        %s\n", item.Address)
		}
		if item.Phone != "" {
			fmt.Printf("     📞 %s\n", item.Phone)
		}
		if item.AppointmentURL != "" {
			fmt.Printf("     🔗 %s\n", item.AppointmentURL)
		}
		if item.WaitDays > 0 {
			fmt.Printf("     ⏱️  Avg wait: %d days\n", item.WaitDays)
		}
		if item.Cost != "" {
			fmt.Printf("     💰 %s\n", item.Cost)
		}
	}
}

func printHotLeadAlert(partner *flows.PartnerRecommendation) {
	fmt.Println()
	fmt.Println("┌─────────────────────────────────────────────────────────────┐")
	fmt.Println("│  🔥 HOT LEAD: Immediate action required!                    │")
	fmt.Println("├─────────────────────────────────────────────────────────────┤")

	if partner != nil {
		fmt.Println("│  Recommended Partner:                                       │")
		fmt.Printf("│    %-57s│\n", partner.Name)
		fmt.Printf("│    📍 %-54s│\n", partner.City)
		if partner.Phone != "" {
			fmt.Printf("│    📞 %-54s│\n", partner.Phone)
		}
		if partner.Email != "" {
			fmt.Printf("│    ✉️  %-54s│\n", partner.Email)
		}
		fmt.Printf("│    ⭐ %.1f rating                                            │\n", partner.Rating)
		if partner.Discount != "" {
			fmt.Printf("│    🎁 %-54s│\n", partner.Discount)
		}
	} else {
		fmt.Println("│  Consider connecting with a verified gestoría.             │")
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
