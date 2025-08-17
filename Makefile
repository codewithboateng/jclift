# ---- jclift Makefile (Go-only MVP) -----------------------------------------
SHELL := /bin/sh
.ONESHELL:

GO        ?= go
BIN       ?= dist/jclift
CFG       ?= ./configs/jclift.yaml
SAMPLES   ?= ./samples/bank-small
REPORTS   ?= ./reports
DB        ?= ./jclift.db
OPEN_CMD  ?= $(shell command -v open 2>/dev/null || command -v xdg-open 2>/dev/null || echo "cat")

# Variables you can override per-invocation:
SEVERITY  ?= LOW            # LOW|MEDIUM|HIGH
DISABLE   ?=                # e.g. "DD-DUPLICATE-DATASET,DD-DISP-OLD-SERIALIZATION"
MIPS_USD  ?=                # overrides config if set (e.g., 250)

# Derived flags
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

.PHONY: help bootstrap tidy fmt vet lint build analyze analyze-low analyze-med analyze-high \
        analyze-disable analyze-ci smoke last-id last-two report-last diff-last \
        db-summary db-top clean realclean open-last seed-sample

help: ## Show this help
	@awk 'BEGIN{FS=":.*##"; printf "\nTargets:\n"} /^[a-zA-Z0-9_-]+:.*##/{printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

bootstrap: ## One-time: tidy deps
	@echo "==> Bootstrapping..."
	$(GO) mod tidy

tidy: ## go mod tidy
	$(GO) mod tidy

fmt: ## go fmt (write changes)
	@echo "==> go fmt"
	@$(GO) fmt ./...

vet: ## go vet
	@echo "==> go vet"
	@$(GO) vet ./...

lint: fmt vet ## quick lint (fmt + vet)

build: ## Build CLI to dist/
	@echo "==> Building $(BIN)"
	@mkdir -p dist
	$(GO) build -o $(BIN) ./cmd/jclift

analyze: build ## Analyze using config + current flags (SEVERITY, DISABLE, MIPS_USD)
	@echo "==> Analyze ($(SEVERITY))"
	$(BIN) analyze --path $(SAMPLES) --out $(REPORTS) $(ANALYZE_FLAGS)

analyze-low: ## Shortcut: SEVERITY=LOW
	@$(MAKE) analyze SEVERITY=LOW

analyze-med: ## Shortcut: SEVERITY=MEDIUM
	@$(MAKE) analyze SEVERITY=MEDIUM

analyze-high: ## Shortcut: SEVERITY=HIGH
	@$(MAKE) analyze SEVERITY=HIGH

analyze-disable: ## Example: disable two rules
	@$(MAKE) analyze DISABLE="DD-DUPLICATE-DATASET,DD-DISP-OLD-SERIALIZATION"

analyze-ci: build ## CI gate (fails if any findings after gating)
	@echo "==> Analyze with CI gate"
	-$(BIN) analyze --path $(SAMPLES) --out $(REPORTS) $(ANALYZE_FLAGS) --fail-on-findings
	@rc=$$?; if [ "$$rc" -eq 0 ]; then echo "No findings (gate passed)"; else echo "Findings present (gate failed) rc=$$rc"; fi

smoke: ## Build + analyze (LOW) + DB summary + open report
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

seed-sample: ## Recreate the demo JCL sample (robust, no heredoc)
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



clean: ## Remove build artifacts (keep DB/reports)
	@rm -rf dist

realclean: ## Remove build, DB, and reports
	@rm -rf dist $(DB) $(REPORTS)
# ---------------------------------------------------------------------------

.PHONY: test-rules
test-rules: build ## Create a rules sampler job that triggers remaining rules, then analyze
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

.PHONY: ci-smoke
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

.PHONY: test fuzz bench
test: ## Run unit tests (including golden-ish checks)
	@echo "==> go test"
	@go test ./test/... -count=1

fuzz: ## Run parser fuzz for a short time (requires Go 1.18+)
	@echo "==> go fuzz (parser)"
	@go test ./test/fuzz -run=Fuzz -fuzz=Fuzz -fuzztime=5s

bench: ## Run micro-benchmarks
	@echo "==> go bench"
	@go test ./test/perf -bench=Analyze -benchtime=2s -run=^$


.PHONY: test-golden update-golden golden-diff
test-golden: ## Validate golden snapshot
	@go test ./test/golden -run TestGolden_PayrollSnapshot -count=1

update-golden: ## Re-write golden snapshot from current analyzer
	@go test ./test/golden -run TestGolden_PayrollSnapshot -count=1 -args -update

golden-diff: ## Show diff between current output and golden (requires update-golden first)
	@go test ./test/golden -run TestGolden_PayrollSnapshot -count=1 || true


.PHONY: docker-build docker-run docker-clean ci-local pkg-airgap

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
	@go mod tidy
	@go vet ./...
	@mkdir -p dist
	@go build -trimpath -ldflags="-s -w" -o dist/jclift ./cmd/jclift
	@go test ./... -count=1
	@go test ./test/fuzz -run=Fuzz -fuzz=Fuzz -fuzztime=5s

pkg-airgap: build ## Create a tarball with binary + configs + sample
	@mkdir -p dist/pkg
	@cp -a dist/jclift dist/pkg/
	@mkdir -p dist/pkg/configs dist/pkg/samples
	@cp -a configs/jclift.yaml dist/pkg/configs/ 2>/dev/null || true
	@cp -a samples/* dist/pkg/samples/ 2>/dev/null || true
	@tar -C dist -czf dist/jclift-airgap.tgz pkg
	@echo "Wrote dist/jclift-airgap.tgz"

.PHONY: analyze-dsl
analyze-dsl: build ## Analyze using a DSL rules pack: make analyze-dsl PACK=configs/rules.example.yaml
	@test -n "$(PACK)" || { echo "Usage: make analyze-dsl PACK=path/to/rules.yaml"; exit 2; }
	$(BIN) analyze --path $(SAMPLES) --out $(REPORTS) --config $(CFG) --rules-pack $(PACK)

.PHONY: serve
serve: build ## Run REST API server (default :8080)
	@./dist/jclift serve --db $(DB)

