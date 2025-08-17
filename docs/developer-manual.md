# jclift – Developer Manual

_Last updated: 2025-08-17_

Welcome! This manual helps you hack on **jclift**, the JCL Cost/Risk Analyzer.

---

## Quick Start

```bash
# one-time
make bootstrap

# build & analyze the sample
make smoke

# start the API (uses ./jclift.db)
make serve

# open API health
make api-health


## Useful Targets

make build – compile CLI to dist/jclift

make analyze[-low|-med|-high] – run analyzer with thresholds

make analyze-dsl PACK=configs/rules.example.yaml – load YAML DSL rules

make db-summary – show DB counts via sqlite

make docs-serve – browse docs/ at http://localhost:8090

make openapi-serve – Swagger UI at http://localhost:8081

make ci-smoke – LOW→MEDIUM→HIGH sweeps

Coding Standards

Go version: 1.22+

Imports: std → third-party → local

Errors: return wrapped errors (fmt.Errorf("...: %w", err))

Logs: slog with structured fields

Tests: table-driven; golden tests under test/golden

Determinism: sort outputs (rules, findings) for stable diffs

Building Blocks

IR (internal/ir): Source of truth for runs, findings, context.

Parser (internal/parser): Minimal JCL lexer; extracts jobs/steps/DD.

Rules (internal/rules): Go rules; register via init() and rules.Register.

Rules DSL (internal/rulesdsl): Load rule packs from YAML.

Cost Model (internal/cost): Heuristic v1; v2 allows SMF/RMF calibration.

Storage (internal/storage): SQLite schema + CRUD; Postgres later.

Reporting (internal/reporting): JSON/HTML + run diffs.

API (internal/api): Read-only REST + cookie auth, waivers, rules/meta.

Running End to End
# seed demo JCL
make seed-sample

# analyze
make analyze SEVERITY=LOW

# start API
make serve

# create admin
./dist/jclift admin --cmd create-user --username admin --password secret --role admin

# login & browse
make login-jar
make runs-auth

Contributing a New Rule (Go)

Create internal/rules/<short-name>.go.

Add init() with rules.Register(Rule{ ID, Summary, Eval }).

Implement Eval(*ir.Job) []ir.Finding.

Include SavingsMIPS if you can; USD auto-derived if MIPSToUSD>0.

Add doc: docs/rules/<ID>.md.

Add sample in samples/ and run make test-rules.

Releasing

Tag semver; build with -ldflags="-s -w".

Create air-gapped bundle: make pkg-airgap.

Update docs/api/openapi.yaml if endpoints changed.

Future Work

Postgres storage, SSO (OIDC/SAML), schedulers, multi-node workers.

Cost Model v2 with site calibration.

Full Quasar UI.