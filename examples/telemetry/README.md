# Simba Telemetry Example (OpenTelemetry Integration)

**This example demonstrates how to wire full observability (distributed tracing, custom metrics, automatic HTTP instrumentation) into your Simba application via an explicit, provider-based OpenTelemetry integration.**

Simba itself is telemetry-agnostic. All tracing and metrics are enabled via injection of a provider object—no static helpers, no built-in instrumentation, no global Simba telemetry state.

---

## 🚀 What This Example Shows

- **OpenTelemetry provider injection pattern** (the only supported approach)
- **Distributed tracing** for HTTP endpoints (custom spans, attributes, nested operations)
- **Custom application metrics** via OTel
- **Error tracking in traces**
- **Full observability stack using Docker Compose:**
  - Simba app (instrumented)
  - OpenTelemetry Collector
  - Jaeger (for traces)
  - Prometheus (for metrics)
  - Grafana (for metrics dashboards)

---

## 📝 Quick Start

**Requirements:**
- [Go](https://golang.org/doc/install) (for running locally)
- [Docker](https://docs.docker.com/get-docker/) and [Docker Compose](https://docs.docker.com/compose/)

### 1. Clone the repo (or use this directory in your Simba monorepo)

```bash
cd examples/telemetry
```

### 2. Launch Everything with Docker Compose

This will start the Simba demo app, OTel collector, Jaeger, Prometheus, and Grafana all at once.

```bash
docker-compose up --build
```

- Wait ~30 seconds for the stack to settle.

### 3. Call Endpoints to Generate Telemetry

- Create a user (triggers several spans):
  ```bash
  curl -X POST http://localhost:9999/users \
    -H "Content-Type: application/json" \
    -d '{"name":"Jane Doe","email":"jane@example.com"}'
  ```
- Get a user (demonstrates nested spans and error propagation):
  ```bash
  curl http://localhost:9999/users/1
  curl http://localhost:9999/users/999  # triggers 404 + error in span
  ```
- Trigger custom metrics:
  ```bash
  curl http://localhost:9999/metrics-demo
  ```

### 4. Explore Observability UIs
- **Jaeger (Traces):** [http://localhost:16686](http://localhost:16686) — Search for service `simba-telemetry-demo`
- **Grafana (Metrics dashboards):** [http://localhost:3000](http://localhost:3000) — Login: admin/admin
- **Prometheus (Raw metrics, queries):** [http://localhost:9090](http://localhost:9090)
- **App itself:** [http://localhost:9999](http://localhost:9999)


---

## 🔍 Example Code Highlights

- **Provider pattern:**
  See [`main.go`](./main.go):
  ```go
  import (
      "github.com/sillen102/simba"
      "github.com/sillen102/simba/telemetry"
      "github.com/sillen102/simba/telemetry/config"
  )

  func main() {
      tcfg := &config.TelemetryConfig{...}
      app := simba.Default()
      prov, err := telemetry.NewOtelTelemetryProvider(context.Background(), tcfg)
      if err != nil { panic(err) }
      app.SetTelemetryProvider(prov) // <-- all instrumentation starts here

      // Create custom meters/embed rich tracing from here...
      app.Start()
  }
  ```
- **No static Simba helpers anywhere.**
- **All HTTP, tracing, and metric logic is opt-in and locally configured.**

---

## 🧩 Endpoints and Demo Behavior

- `POST /users` — Creates a user, triggers spans for validation, DB op, and email simulation
- `GET /users/{id}` — Fetches a user, with spans for DB read and an external API call (failures are traced)
- `GET /metrics-demo` — Increments demo metrics and records histogram
- `GET /health` — Built-in Simba health check (auto-instrumented)

---

## 🆕 How Provider Injection Works (Migration)

- **There are no static helpers like `simba.telemetry.*` or `InitTracer` or `NewMetrics`.**
- Instrumentation is only enabled if you inject a provider object (see above pattern).
- To instrument your own code, use the injected provider to get tracers/meters:
  ```go
  otelProv := prov.(*telemetry.OtelTelemetryProvider)
  tracer := otelProv.Provider().Tracer("my-service")
  meter := otelProv.Provider().Meter("my-service")
  ```

---

## ⚠️ Troubleshooting

- If you see "provider does not implement Shutdown", update Simba and make sure all provider types have a Shutdown method.
- Metrics and traces won’t show up if you don’t inject a provider, or if your collector/stack isn’t running.
- For detailed logs, see `docker-compose logs` and the app logs.
- Stack slow to respond? Wait for all containers to start, or check resource usage.

---

## 🔮 Next Steps / Customization

- Modify endpoints in `main.go` to add your own spans, metrics, or attributes
- Use a different OTel exporter for local development (e.g. stdout)
- Tune config (`config.yaml`) or override with env vars
- Explore dashboards in Grafana or deep traces in Jaeger

---

## 📖 Further Reading

- [Simba Project Docs](https://github.com/sillen102/simba)
- [OpenTelemetry (concepts, exporters, advanced)](https://opentelemetry.io/docs/)

---

**Legacy Note:** This example is up-to-date with Simba ≥vNEXT (provider pattern only). If coming from older Simba, migrate to explicit provider injection—see above. No static telemetry helpers remain.
