# Use Case 01: Vehicle Matriculation (Swedish → Spanish Plates)

## 1. The Scenario (The "Dogfooding" Case)

- **User Profile:** Swedish resident (Lead Dev) moved to Orihuela Costa in September 2025 with a Swedish-registered car.
- **Current Date:** January 2026 (~4 months post-arrival).
- **Problem:** The user is approaching critical legal deadlines and needs to know:
  - How much longer the car can legally stay on Swedish plates.
  - How to avoid the ~12% registration tax (*Impuesto de Matriculación*).
  - The exact steps to convert to Spanish plates (ITV, DGT, SUMA).

---

## 2. Domain Logic & Rules (January 2026 Context)

| Rule | Description |
|------|-------------|
| **183-Day Rule** | Once the user stays 183 days in Spain, they are a tax resident. The car must be registered before this or face fines. |
| **30-Day Padrón Trigger** | From the moment the user registers on the *Padrón* (municipal register), they have 30 days to start the matriculation process. |
| **60-Day Tax Exemption Window** | To avoid registration tax as a "moving good" (*cambio de residencia*), the process must be initiated within 60 days of the official move date. |
| **Local Practice** | ITV stations in Murcia (e.g., San Javier) often have shorter wait times for import inspections than Torrevieja (Alicante). |

---

## 3. Agentic Workflow Requirements

Claude Opus should help design/implement the following logic:

### A. Data Gathering (Tools)

| Tool | Purpose |
|------|---------|
| `CalculateDeadlines(arrivalDate, padronDate)` | Go tool to return hard deadlines. |
| `CheckTaxExemption(carValue, co2Emissions)` | Estimate the *Impuesto de Matriculación* cost if exemption is missed. |
| `SearchLocalITVWaitTimes()` | Use MCP to fetch or simulate current ITV availability. |

### B. Reasoning & Grounding (RAG)

- Query local knowledge base for: *"Requisitos matriculación coche Suecia España 2026"*.
- Contextualize for Orihuela Costa: Identify specific offices (Playa Flamenca City Hall vs. DGT Alicante).

### C. The "Hot Lead" Transition

- **Trigger:** When the agent detects the user is overwhelmed or close to a deadline.
- **Action:** Offer a seamless handoff to a verified *Gestoría* in Orihuela Costa.
- **Vertical AI Goal:** Instead of generic advice, provide a "Pre-check report" PDF that the user can take to the Gestor, or send it directly via API.

---

## 4. Technical Implementation Notes

```
Package:   internal/logic/carimport
Flow:      vehicleImportWorkflow
```

**Objective:** Move from "Inquiry" → "Actionable Schedule" → "Lead Conversion."
