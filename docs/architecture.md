---

### `docs/architecture.md`

```markdown
# jclift – Architecture

_Last updated: 2025-08-17_

## Overview

**Goal**: detect cost/risk issues in mainframe JCL before runtime.

### Core Components

- **CLI / API Server (Go)**
  - `analyze`: parse → annotate cost → run rules → persist → report
  - `serve`: REST API, cookie auth, waivers
- **IR**: normalized representation of jobs
- **Rules**: Go and DSL-based
- **Cost Model v1**: heuristic
- **Storage**: SQLite (MVP), Postgres later
- **Reporting**: JSON/HTML; diff runs
- **UI**: Quasar (future)

## C4 (Text)

**C1 – Context**  
Mainframe batch owners use jclift to reduce MIPS waste and operational risk.

**C2 – Containers**

- CLI/API (Go)
- SQLite DB
- (Future) Worker pool
- Quasar UI

**C3 – Components (CLI/API)**

- Parser
- Rules Engine
- Cost Annotator
- Storage Adapter
- HTTP Handlers
- Waivers & Auth

**C4 – Code**  
See `internal/*` packages.

## Data Flow

1. **Input**: Directory of JCL files (`--path`)
2. **Parse** → `ir.Run{ Jobs, Steps, DD }`
3. **Annotate costs** → `Step.Annotations.Cost`
4. **Rules** → `[]ir.Finding` + waivers applied
5. **Persist** → SQLite
6. **Report** → JSON/HTML
7. **Serve** → REST querying persisted runs

## Cost Model v1

- SORT: `MIPS ≈ α + β × sizeGB`
- IEBGENER/REPRO (identity): small baseline
- Identity SORT: constant baseline
- Size estimation from `SPACE=(CYL,...)` + geometry

## Security & Compliance

- Local users (username/password), hashed (Argon2id/bcrypt depending on build)
- HttpOnly cookie `jclift_session`
- Minimal PII; opt-in telemetry later
- Air-gapped install supported (Docker Compose)

## Extensibility

- DSL rules loader for rapid SME iteration
- Plugin adapters (future) via gRPC
- Connectors (Git/SFTP/NFS) for large estates

## Availability & Scale

- MVP: single binary, single DB
- v1.0: Postgres, DAG planner
- v2.0: multi-node workers + queue + schedulers
```
