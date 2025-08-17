# Pilot Runbook

## Audience

Batch operations and platform team during pilot.

## Pre-Req

- Linux host or macOS workstation
- Go-built `dist/jclift`
- `jclift.db` writeable
- Optional: `sqlite3`, `jq`

## Steps

1. **Install**

   - Place binary at `/usr/local/bin/jclift` or `./dist/jclift`
   - Copy `configs/jclift.yaml` and adjust paths/rates

2. **Seed Demo & Sanity Check**
   ```bash
   make seed-sample
   make analyze SEVERITY=LOW
   make open-last
   ```
3. **Start API**
   ```bash
   make serve
   ./dist/jclift admin --cmd create-user --username admin --password <pw> --role admin
   make login-jar
   make runs-auth
    ```

4. **Load Rule Pack (optional)**
    ```bash
    make analyze-dsl PACK=configs/rules.example.yaml
    ```

5. **Waivers**

    - Create targeted waivers for known acceptable patterns
    - Re-run analyze and verify Waivers applied in report

6. **Weekly Cadence**
    - Run make analyze on updated JCL snapshots

    - Review Top Offenders and chase owners

    - Track trend via diff between runs


How the Analyzer Detects Inefficiencies
1. Parse JCL Into a Machine-Readable Structure

The analyzer reads JCL lines and converts them into an Intermediate Representation (IR):

{
  "job": "DAILY.LEDGER",
  "step": "S1SORT",
  "program": "SORT",
  "dd": [
    {"ddname": "SYSIN", "content": " SORT FIELDS=(1,10,CH,A)"},
    {"ddname": "SORTWK01", "space": "CYL(1000,100)", "unit": "DISK"},
    {"ddname": "SYSUT1", "dataset": "BANK.CARD.INPUT"}
  ]
}


Now the tool “knows” this step is a SORT, what parameters it uses, how much scratch space (SORTWK) it requests, and what dataset it’s reading.

2. Pattern-Match Against Known Inefficiencies

Think of it like linting code, but for JCL. Rules check the IR for red flags:

Oversized SORTWK

Rule: If SORTWK cylinders allocated >> estimated dataset size, flag as waste.

Evidence: “SORTWK01 CYL(1000,100) allocated, input file size 200MB — over 5x too large.”

Identity SORT (data already sorted, or SORT with no effective key)

Rule: If SYSIN control card shows FIELDS=COPY or sorts on key that matches natural order → flag.

Evidence: “SORT step has no real sorting logic — it’s a no-op consuming 40 MIPS.”

Redundant IEBGENER

Rule: If IEBGENER copies full dataset without filtering, and dataset is immediately read again in next step → flag.

Evidence: “Step S2: IEBGENER duplicates file for S3. Combine into single step.”

Unnecessary Dataset Copies

Rule: Detect sequences like IEBGENER → SORT → IEBGENER, where input/output could be merged.

3. Apply a Cost Model

Once a finding is detected, the analyzer estimates how much CPU it wastes.

SORT cost model: ~O(N log N) operations based on record count.

IEBGENER model: proportional to dataset size.

Calibrated with SMF/RMF logs if available.

Example:

Identity SORT of 100M records = ~50 CPU seconds = ~30 MIPS = ~$75,000/year.

That goes into the report.

4. Explain Findings Like an SME

The analyzer produces SME-style explanations:

Finding: Oversized SORTWK allocation in job DAILY.LEDGER, step S1SORT.
SORTWK01 requests CYL(1000,100), but dataset BANK.CARD.INPUT is only 200MB.
Estimated waste: 12.3 MIPS per run ($15,000 annualized).
Recommendation: Reduce SORTWK to CYL(200,50).


This mirrors what a human SME would write in a review — but the tool does it consistently across thousands of jobs overnight.

5. Continuously Prevent New Waste

Integrated in CI/CD: Any new JCL member with a “bad SORTWK” fails the pipeline.

Waivers exist if intentional, but are tracked (owner + reason).

Over time, the estate gets cleaner automatically, instead of depending on aging SMEs.

✅ Why This Is Powerful

SME brain codified: Rules + cost model = SME knowledge captured.

Scale: An SME might review 10–20 jobs/day. Your analyzer reviews 10,000 jobs in minutes.

Evidence-based: Every finding is backed by JCL snippet + quantified waste → hard for managers to ignore.
