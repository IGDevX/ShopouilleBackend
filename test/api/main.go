package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type SimulationStatus struct {
	IsRunning       bool      `json:"isRunning"`
	ActiveUsers     int       `json:"activeUsers"`
	SimulationClass string    `json:"simulationClass,omitempty"`
	StartTime       time.Time `json:"startTime,omitempty"`
	ReportPath      string    `json:"reportPath,omitempty"`
}

type SimulationRequest struct {
	SimulationClass string `json:"simulationClass"`
}

type GatlingAPI struct {
	status            SimulationStatus
	mutex             sync.RWMutex
	projectDir        string
	reportsDir        string
	activeUsersGauge  prometheus.Gauge
	simulationRunning prometheus.Gauge
}

func NewGatlingAPI(projectDir string) *GatlingAPI {
	// Créer les métriques Prometheus
	activeUsersGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "gatling_active_users",
		Help: "Number of active users in the current simulation",
	})

	simulationRunning := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "gatling_simulation_running",
		Help: "Whether a simulation is currently running (1) or not (0)",
	})

	// Enregistrer les métriques
	prometheus.MustRegister(activeUsersGauge)
	prometheus.MustRegister(simulationRunning)

	return &GatlingAPI{
		projectDir:        projectDir,
		reportsDir:        filepath.Join(projectDir, "target", "gatling"),
		activeUsersGauge:  activeUsersGauge,
		simulationRunning: simulationRunning,
		status: SimulationStatus{
			IsRunning:   false,
			ActiveUsers: 0,
		},
	}
}

func (api *GatlingAPI) parseActiveUsers(output string) int {
	// Plusieurs patterns pour capturer les utilisateurs actifs
	patterns := []string{
		`active:\s+(\d+)`,         // Pattern: active: 123
		`(\d+)\s+active`,          // Pattern: 123 active
		`users\s+active:\s+(\d+)`, // Pattern: users active: 123
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(output)
		if len(matches) > 1 {
			var users int
			fmt.Sscanf(matches[1], "%d", &users)
			return users
		}
	}
	return -1 // Retourne -1 si aucun pattern ne correspond (pour ne pas réinitialiser à 0)
}

func (api *GatlingAPI) findLatestReport() string {
	entries, err := os.ReadDir(api.reportsDir)
	if err != nil {
		return ""
	}

	var latestDir string
	var latestTime time.Time

	for _, entry := range entries {
		if entry.IsDir() {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			if info.ModTime().After(latestTime) {
				latestTime = info.ModTime()
				latestDir = entry.Name()
			}
		}
	}

	if latestDir != "" {
		return filepath.Join(api.reportsDir, latestDir, "index.html")
	}
	return ""
}

func (api *GatlingAPI) runGatling(simulationClass string) error {
	log.Printf("[GATLING] Starting simulation: %s", simulationClass)

	api.mutex.Lock()
	api.status.IsRunning = true
	api.status.SimulationClass = simulationClass
	api.status.StartTime = time.Now()
	api.status.ActiveUsers = 0
	api.mutex.Unlock()

	api.simulationRunning.Set(1)
	api.activeUsersGauge.Set(0)

	defer func() {
		log.Printf("[GATLING] Simulation completed")
		api.mutex.Lock()
		api.status.IsRunning = false
		api.status.ActiveUsers = 0
		reportPath := api.findLatestReport()
		api.status.ReportPath = reportPath
		api.mutex.Unlock()

		api.simulationRunning.Set(0)
		api.activeUsersGauge.Set(0)
		log.Printf("[GATLING] Report available at: %s", reportPath)
	}()

	cmd := exec.Command(
		"./mvnw", "gatling:test",
		"-Dgatling.simulationClass="+simulationClass,
	)
	cmd.Dir = api.projectDir

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil {
		return err
	}

	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := stdout.Read(buf)
			if n > 0 {
				output := string(buf[:n])
				// Écrire directement sur stdout pour que Docker capture les logs
				log.Print(output)

				activeUsers := api.parseActiveUsers(output)
				if activeUsers >= 0 {
					api.mutex.Lock()
					api.status.ActiveUsers = activeUsers
					api.mutex.Unlock()
					api.activeUsersGauge.Set(float64(activeUsers))
					log.Printf("[GATLING] Active users: %d", activeUsers)
				}
			}
			if err != nil {
				if err != io.EOF {
					log.Printf("Error reading stdout: %v", err)
				}
				break
			}
		}
	}()

	return cmd.Wait()
}

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Type, Content-Length")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "3600")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next(w, r)
	}
}

func (api *GatlingAPI) handleStatus(w http.ResponseWriter, r *http.Request) {
	api.mutex.RLock()
	status := api.status
	api.mutex.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (api *GatlingAPI) handleStartSimulation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	api.mutex.RLock()
	isRunning := api.status.IsRunning
	api.mutex.RUnlock()

	if isRunning {
		http.Error(w, "A simulation is already running", http.StatusConflict)
		return
	}

	var req SimulationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.SimulationClass == "" {
		http.Error(w, "simulationClass is required", http.StatusBadRequest)
		return
	}

	go func() {
		if err := api.runGatling(req.SimulationClass); err != nil {
			log.Printf("Error running simulation: %v", err)
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"message":         "Simulation started",
		"simulationClass": req.SimulationClass,
	})
}

func (api *GatlingAPI) handleGetReport(w http.ResponseWriter, r *http.Request) {
	api.mutex.RLock()
	reportPath := api.status.ReportPath
	api.mutex.RUnlock()

	if reportPath == "" {
		http.Error(w, "No report available", http.StatusNotFound)
		return
	}

	reportDir := filepath.Dir(reportPath)
	requestedFile := r.URL.Query().Get("file")

	if requestedFile == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"reportPath": reportPath,
			"reportUrl":  "/report?file=index.html",
		})
		return
	}

	filePath := filepath.Join(reportDir, filepath.Clean(requestedFile))

	if !filepath.HasPrefix(filePath, reportDir) {
		http.Error(w, "Invalid file path", http.StatusForbidden)
		return
	}

	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	ext := filepath.Ext(filePath)
	contentType := "text/html"
	switch ext {
	case ".css":
		contentType = "text/css"
	case ".js":
		contentType = "application/javascript"
	case ".json":
		contentType = "application/json"
	case ".png":
		contentType = "image/png"
	}
	w.Header().Set("Content-Type", contentType)

	io.Copy(w, file)
}

func (api *GatlingAPI) handleMetrics(w http.ResponseWriter, r *http.Request) {
	api.mutex.RLock()
	activeUsers := api.status.ActiveUsers
	isRunning := api.status.IsRunning
	api.mutex.RUnlock()

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(map[string]interface{}{
		"timestamp":   time.Now().Unix(),
		"activeUsers": activeUsers,
		"isRunning":   isRunning,
	})
}

func main() {
	projectDir := os.Getenv("PROJECT_DIR")
	if projectDir == "" {
		projectDir = "/app/load-test"
	}
	api := NewGatlingAPI(projectDir)

	http.HandleFunc("/status", corsMiddleware(api.handleStatus))
	http.HandleFunc("/start", corsMiddleware(api.handleStartSimulation))
	http.HandleFunc("/report", corsMiddleware(api.handleGetReport))
	http.HandleFunc("/metrics/active-users", corsMiddleware(api.handleMetrics))

	http.Handle("/metrics", promhttp.Handler())

	port := ":8080"
	log.Printf("Gatling API server starting on port %s", port)
	log.Printf("Endpoints:")
	log.Printf("  GET  /status - Get simulation status and active users")
	log.Printf("  POST /start  - Start a new simulation")
	log.Printf("  GET  /report - Get the latest report")
	log.Printf("  GET  /metrics - Prometheus metrics endpoint")
	log.Printf("  GET  /metrics/active-users - Get active users metrics (JSON, legacy)")

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}
