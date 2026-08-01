package main

import (
	"embed"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"log"
	"net/http"
	"strings"

	"glowing-agent/simulator"
)

//go:embed static/*
var staticFiles embed.FS

type simulationRequest struct {
	Task     string `json:"task"`
	PresetID string `json:"presetId"`
	Seed     *int64 `json:"seed"`
}

func main() {
	server, err := newServer()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Glowing Agent is pretending to work at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", server))
}

func newServer() (http.Handler, error) {
	assets, err := fs.Sub(staticFiles, "static")
	if err != nil {
		return nil, err
	}

	mux := http.NewServeMux()
	mux.Handle("GET /", http.FileServer(http.FS(assets)))
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "delightfully simulated"})
	})
	mux.HandleFunc("POST /api/simulations", handleSimulation)
	return mux, nil
}

func handleSimulation(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	r.Body = http.MaxBytesReader(w, r.Body, 8<<10)
	var request simulationRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "Send a valid JSON simulation request.")
		return
	}
	if err := ensureSingleJSONValue(decoder); err != nil {
		writeError(w, http.StatusBadRequest, "Send exactly one JSON object.")
		return
	}

	task := strings.TrimSpace(request.Task)
	if request.PresetID != "" {
		preset, ok := simulator.PresetByID(request.PresetID)
		if !ok {
			writeError(w, http.StatusBadRequest, "That preset escaped the backlog.")
			return
		}
		task = preset.Task
	}
	if len(task) == 0 || len(task) > 1000 {
		writeError(w, http.StatusBadRequest, "Task text must be between 1 and 1000 characters.")
		return
	}

	result := simulator.Generate(task, request.Seed)
	writeJSON(w, http.StatusOK, result)
}

func ensureSingleJSONValue(decoder *json.Decoder) error {
	var extra any
	err := decoder.Decode(&extra)
	if errors.Is(err, io.EOF) {
		return nil
	}
	return errors.New("unexpected trailing JSON")
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("write response: %v", err)
	}
}
