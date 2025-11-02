# JCLift 🚀
**Automated JCL Cost & Risk Analyzer**

## 🔎 Problem

Enterprises running IBM mainframes execute thousands—sometimes millions—of JCL jobs daily. These jobs drive mission‑critical processes, yet few teams can explain which jobs are essential, which are wasteful, or what they actually cost. As COBOL/JCL experts retire, mainframe bills rise without visibility or control.

Modernization vendors often spend 12–18 months just *discovering* what the JCL estate does—costing organizations $2M+ before any migration even begins.

---

## ⚙️ What JCLift Does

**JCLift** is a static analysis and cost‑modeling tool for JCL.  
Upload your JCL library, and within hours you’ll see:

- ✅ Job and step dependencies visualized  
- 💸 Estimated CPU/MIPS cost per job and total estate  
- ⚠️ Risk & inefficiency findings (oversized SORTWK, redundant IEBGENER, etc.)  
- 📊 Savings and optimization recommendations  
- 🧭 CI/CD integration to catch inefficiencies before production

---

## 🧩 Core Features

| Capability | Description |
|-------------|--------------|
| **Parsing & Analysis** | Scans thousands of JCL members, builds an Intermediate Representation (IR) of jobs, steps, DDs, and datasets. |
| **Cost Modeling** | Applies a heuristic or calibrated model to estimate CPU/MIPS cost per step. |
| **Risk Detection** | Flags missing restart points, unsafe DISP settings, GDG misuse, and inefficient utilities. |
| **Forecasting** | Predicts operational impact of new JCL before it runs. |
| **Governance & Reporting** | Generates PDF/CSV/HTML reports with projected savings, waiver tracking, and audit logs. |

---

## 💰 Why It Matters

- **Prevent waste before execution.** Detect inefficiencies statically—before jobs run.  
- **Zero‑risk analysis.** No production data touched. Safe for air‑gapped environments.  
- **Continuous guardrail.** Integrates into Jenkins/GitHub CI for every JCL change.  
- **Forecast cost impact.** Estimate MIPS usage and financial impact of changes.  
- **Prove ROI.** Show management how optimization saves real money every billing cycle.

---

## 🧠 Example Finding

**Job:** `LEDGER01`  
**Finding:** SORTWK overallocation by 5× → 30 MIPS wasted  
**Recommendation:** Reduce SORTWK to 5 cylinders  
**Estimated Savings:** ~$75,000/year

---

## 🛠️ Architecture

| Component | Language | Role |
|------------|-----------|------|
| **Analyzer Core** | Go | Parses JCL, builds IR, runs cost/risk rules, exposes REST/CLI. |
| **Plugin Runtime** | Python | Runs extensible customer rules via gRPC. |
| **Rules DSL** | YAML | Declarative thresholds and patterns. |
| **Admin Console** | Vue/Quasar | On‑prem UI: dashboards, findings, waiver workflow, reports. |
| **Database** | PostgreSQL/SQLite | Stores runs, jobs, steps, findings, waivers, calibration data. |

---

## 🧰 CLI Examples

```bash
# Analyze a JCL folder
jclift analyze --path /mnt/jcl --out ./reports/

# Compare runs
jclift diff --base run_2025-10-01 --head run_2025-10-15 --report html

# Calibrate cost model using SMF data
jclift calibrate --smf ./smf.csv
```

---

## 🧩 Integrations

- **CI/CD:** Jenkins, GitHub Actions, Azure DevOps  
- **Export Formats:** JSON, HTML, PDF, XLSX  
- **Auth & Security:** RBAC, audit logs, air‑gapped Docker or binary install

---

## 🧭 Roadmap

| Phase | Deliverables |
|--------|---------------|
| **MVP (8 Weeks)** | Parser, 12 rules, cost v1, CLI, minimal UI |
| **v1.0 (12–14 Weeks)** | Postgres, waivers, planner v1, PDF/XLSX, CI plugins |
| **v2.0 (5–8 Months)** | SMF calibration v2, incremental indexing, worker queue, SSO |

---

## 🏁 Why JCLift Wins

| Factor | Legacy Tools | JCLift |
|---------|---------------|---------|
| Syntax validation | ✅ | ✅ |
| Cost modeling | ❌ | ✅ |
| Risk detection | ⚠️ Limited | ✅ Deep |
| CI/CD prevention | ❌ | ✅ |
| ROI dashboard | ❌ | ✅ |

---

## 📦 Deployment

- **Docker Compose:** core, python‑runner, db, ui  
- **Air‑gapped:** offline tarball + license file  
- **Single binary:** `jclift` CLI for batch and CI use

---

## 🧱 License

Proprietary © 2025 JCLift Technologies.  
All rights reserved. Contact for enterprise licensing and pilots.

---

## 📬 Contact

**Email:** team@jclift.io  
**Website:** [https://jclift.io](https://jclift.io)

> “Cut MIPS spend and eliminate JCL risk — without touching COBOL.”
