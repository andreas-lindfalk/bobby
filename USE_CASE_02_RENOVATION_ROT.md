# Use Case 02: Managed Renovation & Cross-Border Tax Compliance

## 1. The Scenario (The "Managed Marketplace" Case)

- **User Profile:** Homeowner in Orihuela Costa (Swedish, German, or British).
- **Current Need:** Wants to renovate a kitchen, install solar panels, or build a pool.
- **The Problem:**
  1. **Trust:** Fear of "cowboy builders" or overcharging (*Gringo-tax*).
  2. **Language:** Inability to manage a Spanish construction crew.
  3. **Missed Savings:** Can't claim Swedish ROT/RUT (or German equivalent) because the Spanish contractor doesn't know how to split labor/materials or provide Skatteverket-compliant documentation.

---

## 2. The "Bridge" Logic (Swedish ROT/RUT Focus)

Claude Opus should implement/reason about the following business rules:

| Rule | Description |
|------|-------------|
| **30% Rule** | Labor costs (*mano de obra*) are deductible (up to a limit), but materials (*materiales*) are not. |
| **Invoice Requirements** | A Spanish *Factura Legal* must include Swedish metadata: Contractor's CIF, explicit labor/material split, *fastighetsbeteckning*, and *Personnummer*. |
| **Application Model** | Orihuelacosta.ai acts as the "Service Provider" toward the Swedish Tax Agency, or provides pre-filled documentation for the user. |

---

## 3. Managed Marketplace & Escrow Flow

This is a state-machine logic that Claude Opus needs to scaffold in Go:

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│  1. Quote   │───▶│  2. Escrow  │───▶│ 3. Milestone│───▶│ 4. AI Check │───▶│ 5. Release  │
│   Request   │    │   Deposit   │    │   Upload    │    │  (Vision)   │    │  & Invoice  │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
```

| Step | Description |
|------|-------------|
| **Quoting** | Agent matches user with a "Vetted Contractor" from the internal database. |
| **Escrow Deposit** | User pays the platform (Main Contractor). Money is held in Escrow. |
| **Milestone Tracking** | Contractor uploads progress photos via WhatsApp/App. |
| **AI Verification** | LLM (Vision) checks progress against contract milestones. |
| **Payment Release** | Platform pays contractor (minus margin) and generates tax-compliant invoice. |

---

## 4. Technical Implementation Notes

```
Package:   internal/logic/renovation
           internal/finance/escrow

Core Tool: GenerateComplianceInvoice(countryCode, laborCost, materialCost)
```

**The "Vertical" Advantage:** The system doesn't just *talk* about renovation — it manages the legal and financial bridge between two different national legal systems.

---

## 5. Goal for Claude Opus

| Task | Description |
|------|-------------|
| **Contractor Vetting Schema** | Design schema with CIF verification, insurance checks, and rating history. |
| **Tax Splitting Logic** | Ensure Spanish VAT/IVA is handled correctly while calculating Swedish ROT deduction. |
| **Agentic Nudge** | *"I see you're planning a renovation. If we manage the project, I can guarantee you get your 50,000 SEK tax refund and ensure the builder is fully insured."* |
