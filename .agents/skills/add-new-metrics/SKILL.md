---
name: add-new-metrics
description: Step-by-step guide for adding new Prometheus metrics or entire metric collection modules to the klipper exporter.
---

# Adding New Metrics

This skill covers every file that must be created or modified when adding new
metrics or a new collection module to the Prometheus Klipper Exporter.

**Always run verification after making changes:**
```sh
make build && make test && (cd docs && npm run build)
```

---

## Step 1 — Collector Implementation

### 1A. New collector file (`collector/<name>.go`)

Pattern for a simple module:

```go
package collector

import (
    "github.com/prometheus/client_golang/prometheus"
    log "github.com/sirupsen/logrus"
)

type MoonrakerYourModuleResponse struct {
    Result struct {
        SomeField string `json:"some_field"`
    } `json:"result"`
}

func (c Collector) collectYourModule(ch chan<- prometheus.Metric) {
    log.Infof("Collecting your_module for %s", c.target)

    var result MoonrakerYourModuleResponse
    if err := c.fetchFromMoonraker("/api/path", &result); err != nil {
        log.Error(err)
        return
    }

    // Unlabeled scalar gauge
    c.emitGauge(ch, "klipper_your_metric", "Description.", float64Value)

    // Unlabeled scalar counter
    c.emitCounter(ch, "klipper_your_counter_total", "Description.", float64Value)

    // Info-style gauge (value=1, state in label) — only emits if non-empty
    emitStateInfoMetric(ch, "klipper_your_state_info", "Description.", "state", result.Result.SomeField)

    // Labeled metric with dynamic instances
    labels := []string{"sensor"}
    desc := prometheus.NewDesc("klipper_your_labeled_metric", "Description", labels, nil)
    for key, value := range someMap {
        ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, value, GetValidLabelName(key))
    }

    // Bool → 0/1
    c.emitGauge(ch, "klipper_your_bool", "Description.", boolToFloat64(someBool))
}
```

**Key patterns:**
- Response structs: `Moonraker<ModuleName>Response` with anonymous `Result struct{}`
- Exception: `printer_object.go` uses custom `UnmarshalJSON` and `mapstructure` for dynamic keys
- Exception: `mmu.go` uses `MMUResponse` / `MMUStatus` naming
- Functions that make distinct API calls: `fetch<MoonrakerObject>()` helper methods

### 1B. Adding metrics to an existing module

If adding to a file that already exists (e.g. `process_stats.go`, `printer_object.go`):

1. Add new fields to the response struct if a new API field is needed
2. Add metric emission inside the existing `collect*()` method

---

## Step 2 — Register in `Collect()`

**File:** `collector/collector.go`

Add a new `slices.Contains` guard in the `Collect()` method:

```go
// Your Module
if slices.Contains(c.modules, "your_module") {
    c.collectYourModule(ch)
}
```

- Maintain alphabetical-ish grouping (existing modules are grouped logically)
- If the module shares an endpoint with another (like `process_stats`/`network_stats` share `/machine/proc_stats`), gate them together
- If deprecating a module, keep the old name with a `log.Errorf` warning

---

## Step 3 — Default Modules

**File:** `main.go`

If the module should be enabled by default (recommended for lightweight modules
that provide universally useful metrics), add it to the default list:

```go
modules := []string{"server_info", "process_stats", "job_queue", "system_info"}
```

Modules that are opt-in (heavy, niche, or have side effects): leave out of the
default list.

---

## Step 4 — Documentation

### 4A. Create metric reference page

**File:** `docs/metrics/<name>.md`

Template:

```markdown
# Your Module Name

**Module:** `<module_name>` (default|optional)
**API Endpoint:** [`/api/path`](https://moonraker.readthedocs.io/en/latest/web_api/#anchor)

Brief description.

## Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `klipper_your_metric` | Gauge | Description |
| `klipper_your_state_info` | Gauge=1 | Description with `label` label |

## Example PromQL

```promql
# Comment
klipper_your_metric
```

**Type conventions:**
- `Gauge` — unlabeled scalar gauge
- `Counter` — unlabeled scalar counter
- `Gauge=1` — info-style gauge with label (emitted via `emitStateInfoMetric`)
- Include labels column when metrics have labels

### 4B. Update sidebar

**File:** `docs/.vitepress/config.js`

Add a new sidebar entry in the alphabetical position:

```js
{ text: 'Your Module', link: '/metrics/your-module' },
```

### 4C. Update summary index

**File:** `docs/metrics/index.md`

1. **Module overview table** — add row with module name, default marker, API endpoint link, and total metric count
2. **Default modules sentence** — update if adding a new default module
3. **All Metrics by Module** — add a `### your_module` subsection with a compact metric table and `[Full reference →](./your-module)` link
4. If adding metrics to an existing module, update the metric count in the overview table and add rows to the existing subsection table
5. For large modules (MMU, printer_objects), use category tables instead of individual metric rows

---

## Step 5 — Grafana Dashboards

**Files:** `test/grafana/provisioning/dashboards/*.json`

### 5A. When to update dashboards

- **New metric added to an existing module** — add a panel to the corresponding dashboard if the metric is useful for visualisation
- **New module** — add a new dashboard JSON file if the module provides enough meaningful data for a dedicated view, OR add panels to an existing dashboard

### 5B. Panel patterns

**Stat panel (numeric value):**
```json
{
  "datasource": { "type": "prometheus", "uid": "prometheus" },
  "fieldConfig": {
    "defaults": {
      "color": { "mode": "thresholds" },
      "mappings": [
        {
          "options": { "0": { "text": "Disconnected" }, "1": { "text": "Connected" } },
          "type": "value"
        }
      ],
      "thresholds": {
        "mode": "absolute",
        "steps": [
          { "color": "red", "value": null },
          { "color": "green", "value": 1 }
        ]
      }
    },
    "overrides": []
  },
  "options": {
    "colorMode": "value",
    "graphMode": "none",
    "reduceOptions": { "calcs": ["lastNotNull"], "fields": "", "values": false },
    "textMode": "auto"
  },
  "targets": [{
    "datasource": { "type": "prometheus", "uid": "prometheus" },
    "expr": "klipper_your_metric{job=\"$job\", instance=\"$instance\"}",
    "instant": true,
    "legendFormat": "__auto"
  }],
  "title": "Your Title",
  "type": "stat"
}
```

**Stat panel (info label display — Gauge=1 with label):**
```json
{
  "options": {
    "colorMode": "value",
    "graphMode": "none",
    "reduceOptions": { "calcs": ["lastNotNull"], "fields": "", "values": false },
    "textMode": "name"
  },
  "targets": [{
    "expr": "klipper_your_state_info{job=\"$job\", instance=\"$instance\"}",
    "instant": true,
    "legendFormat": "{{state}}"
  }]
}
```

**Stat panel (count of labeled instances):**
```json
{
  "targets": [{
    "expr": "count(klipper_component_info{job=\"$job\", instance=\"$instance\"})",
    "instant": true
  }]
}
```

**Timeseries panel:**
Use `"type": "timeseries"` and include `custom` field config with axis/line settings.

### 5C. Template variables

Every dashboard should define `job` and `instance` template variables:

```json
"templating": {
  "list": [
    {
      "name": "job",
      "query": { "query": "label_values(job)", "refId": "StandardVariableQuery" },
      "type": "query",
      "hide": 0
    },
    {
      "name": "instance",
      "query": { "query": "label_values(up{job=\"$job\"}, instance)", "refId": "StandardVariableQuery" },
      "type": "query",
      "hide": 0
    }
  ]
}
```

Use `up{job="$job"}` for the instance query so it works even before module-specific
metrics exist.

### 5D. Provisioning

- Dashboard JSON goes in `test/grafana/provisioning/dashboards/`
- The YAML provisioner at `test/grafana/provisioning/dashboards/klipper.yml` auto-discovers all `.json` files in the same directory — no registration needed
- Update the dashboard table in `docs/developers/index.md` if adding a new dashboard
- Update the dashboard table in `test/README.md` if one exists

---

## Step 6 — Tests

### 6A. When to add tests

- Unit test: label sanitization, helper functions
- Integration test: verify metrics are emitted with correct names from a fixture.

### 6B. Integration test pattern

**File:** `tests/<module>_test.go`

```go
package test

import (
    "encoding/json"
    "net/http/httptest"
    "testing"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/scross01/prometheus-klipper-exporter/collector"
)

func TestYourModuleMetrics(t *testing.T) {
    // Load fixture or create response dynamically
    fixture := `{ "result": { ... } }`

    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte(fixture))
    }))
    defer server.Close()

    c := collector.New(r.Context(), server.Listener.Addr().String(), []string{"your_module"}, "")

    ch := make(chan prometheus.Metric, 100)
    go func() {
        c.Collect(ch)
        close(ch)
    }()

    var collected []string
    for m := range ch {
        collected = append(collected, m.Desc().String())
    }

    // Check expected metric names are present
    // Use t.Run subtests for clarity
}
```

### 6C. Test fixtures

- Store large JSON fixtures in `tests/<name>_response.json`
- Load with `os.ReadFile` + `json.Unmarshal`
- For dynamic responses, construct maps and `json.Marshal` in the test

### 6D. Run tests

```sh
make test    # runs go test ./tests/...
```

---

## Step 7 — Example and Deployment Configs

### 7A. Prometheus scrape config

**Files:**
- `example/prometheus.yml` — production example
- `test/prometheus.yml` — dev/test environment

If adding a new module, add it to the `params.modules` list in both files if it
should be scraped by default in those environments:

```yaml
params:
  modules:
    - your_module
    - process_stats
    - ...
```

### 7B. Virtual printer addon config

**File:** `test/printer_data/config/addons/`

If the module exercises a Klipper printer feature that isn't covered by existing
addon configs:

1. Check pin conflicts in the existing addon table (in `docs/developers/index.md`)
2. Create a new `.cfg` file in `test/printer_data/config/addons/`
3. Add `[include addons/your_file.cfg]` to `test/printer_data/config/printer.cfg`
4. Update the addon table in `docs/developers/index.md`
5. Update the addon table in `test/README.md`

---

## Step 8 — Changelog

**File:** `CHANGELOG.md`

Add entries under the current `vX.Y.Z` heading using present tense bullet points:

```markdown
- Add `klipper_your_metric` and `klipper_your_other_metric` metrics to `your_module` module
- Add `your_module` module with Apollo integration
- Add corresponding panels to Grafana dashboards (dashboard-name)
- Add support for `<feature>` with `<addon.cfg>` config section in virtual test environment
```

---

## Full Checklist

Use this checklist to ensure nothing is missed:

- [ ] **Implementation:** collector file created/modified with response structs and metric emission
- [ ] **Registration:** `slices.Contains` guard added in `collector.go`'s `Collect()`
- [ ] **Default modules:** added to `main.go` if appropriate
- [ ] **Docs page:** `docs/metrics/<name>.md` created with metric table and PromQL examples
- [ ] **Sidebar:** entry added in `docs/.vitepress/config.js`
- [ ] **Summary index:** `docs/metrics/index.md` — overview table, default sentence, module subsection
- [ ] **Grafana dashboard:** panels or whole dashboard added in `test/grafana/provisioning/dashboards/`
- [ ] **Dev docs:** dashboard table updated in `docs/developers/index.md`
- [ ] **Tests:** `tests/<module>_test.go` with fixture data and metric verification
- [ ] **Example configs:** `example/prometheus.yml` and/or `test/prometheus.yml` updated
- [ ] **Virtual printer:** addon config added if exercising a new printer feature
- [ ] **Changelog:** entries added under current version
- [ ] **Format & verify:** `make fmt && make build && make test && (cd docs && npm run build)`
- [ ] **AGENTS.md update:** add any new gotchas or patterns discovered during implementation
