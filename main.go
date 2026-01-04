package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// Build-time variables (set via ldflags)
	Version   = "dev"
	Commit    = "unknown"
	BuildTime = "unknown"
)

// Prometheus metrics
var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "epochcloud_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "epochcloud_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	appInfo = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "epochcloud_app_info",
			Help: "Application build information",
		},
		[]string{"version", "commit", "build_time", "environment"},
	)
)

type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

type VersionResponse struct {
	Version     string `json:"version"`
	Commit      string `json:"commit"`
	BuildTime   string `json:"buildTime"`
	Hostname    string `json:"hostname"`
	Environment string `json:"environment"`
}

type PageData struct {
	Version     string
	Commit      string
	BuildTime   string
	Hostname    string
	Environment string
	Timestamp   string
}

const pageTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>EpochCloud Test | {{.Environment}}</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
            background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
            min-height: 100vh;
            color: #e8e8e8;
            display: flex;
            flex-direction: column;
            align-items: center;
            justify-content: center;
            padding: 2rem;
        }
        .container {
            max-width: 800px;
            width: 100%;
        }
        .header {
            text-align: center;
            margin-bottom: 3rem;
        }
        .header h1 {
            font-size: 3rem;
            font-weight: 700;
            margin-bottom: 0.5rem;
            background: linear-gradient(90deg, #00d9ff, #00ff88);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
        }
        .env-badge {
            display: inline-block;
            padding: 0.5rem 1.5rem;
            border-radius: 50px;
            font-weight: 600;
            text-transform: uppercase;
            letter-spacing: 0.1em;
            margin-top: 1rem;
        }
        .env-dev { background: #ff6b6b; color: #fff; }
        .env-staging { background: #feca57; color: #1a1a2e; }
        .env-prod { background: #00d26a; color: #fff; }
        .card {
            background: rgba(255, 255, 255, 0.05);
            backdrop-filter: blur(10px);
            border: 1px solid rgba(255, 255, 255, 0.1);
            border-radius: 16px;
            padding: 2rem;
            margin-bottom: 2rem;
        }
        .card h2 {
            font-size: 1.25rem;
            color: #00d9ff;
            margin-bottom: 1rem;
            display: flex;
            align-items: center;
            gap: 0.5rem;
        }
        .info-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 1rem;
        }
        .info-item {
            background: rgba(255, 255, 255, 0.03);
            padding: 1rem;
            border-radius: 8px;
        }
        .info-item label {
            font-size: 0.75rem;
            text-transform: uppercase;
            color: #888;
            letter-spacing: 0.05em;
        }
        .info-item p {
            font-family: 'Monaco', 'Consolas', monospace;
            font-size: 0.9rem;
            color: #fff;
            word-break: break-all;
            margin-top: 0.25rem;
        }
        .pipeline {
            display: flex;
            justify-content: center;
            align-items: center;
            gap: 0.5rem;
            flex-wrap: wrap;
        }
        .pipeline-step {
            background: rgba(255, 255, 255, 0.1);
            padding: 0.75rem 1rem;
            border-radius: 8px;
            font-size: 0.85rem;
            white-space: nowrap;
        }
        .pipeline-arrow {
            color: #00d9ff;
            font-size: 1.5rem;
        }
        .footer {
            text-align: center;
            margin-top: 2rem;
            font-size: 0.85rem;
            color: #666;
        }
        .footer a {
            color: #00d9ff;
            text-decoration: none;
        }
        .footer a:hover { text-decoration: underline; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🚀 EpochCloud Test</h1>
            <p>GitOps proof-of-concept demonstrating the full CI/CD pipeline</p>
            <span class="env-badge env-{{.Environment}}">{{.Environment}}</span>
        </div>

        <div class="card">
            <h2>📦 Build Information</h2>
            <div class="info-grid">
                <div class="info-item">
                    <label>Version</label>
                    <p>{{.Version}}</p>
                </div>
                <div class="info-item">
                    <label>Commit</label>
                    <p>{{.Commit}}</p>
                </div>
                <div class="info-item">
                    <label>Build Time</label>
                    <p>{{.BuildTime}}</p>
                </div>
                <div class="info-item">
                    <label>Hostname</label>
                    <p>{{.Hostname}}</p>
                </div>
            </div>
        </div>

        <div class="card">
            <h2>⚡ Deployment Pipeline</h2>
            <p style="color: #888; margin-bottom: 1rem;">This app was deployed through the following automated pipeline:</p>
            <div class="pipeline">
                <span class="pipeline-step">📝 Git Push</span>
                <span class="pipeline-arrow">→</span>
                <span class="pipeline-step">⚙️ Argo Workflows</span>
                <span class="pipeline-arrow">→</span>
                <span class="pipeline-step">🐳 Harbor Registry</span>
                <span class="pipeline-arrow">→</span>
                <span class="pipeline-step">📦 Kargo</span>
                <span class="pipeline-arrow">→</span>
                <span class="pipeline-step">🚀 ArgoCD</span>
            </div>
        </div>

        <div class="card">
            <h2>🔗 Related Links</h2>
            <div class="info-grid">
                <div class="info-item">
                    <label>Source Code</label>
                    <p><a href="https://github.com/EpochBoy/epochcloud-test" style="color: #00d9ff;">github.com/EpochBoy/epochcloud-test</a></p>
                </div>
                <div class="info-item">
                    <label>Infrastructure</label>
                    <p><a href="https://github.com/EpochBoy/epochcloud" style="color: #00d9ff;">github.com/EpochBoy/epochcloud</a></p>
                </div>
            </div>
        </div>

        <div class="footer">
            <p>Last refreshed: {{.Timestamp}}</p>
            <p style="margin-top: 0.5rem;">Built with ❤️ using <a href="https://talos.dev">Talos</a>, <a href="https://argoproj.github.io/cd/">ArgoCD</a>, and <a href="https://kargo.io">Kargo</a></p>
        </div>
    </div>
</body>
</html>`

func getEnvironment() string {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "dev"
	}
	return env
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	resp := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func versionHandler(w http.ResponseWriter, r *http.Request) {
	hostname, _ := os.Hostname()
	resp := VersionResponse{
		Version:     Version,
		Commit:      Commit,
		BuildTime:   BuildTime,
		Hostname:    hostname,
		Environment: getEnvironment(),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	hostname, _ := os.Hostname()
	data := PageData{
		Version:     Version,
		Commit:      Commit,
		BuildTime:   BuildTime,
		Hostname:    hostname,
		Environment: getEnvironment(),
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
	}

	tmpl, err := template.New("page").Parse(pageTemplate)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl.Execute(w, data)
}

// metricsMiddleware wraps handlers to record metrics
func metricsMiddleware(path string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap ResponseWriter to capture status code
		wrapped := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next(wrapped, r)

		duration := time.Since(start).Seconds()
		httpRequestsTotal.WithLabelValues(r.Method, path, http.StatusText(wrapped.statusCode)).Inc()
		httpRequestDuration.WithLabelValues(r.Method, path).Observe(duration)
	}
}

// statusRecorder wraps http.ResponseWriter to capture the status code
type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Record build info metric
	appInfo.WithLabelValues(Version, Commit, BuildTime, getEnvironment()).Set(1)

	// Application endpoints with metrics middleware
	http.HandleFunc("/", metricsMiddleware("/", rootHandler))
	http.HandleFunc("/health", metricsMiddleware("/health", healthHandler))
	http.HandleFunc("/version", metricsMiddleware("/version", versionHandler))

	// Prometheus metrics endpoint (not wrapped to avoid recursive metrics)
	http.Handle("/metrics", promhttp.Handler())

	log.Printf("Starting epochcloud-test v%s on port %s", Version, port)
	log.Printf("Metrics available at /metrics")
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
// test-1766232964
// test2-1766233046
// test3-1766233243
// Test Sat Dec 20 14:08:44 CET 2025
// Test commit status Sat Dec 20 14:23:18 CET 2025
// Test summary fix Sat Dec 20 14:37:37 CET 2025
