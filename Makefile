# ---- jclift Makefile (Go-only MVP) -----------------------------------------
SHELL := /bin/sh
.ONESHELL:

# --- Tooling & paths ---------------------------------------------------------
GO        ?= go
BIN       ?= dist/jclift
CFG       ?= ./configs/jclift.yaml
SAMPLES   ?= ./samples/bank-small
REPORTS   ?= ./reports
DB        ?= ./jclift.db
OPEN_CMD  ?= $(shell command -v open 2>/dev/null || command -v xdg-open 2>/dev/null || echo "cat")

# --- Analysis flags (overridable) -------------------------------------------
SEVERITY  ?= LOW            # LOW|MEDIUM|HIGH
DISABLE   ?=                # e.g. "DD-DUPLICATE-DATASET,DD-DISP-OLD-SERIALIZATION"
MIPS_USD  ?=                # overrides config if set (e.g., 250)

ANALYZE_FLAGS := --config $(CFG)
ifneq ($(strip $(SEVERITY)),)
ANALYZE_FLAGS += --severity-threshold $(SEVERITY)
endif
ifneq ($(strip $(DISABLE)),)
ANALYZE_FLAGS += --rules-disable $(DISABLE)
endif
ifneq ($(strip $(MIPS_USD)),)
ANALYZE_FLAGS += --mips-usd $(MIPS_USD)
endif

# --- API/Serve helpers -------------------------------------------------------
API        ?= http://localhost:8080   # Where `serve` listens
LISTEN     ?= :8080                   # CLI listen addr
CORS_ALLOW ?= *                       # CSV origins or '*'
COOKIE     ?= .cookies
USER       ?= admin
PASS       ?= secret

# --- Phonies -----------------------------------------------------------------
.PHONY: help deps bootstrap tidy fmt vet lint build \
        analyze analyze-low analyze-med analyze-high analyze-disable analyze-ci \
        smoke last-id last-two report-last diff-last db-summary open-last \
        seed-sample test-rules ci-smoke test fuzz bench test-golden update-golden golden-diff \
        docker-build docker-run docker-clean ci-local pkg-airgap \
        analyze-dsl serve api-health api-runs api-latest api-findings api-rules \
        login-jar me-auth runs-auth findings-auth create-admin

# --- Help --------------------------------------------------------------------
help: ## Show this help
	@awk 'BEGIN{FS=":.*##"; printf "\nTargets:\n"} /^[a-zA-Z0-9_.-]+:.*##/{printf "  \033[36m%-24s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# --- Deps & hygiene ----------------------------------------------------------
deps: ## Get module deps (SQLite driver, crypto) + tidy
	@$(GO) get modernc.org/sqlite@latest golang.org/x/crypto@latest
	@$(GO) mod tidy

bootstrap: deps ## One-time bootstrap: deps + tidy
	@echo "==> Bootstrapping..."
	@$(GO) mod tidy

tidy: ## go mod tidy
	$(GO) mod tidy

fmt: ## go fmt (write changes)
	@echo "==> go fmt"
	@$(GO) fmt ./...

vet: ## go vet
	@echo "==> go vet"
	@$(GO) vet ./...

lint: fmt vet ## quick lint (fmt + vet)

# --- Build -------------------------------------------------------------------
build: ## Build CLI to dist/
	@echo "==> Building $(BIN)"
	@mkdir -p dist
	$(GO) build -o $(BIN) ./cmd/jclift

# --- Analyze & reports -------------------------------------------------------
analyze: build ## Analyze using config + current flags (SEVERITY, DISABLE, MIPS_USD)
	@echo "==> Analyze ($(SEVERITY))"
	$(BIN) analyze --path $(SAMPLES) --out $(REPORTS) $(ANALYZE_FLAGS)

analyze-low: ## Shortcut: SEVERITY=LOW
	@$(MAKE) analyze SEVERITY=LOW

analyze-med: ## Shortcut: SEVERITY=MEDIUM
	@$(MAKE) analyze SEVERITY=MEDIUM

analyze-high: ## Shortcut: SEVERITY=HIGH
	@$(MAKE) analyze SEVERITY=HIGH

analyze-disable: ## Example: disable two rules (set DISABLE var)
	@$(MAKE) analyze DISABLE="DD-DUPLICATE-DATASET,DD-DISP-OLD-SERIALIZATION"

analyze-ci: build ## CI gate (fails if any findings after gating)
	@echo "==> Analyze with CI gate"
	-$(BIN) analyze --path $(SAMPLES) --out $(REPORTS) $(ANALYZE_FLAGS) --fail-on-findings
	@rc=$$?; if [ "$$rc" -eq 0 ]; then echo "No findings (gate passed)"; else echo "Findings present (gate failed) rc=$$rc"; fi

# --- Convenience: smoke + open report ---------------------------------------
smoke: ## Build + analyze (LOW) + DB summary + open last HTML
	@$(MAKE) build
	@$(MAKE) analyze-low
	@$(MAKE) db-summary
	@$(MAKE) open-last || true

last-id: ## Print the most recent run id from reports/
	@ls -t $(REPORTS)/run-*.json 2>/dev/null | head -1 | sed -E 's|.*/(run-[0-9]+)\.json|\1|'

last-two: ## Print two most recent run ids (base then head)
	@ids=`ls -t $(REPORTS)/run-*.json 2>/dev/null | head -2 | sed -E 's|.*/(run-[0-9]+)\.json|\1|'`; \
	set -- $$ids; if [ $$# -lt 2 ]; then echo "Need at least two runs" >&2; exit 1; fi; \
	echo $$1; echo $$2

report-last: build ## Re-render HTML/JSON for most recent run from DB
	@rid=`$(MAKE) -s last-id`; \
	test -n "$$rid" || { echo "no runs found"; exit 1; }; \
	echo "==> Report $$rid"; \
	$(BIN) report --run $$rid --out $(REPORTS) --config $(CFG)

diff-last: build ## Diff the two most recent runs
	@set -e; \
	read base head <<EOF ; \
$$($(MAKE) -s last-two) \
EOF \
	; echo "==> Diff $$base -> $$head"; \
	$(BIN) diff --base $$base --head $$head --out $(REPORTS) --config $(CFG)

db-summary: ## Show counts from SQLite (requires sqlite3)
	@which sqlite3 >/dev/null 2>&1 || { echo "sqlite3 not found; skipping."; exit 0; }
	@echo "==> DB summary ($(DB))"
	@sqlite3 $(DB) 'select count(*) || " jobs" from jobs;'
	@sqlite3 $(DB) 'select count(*) || " steps" from steps;'
	@sqlite3 $(DB) 'select "findings: " || printf("%-24s %d", rule_id, count(*)) from findings group by rule_id order by count(*) desc;'

open-last: ## Open last HTML report (macOS 'open' or Linux 'xdg-open'; else cat)
	@html=$$(ls -t $(REPORTS)/run-*.html 2>/dev/null | head -1); \
	if [ -n "$$html" ]; then echo "==> Opening $$html"; $(OPEN_CMD) "$$html"; else echo "No reports yet"; fi

# --- Samples -----------------------------------------------------------------
seed-sample: ## Recreate the demo JCL sample (portable; no heredoc)
	@mkdir -p $(SAMPLES)
	@printf '%s\n' \
	'//PAYROLL  JOB (12345),'\''DEMO RUN'\'',CLASS=A,MSGCLASS=X,NOTIFY=&SYSUID' \
	'//S1       EXEC PGM=SORT' \
	'//SYSIN    DD *' \
	'  SORT  FIELDS=COPY' \
	'/*' \
	'//SORTWK01 DD UNIT=SYSDA,SPACE=(CYL,(900,50))' \
	'//SORTWK02 DD UNIT=SYSDA,SPACE=(CYL,(700,50))' \
	'//S2       EXEC PGM=IEBGENER' \
	'//SYSUT1   DD DSN=INPUT.FILE,DISP=SHR' \
	'//SYSUT2   DD DSN=OUTPUT.FILE,DISP=(NEW,CATLG,DELETE)' \
	'//SYSIN    DD DUMMY' \
	'//X1       DD DSN=SHARED.DATA.SET,DISP=OLD' \
	'//X2       DD DSN=SHARED.DATA.SET,DISP=OLD' \
	'//X3       DD DSN=SHARED.DATA.SET,DISP=OLD' \
	> $(SAMPLES)/payroll.jcl
	@echo "Seeded $(SAMPLES)/payroll.jcl"

test-rules: build ## Create a rules sampler job to hit many rules, then analyze
	@mkdir -p $(SAMPLES)
	@printf '%s\n' \
	'//RULES    JOB (12345),'\''RULES'\'',CLASS=A,MSGCLASS=X,NOTIFY=&SYSUID' \
	'//* S0: COND misuse on FIRST step' \
	'//S0       EXEC PGM=IEFBR14,COND=EVEN' \
	'//S0DD     DD   DSN=DUMMY.DATA,DISP=SHR' \
	'//* S1: IDCAMS REPRO identity copy (no selection clauses)' \
	'//S1       EXEC PGM=IDCAMS' \
	'//SYSIN    DD   *' \
	'  REPRO INDATASET(IN.VSAM) OUTDATASET(OUT.VSAM)' \
	'/*' \
	'//* S2: SORT missing SYSIN' \
	'//S2       EXEC PGM=SORT' \
	'//SYSUT1   DD   DSN=IN.FILE,DISP=SHR' \
	'//SYSUT2   DD   DSN=OUT.FILE,DISP=(NEW,CATLG,DELETE)' \
	'//* (no SYSIN on purpose)' \
	'//* S3: Temp dataset leakage (&& kept)' \
	'//S3       EXEC PGM=IEFBR14' \
	'//TMP1     DD   DSN=&&TMP1,DISP=(NEW,CATLG)' \
	'//* S4: DISP=MOD append' \
	'//S4       EXEC PGM=IEFBR14' \
	'//MOD1     DD   DSN=SOME.DATA,DISP=MOD' \
	'//* S5: GDG roll-off (read -1, write 0)' \
	'//S5       EXEC PGM=IEFBR14' \
	'//IN1      DD   DSN=BASE.FILE.GDG(-1),DISP=SHR' \
	'//OUT1     DD   DSN=BASE.FILE.GDG(0),DISP=(NEW,CATLG)' \
	> $(SAMPLES)/rules-sampler.jcl
	@echo "Seeded $(SAMPLES)/rules-sampler.jcl"
	@$(BIN) analyze --path $(SAMPLES) --out $(REPORTS) $(ANALYZE_FLAGS)
	@$(MAKE) db-summary
	@$(MAKE) open-last || true

# --- CI-style smoke across thresholds ---------------------------------------
ci-smoke: build ## Run LOW→MEDIUM→HIGH passes and print rule counts for each
	@for sev in LOW MEDIUM HIGH; do \
	  echo ""; echo "==> CI Smoke ($$sev)"; \
	  $(BIN) analyze --path $(SAMPLES) --out $(REPORTS) --config $(CFG) --severity-threshold $$sev >/dev/null 2>&1; \
	  rc=$$?; \
	  if [ "$$rc" -ne 0 ]; then \
	    echo "analyze exited with rc=$$rc (continuing)"; \
	  fi; \
	  rid=$$(ls -t $(REPORTS)/run-*.json 2>/dev/null | head -1 | sed -E 's|.*/(run-[0-9]+)\.json|\1|'); \
	  if [ -z "$$rid" ]; then \
	    echo "no run JSON found; skipping sqlite summary"; \
	    continue; \
	  fi; \
	  if command -v sqlite3 >/dev/null 2>&1; then \
	    sqlite3 $(DB) "select printf('%-26s %d', rule_id, count(*)) from findings where run_id='$$rid' group by rule_id order by count(*) desc;"; \
	  else \
	    echo "(sqlite3 not found; skipping DB summary)"; \
	  fi; \
	done

# --- go test suite & benchmarks ---------------------------------------------
test: ## Run unit tests (including golden-ish checks)
	@echo "==> go test"
	@$(GO) test ./test/... -count=1

fuzz: ## Run parser fuzz for a short time (requires Go 1.18+)
	@echo "==> go fuzz (parser)"
	@$(GO) test ./test/fuzz -run=Fuzz -fuzz=Fuzz -fuzztime=5s

bench: ## Run micro-benchmarks
	@echo "==> go bench"
	@$(GO) test ./test/perf -bench=Analyze -benchtime=2s -run=^$

test-golden: ## Validate golden snapshot
	@$(GO) test ./test/golden -run TestGolden_PayrollSnapshot -count=1

update-golden: ## Re-write golden snapshot from current analyzer
	@$(GO) test ./test/golden -run TestGolden_PayrollSnapshot -count=1 -args -update

golden-diff: ## Show diff between current output and golden
	@$(GO) test ./test/golden -run TestGolden_PayrollSnapshot -count=1 || true

# --- Docker helpers ----------------------------------------------------------
docker-build: ## Build Docker image jclift/core:dev
	@docker build -f packaging/docker/Dockerfile.core -t jclift/core:dev .

docker-run: ## Run analyzer in container using local samples/configs/reports
	@docker run --rm \
	  -v $(PWD)/samples:/app/samples:ro \
	  -v $(PWD)/reports:/app/reports \
	  -v $(PWD)/configs:/app/configs:ro \
	  -v $(PWD)/jclift.db:/app/data/jclift.db \
	  jclift/core:dev \
	  analyze --path /app/samples --out /app/reports --config /app/configs/jclift.yaml

docker-clean: ## Remove local image
	-@docker rmi jclift/core:dev 2>/dev/null || true

ci-local: ## Quick local CI: tidy, vet, build, tests, fuzz(5s)
	@echo "==> Local CI"
	@$(GO) mod tidy
	@$(GO) vet ./...
	@mkdir -p dist
	@$(GO) build -trimpath -ldflags="-s -w" -o dist/jclift ./cmd/jclift
	@$(GO) test ./... -count=1
	@$(GO) test ./test/fuzz -run=Fuzz -fuzz=Fuzz -fuzztime=5s

pkg-airgap: build ## Create a tarball with binary + configs + sample
	@mkdir -p dist/pkg
	@cp -a dist/jclift dist/pkg/
	@mkdir -p dist/pkg/configs dist/pkg/samples
	@cp -a configs/jclift.yaml dist/pkg/configs/ 2>/dev/null || true
	@cp -a samples/* dist/pkg/samples/ 2>/dev/null || true
	@tar -C dist -czf dist/jclift-airgap.tgz pkg
	@echo "Wrote dist/jclift-airgap.tgz"

# --- DSL rule packs ----------------------------------------------------------
analyze-dsl: build ## Analyze using a DSL rules pack: make analyze-dsl PACK=configs/rules.example.yaml
	@test -n "$(PACK)" || { echo "Usage: make analyze-dsl PACK=path/to/rules.yaml"; exit 2; }
	@$(BIN) analyze --path $(SAMPLES) --out $(REPORTS) --config $(CFG) --rules-pack $(PACK)

# --- API server & endpoints --------------------------------------------------
serve: build ## Run REST API server (Ctrl+C to stop)
	@./dist/jclift serve --db $(DB) --listen $(LISTEN) --cors-allow "$(CORS_ALLOW)"

api-health: ## Hit /health (server must be running)
	@curl -s "$(API)/api/v1/health" | jq .

api-runs: ## List recent runs (limit=5)
	@curl -s "$(API)/api/v1/runs?limit=5" | jq .

api-latest: ## Fetch the most recent run payload
	@curl -s "$(API)/api/v1/runs/latest" | jq .

api-findings: ## Findings for a run: make api-findings RUN=<run-id> [MIN=MEDIUM]
	@test -n "$(RUN)" || { echo "Usage: make api-findings RUN=<run-id> [MIN=MEDIUM]"; exit 2; }
	@curl -s "$(API)/api/v1/runs/$(RUN)/findings?min_severity=$(if $(MIN),$(MIN),LOW)" | jq .

api-rules: ## List registered rules (IDs + summaries)
	@curl -s "$(API)/api/v1/rules" | jq .

# --- Auth helpers (cookie jar) ----------------------------------------------
create-admin: build ## Create local user: make create-admin USER=admin PASS=secret [ROLE=admin|viewer]
	@test -n "$(USER)" && test -n "$(PASS)" || { echo "Usage: make create-admin USER=... PASS=... [ROLE=admin|viewer]"; exit 2; }
	@./dist/jclift admin --cmd create-user \
	  --username "$(USER)" --password "$(PASS)" --role "$(if $(ROLE),$(ROLE),admin)" --db $(DB)

login-jar: ## Login and save cookie jar to $(COOKIE)
	@curl -s -c $(COOKIE) -X POST '$(API)/api/v1/auth/login' \
	  -H 'Content-Type: application/json' \
	  -d '{"username":"$(USER)","password":"$(PASS)"}' | jq .

me-auth: ## Who am I (requires cookie jar)
	@curl -s -b $(COOKIE) '$(API)/api/v1/me' | jq .

runs-auth: ## Authenticated /runs (requires cookie jar)
	@curl -s -b $(COOKIE) '$(API)/api/v1/runs?limit=5' | jq .

findings-auth: ## Authenticated findings: make findings-auth RUN=<run-id> [MIN=LOW|MEDIUM|HIGH]
	@test -n "$(RUN)" || { echo "Usage: make findings-auth RUN=<run-id> [MIN=LOW|MEDIUM|HIGH]"; exit 2; }
	@curl -s -b $(COOKIE) '$(API)/api/v1/runs/$(RUN)/findings?min_severity=$(if $(MIN),$(MIN),LOW)' | jq .

# --- Cleanup -----------------------------------------------------------------
clean: ## Remove build artifacts (keep DB/reports)
	@rm -rf dist

realclean: ## Remove build, DB, and reports
	@rm -rf dist $(DB) $(REPORTS)
# ---------------------------------------------------------------------------

# --- Waiver helpers (authenticated; use cookie jar) --------------------------
.PHONY: waivers-list waiver-create waiver-revoke
waivers-list: ## List waivers (make waivers-list [ACTIVE=1])
	@curl -s -b $(COOKIE) '$(API)/api/v1/waivers?active=$(if $(ACTIVE),$(ACTIVE),1)' | jq .

waiver-create: ## Create waiver: RULE=<id> [JOB=payroll STEP=S1 PAT=substring REASON=... UNTIL=2025-12-31T00:00:00Z]
	@test -n "$(RULE)" && test -n "$(REASON)" && test -n "$(UNTIL)" || { echo "Usage: make waiver-create RULE=<rule_id> REASON=... UNTIL=<RFC3339> [JOB=.. STEP=.. PAT=..]"; exit 2; }
	@jq -n --arg rule "$(RULE)" --arg job "$(JOB)" --arg step "$(STEP)" --arg pat "$(PAT)" --arg reason "$(REASON)" --arg until "$(UNTIL)" \
	  '{rule_id:$$rule, job:$$job, step:$$step, pattern_sub:$$pat, reason:$$reason, expires_at:$$until}' \
	| curl -s -b $(COOKIE) -H 'Content-Type: application/json' -d @- '$(API)/api/v1/waivers' | jq .

waiver-revoke: ## Revoke waiver: make waiver-revoke ID=1
	@test -n "$(ID)" || { echo "Usage: make waiver-revoke ID=<waiver_id>"; exit 2; }
	@curl -s -b $(COOKIE) -X POST '$(API)/api/v1/waivers/$(ID)/revoke' | jq .

.PHONY: api-rules-meta
api-rules-meta: ## Extended rule metadata
	@curl -s "$(API)/api/v1/rules/meta" | jq .

