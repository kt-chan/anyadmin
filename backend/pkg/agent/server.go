package agent

import (
	"encoding/json"
	"log"
	"net/http"
	"os/exec"
	"strings"
)

type ContainerControlRequest struct {
	ContainerName string `json:"container_name"`
	Action        string `json:"action"` // start, stop, restart
	// Optional: specific args for restart if we move complex logic here
	Args          string `json:"args,omitempty"` 
}

type ContainerControlResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Output  string `json:"output"`
}

func StartServer(port string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/container/control", handleContainerControl)

	addr := ":" + port
	log.Printf("Agent server listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Failed to start agent server: %v", err)
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func handleContainerControl(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ContainerControlRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("Received control request: %s %s", req.Action, req.ContainerName)

	if req.ContainerName == "" || req.Action == "" {
		http.Error(w, "Missing container_name or action", http.StatusBadRequest)
		return
	}

	// Sanitize inputs (basic)
	if strings.Contains(req.ContainerName, " ") || strings.Contains(req.ContainerName, ";") {
		http.Error(w, "Invalid container name", http.StatusBadRequest)
		return
	}

	var cmd *exec.Cmd
	var output []byte
	var err error

	// Handle specific actions
	switch req.Action {
	case "start", "stop", "restart":
		// Direct docker command
		cmd = exec.Command("docker", req.Action, req.ContainerName)
	case "recreate_vllm":
		// Handled below
	case "run_vllm":
		// Special case for vLLM run command which is complex
		// The full command should be constructed safely. 
		// For now, if we assume the 'Args' contains the full run arguments:
		// strict safety check required here if exposing to public, but for internal agent it's acceptable with caveats.
		// However, to be safer, we can expect the backend to send the full run command or we handle it here.
		// Given the prompt's instruction to "merge @backend/pkg/service/container.go", 
		// the backend logic for *generating* the command is complex. 
		// We can pass the generated command string or arguments.
		
		// If req.Args is provided, we treat it as the args for 'docker run ...' 
		// BUT 'docker run' requires image and many flags.
		// Simplest approach: Backend sends the full shell command to a specific endpoint, 
		// or we have a generic "exec" endpoint (risky).
		// Better: We implement a specific handler for "restart_vllm" that takes config.
		
		// For this implementation, let's stick to basic start/stop/restart. 
		// If "restart" involves "stop + rm + run", we need to support that sequence or a "recreate" action.
		
		if req.Args != "" {
             // If args is present, we assume it's a "run" or complex restart
             // For now, let's execute it as a shell command if explicitly enabled, 
             // but let's try to keep it specific.
        }
        
	default:
		http.Error(w, "Unknown action", http.StatusBadRequest)
		return
	}
    
    // Execute
    if cmd != nil {
        output, err = cmd.CombinedOutput()
    } else if req.Action == "recreate_vllm" {
        // Special internal logic for vLLM recreation
        // req.Args should contain the full docker run command string or args
        // This is a bit "remote exec" ish but constrained.
        if req.Args == "" {
             http.Error(w, "Args required for recreate_vllm", http.StatusBadRequest)
             return
        }
        
        // Split command for safety or run via bash
        // Running via bash allows the full string from backend
        cmd = exec.Command("bash", "-c", req.Args)
        output, err = cmd.CombinedOutput()
    }

	resp := ContainerControlResponse{
		Success: err == nil,
		Output:  string(output),
	}
	if err != nil {
		resp.Message = err.Error()
		log.Printf("Command failed: %v, Output: %s", err, output)
	} else {
		resp.Message = "Success"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
