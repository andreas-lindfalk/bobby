// Package seeds contains RAG document seeds for the knowledge base.
package seeds

import "github.com/andreas-lindfalk/bobby/internal/store"

// CarImportDocs returns documents about Swedish car import to Spain.
func CarImportDocs() []store.Document {
	return []store.Document{
		// Legal framework and deadlines
		{
			Source:  "dgt",
			Title:   "183-Day Rule for Vehicle Matriculation",
			Content: `When you become a tax resident in Spain, you have 183 days to register (matriculate) your foreign vehicle with Spanish plates. This deadline starts from your official arrival date or when you register on the padrón municipal. After this period, driving with foreign plates is illegal and can result in fines up to €6,000 and vehicle seizure. The 183-day rule is based on Spanish tax residency laws - once you spend more than 183 days in Spain in a calendar year, you are considered a fiscal resident.`,
			Summary: "Legal deadline for registering foreign vehicles in Spain",
			URL:     "https://sede.dgt.gob.es/es/vehiculos/matriculacion/",
		},
		{
			Source:  "dgt",
			Title:   "30-Day Rule After Padrón Registration",
			Content: `Once you register on the padrón municipal (census registration), you have 30 days to complete your vehicle matriculation. This is a stricter deadline than the 183-day rule. The padrón registration is proof of your residence and triggers this shorter deadline. Many expats are unaware of this rule and face penalties. It's advisable to delay padrón registration until you're ready to start the matriculation process, or begin the process immediately after registering.`,
			Summary: "30-day deadline after padrón registration for vehicle matriculation",
			URL:     "https://sede.dgt.gob.es/es/vehiculos/matriculacion/",
		},

		// ITV Process
		{
			Source:  "itv",
			Title:   "ITV Homologation for Imported Vehicles",
			Content: `Foreign vehicles must pass an ITV (Inspección Técnica de Vehículos) homologation inspection before receiving Spanish plates. This is different from a regular ITV inspection. The homologation verifies that the vehicle meets Spanish/EU safety and emissions standards. Required documents: Certificate of Conformity (COC), original registration document, proof of ownership, passport/NIE. The inspection checks lights, mirrors, speedometer (must show km/h), rear fog light, and emissions. Cost is approximately €150-200. Book appointments early as wait times can be 2-4 weeks.`,
			Summary: "ITV homologation inspection requirements for imported cars",
			URL:     "https://www.itvcita.com/",
		},
		{
			Source:  "itv",
			Title:   "Certificate of Conformity (COC) Requirements",
			Content: `The Certificate of Conformity (COC) or Certificado de Conformidad is essential for vehicle homologation. This document proves your vehicle meets EU type approval standards. For Swedish vehicles, request the COC from the manufacturer or dealer before leaving Sweden. If you don't have a COC, you'll need individual vehicle approval (IVA) which is more expensive (€300-500) and time-consuming. Volvo, SAAB, and other Swedish manufacturers can provide COC documents. Some vehicles over 10 years old may be exempt from COC requirements.`,
			Summary: "COC document requirements for vehicle import",
		},

		// Tax Information
		{
			Source:  "aeat",
			Title:   "Registration Tax (Impuesto de Matriculación)",
			Content: `The Impuesto de Matriculación is a one-time registration tax based on the vehicle's CO2 emissions and value. Tax rates: 0% for emissions ≤120 g/km, 4.75% for 121-159 g/km, 9.75% for 160-199 g/km, 14.75% for ≥200 g/km. The tax is calculated on the vehicle's market value in Spain (not purchase price). Electric vehicles (BEV) and plug-in hybrids with <50 g/km emissions are exempt. Vehicles imported from EU countries as personal property may qualify for reduced rates if owned for more than 6 months.`,
			Summary: "Spanish vehicle registration tax rates and exemptions",
			URL:     "https://sede.agenciatributaria.gob.es/",
		},
		{
			Source:  "aeat",
			Title:   "Electric Vehicle Tax Exemptions in Spain",
			Content: `Electric vehicles (BEV - Battery Electric Vehicles) are fully exempt from the Impuesto de Matriculación in Spain. This represents a significant saving of up to 14.75% of the vehicle's value. Plug-in hybrid vehicles (PHEV) with emissions below 50 g/km also qualify for exemption. Additionally, some municipalities offer reduced IVTM (annual road tax) for electric vehicles. Tesla, Polestar, and other electric vehicles from Sweden qualify for these exemptions. Keep documentation of the vehicle's electric/hybrid status for the registration process.`,
			Summary: "Tax exemptions for electric and hybrid vehicles",
		},
		{
			Source:  "suma",
			Title:   "IVTM Annual Road Tax (Impuesto de Vehículos)",
			Content: `The IVTM (Impuesto sobre Vehículos de Tracción Mecánica) is the annual road tax paid to your municipality. In Orihuela Costa/Torrevieja area, rates range from €60-150 depending on engine power. Payment is due annually, typically in spring. New vehicle registrations require payment of a prorated amount for the remaining year. SUMA is the tax collection agency for Alicante province. You can pay online, at banks, or at SUMA offices. Keep proof of payment as it's required for ITV inspections.`,
			Summary: "Annual road tax information for Alicante province",
			URL:     "https://www.suma.es/",
		},

		// Swedish-specific
		{
			Source:  "transportstyrelsen",
			Title:   "Deregistering Your Vehicle in Sweden",
			Content: `Before or after registering your vehicle in Spain, you must deregister it in Sweden with Transportstyrelsen. You can do this online at transportstyrelsen.se or by submitting form "Anmälan om avregistrering". Required information: registration number, personal number, reason for export. The Swedish registration plates should be removed but can be kept as souvenirs. You'll receive confirmation of deregistration which may be requested by Spanish authorities. Insurance in Sweden will be automatically cancelled upon deregistration.`,
			Summary: "How to deregister your car in Sweden before Spanish registration",
			URL:     "https://www.transportstyrelsen.se/",
		},
		{
			Source:  "skatteverket",
			Title:   "Swedish ROT Deduction for Spanish Property Work",
			Content: `Swedish tax residents can claim ROT (Repairs, Conversion, Extension) deductions for work on properties they own, including in Spain. The deduction is 30% of labor costs, up to 50,000 SEK per person per year. The work must be performed by a Swedish company (F-skatt registered). This applies to renovations, repairs, and extensions - not new construction. Keep detailed invoices showing labor vs material costs. Claim the deduction in your Swedish tax return. Note: You must still be a Swedish tax resident to claim.`,
			Summary: "Swedish ROT tax deduction for property work abroad",
			URL:     "https://www.skatteverket.se/rot",
		},

		// Process guides
		{
			Source: "manual",
			Title:  "Complete Guide: Swedish Car Import to Spain",
			Content: `Step-by-step process for importing your Swedish vehicle to Spain:
1. BEFORE LEAVING SWEDEN: Obtain Certificate of Conformity (COC) from manufacturer, get vehicle history report from Transportstyrelsen, consider timing relative to padrón registration.
2. ARRIVAL IN SPAIN: Get NIE if you don't have one, find accommodation, decide on padrón registration timing.
3. PADRÓN REGISTRATION: Register at your local Ayuntamiento - this starts the 30-day countdown.
4. ITV APPOINTMENT: Book ITV homologation inspection (expect 2-4 week wait), gather documents.
5. ITV INSPECTION: Bring COC, registration document, passport, NIE. Cost ~€150-200.
6. TAX PAYMENT: Pay Impuesto de Matriculación at AEAT or through gestoría. Electric vehicles exempt.
7. DGT REGISTRATION: Submit application for Spanish plates at Jefatura de Tráfico.
8. RECEIVE PLATES: Pick up or receive Spanish plates, install on vehicle.
9. SWEDISH DEREGISTRATION: Notify Transportstyrelsen of export.
10. INSURANCE: Cancel Swedish insurance, arrange Spanish coverage.
Total typical timeline: 4-8 weeks. Cost: €500-2000 depending on vehicle value and emissions.`,
			Summary: "Complete step-by-step guide for Swedish car import to Spain",
		},
		{
			Source:  "manual",
			Title:   "Recommended Gestoría Services for Vehicle Import",
			Content: `A gestoría is an administrative agency that handles bureaucratic processes in Spain. For vehicle import, a gestoría can manage the entire process for €200-400, saving you multiple trips and paperwork headaches. They handle: ITV appointment booking, tax calculations and payment, DGT registration, document translation if needed. Recommended for: people with limited Spanish, those with time constraints, complex cases (missing COC, older vehicles). In Orihuela Costa, look for gestorías that specifically advertise 'matriculación de vehículos extranjeros' and speak Swedish or English.`,
			Summary: "Using a gestoría for vehicle import administration",
		},

		// Common issues
		{
			Source: "manual",
			Title:  "Common Problems with Swedish Vehicle Imports",
			Content: `Frequent issues encountered when importing Swedish cars to Spain:
1. MISSING COC: If you didn't get the Certificate of Conformity before leaving Sweden, contact the manufacturer directly. Volvo, SAAB dealers can help. Without COC, you need expensive individual approval.
2. HEADLIGHTS: Swedish vehicles often have asymmetric headlight patterns for right-side driving. May need adjustment or replacement for ITV approval. Cost: €50-200.
3. REAR FOG LIGHT: Spain requires a rear fog light on the left side. Some Swedish vehicles have it on the right. May need modification.
4. SPEEDOMETER: Must display km/h prominently. Most modern Swedish cars already do this.
5. EMISSIONS: Older vehicles may fail emissions testing. Consider having the car serviced before ITV.
6. TIMING: Don't register on padrón until ready to start the process - the 30-day clock starts immediately.`,
			Summary: "Common problems and solutions for Swedish car imports",
		},

		// Local information
		{
			Source: "local",
			Title:  "ITV Stations Near Orihuela Costa",
			Content: `ITV stations serving the Orihuela Costa area:
1. ITV TORREVIEJA - Av. Cortes Valencianas, 03181 Torrevieja. Phone: 966 700 706. Typically shortest wait times (1-2 weeks). Handles import homologation.
2. ITV ORIHUELA - Ctra. de Callosa, 03300 Orihuela. Phone: 965 303 268. Larger facility but further from coast.
3. ITV ELCHE - Polígono Industrial Altabix. Phone: 966 612 200. Good alternative if others are fully booked.
4. ITV ALICANTE - Multiple locations in the city. Good for urgent cases.
All stations accept online booking. Import inspections are typically scheduled in the morning. Arrive 15 minutes early with all documents.`,
			Summary: "ITV station locations near Orihuela Costa",
		},
		{
			Source: "local",
			Title:  "DGT Office for Vehicle Registration - Alicante",
			Content: `Jefatura Provincial de Tráfico de Alicante handles all vehicle registrations for Alicante province including Orihuela Costa, Torrevieja, and surrounding areas.
Address: C/ Churruca 24, 03003 Alicante
Phone: 965 935 800
Hours: Monday-Friday 9:00-14:00
Appointments: Required - book at sede.dgt.gob.es
Documents needed: ITV approval, tax payment receipt, original foreign registration, NIE, padrón certificate, insurance (at least third-party).
Tip: Appointments can be scarce. Book 2-3 weeks in advance. Alternatively, a gestoría can submit on your behalf without you attending.`,
			Summary: "DGT traffic office information for Alicante province",
			URL:     "https://sede.dgt.gob.es/",
		},
	}
}
