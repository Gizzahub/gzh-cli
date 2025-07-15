// Package debug provides HTTP API for dynamic log level management
package debug

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

// startHTTPServer starts the HTTP server for log level management
func (lm *LogLevelManager) startHTTPServer(port int) error {
	router := mux.NewRouter()

	// API routes
	api := router.PathPrefix("/api/v1/logging").Subrouter()

	// Profile management
	api.HandleFunc("/profiles", lm.handleGetProfiles).Methods("GET")
	api.HandleFunc("/profiles/{name}", lm.handleGetProfile).Methods("GET")
	api.HandleFunc("/profiles/{name}", lm.handleApplyProfile).Methods("POST")
	api.HandleFunc("/profiles", lm.handleCreateProfile).Methods("POST")
	api.HandleFunc("/profiles/{name}", lm.handleDeleteProfile).Methods("DELETE")

	// Rule management
	api.HandleFunc("/rules", lm.handleGetRules).Methods("GET")
	api.HandleFunc("/rules", lm.handleCreateRule).Methods("POST")
	api.HandleFunc("/rules/{id}", lm.handleGetRule).Methods("GET")
	api.HandleFunc("/rules/{id}", lm.handleUpdateRule).Methods("PUT")
	api.HandleFunc("/rules/{id}", lm.handleDeleteRule).Methods("DELETE")
	api.HandleFunc("/rules/{id}/enable", lm.handleEnableRule).Methods("POST")
	api.HandleFunc("/rules/{id}/disable", lm.handleDisableRule).Methods("POST")

	// Level management
	api.HandleFunc("/level", lm.handleGetLevel).Methods("GET")
	api.HandleFunc("/level", lm.handleSetLevel).Methods("POST")
	api.HandleFunc("/level/module/{module}", lm.handleGetModuleLevel).Methods("GET")
	api.HandleFunc("/level/module/{module}", lm.handleSetModuleLevel).Methods("POST")

	// Metrics and status
	api.HandleFunc("/metrics", lm.handleGetMetrics).Methods("GET")
	api.HandleFunc("/status", lm.handleGetStatus).Methods("GET")

	// Health check
	router.HandleFunc("/health", lm.handleHealthCheck).Methods("GET")

	// CORS middleware
	router.Use(lm.corsMiddleware)

	lm.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		if err := lm.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			lm.logger.ErrorLevel(context.Background(), "HTTP server error",
				map[string]interface{}{"error": err.Error()})
		}
	}()

	lm.logger.InfoLevel(context.Background(), "Log level HTTP server started",
		map[string]interface{}{"port": port})

	return nil
}

// CORS middleware
func (lm *LogLevelManager) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Profile handlers

func (lm *LogLevelManager) handleGetProfiles(w http.ResponseWriter, r *http.Request) {
	profiles := lm.GetProfiles()
	lm.writeJSONResponse(w, http.StatusOK, profiles)
}

func (lm *LogLevelManager) handleGetProfile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	profiles := lm.GetProfiles()
	if profile, exists := profiles[name]; exists {
		lm.writeJSONResponse(w, http.StatusOK, profile)
	} else {
		lm.writeErrorResponse(w, http.StatusNotFound, "Profile not found")
	}
}

func (lm *LogLevelManager) handleApplyProfile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	if err := lm.ApplyProfile(name); err != nil {
		lm.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	lm.writeJSONResponse(w, http.StatusOK, map[string]string{
		"message": "Profile applied successfully",
		"profile": name,
	})
}

func (lm *LogLevelManager) handleCreateProfile(w http.ResponseWriter, r *http.Request) {
	var profile LogLevelProfile
	if err := json.NewDecoder(r.Body).Decode(&profile); err != nil {
		lm.writeErrorResponse(w, http.StatusBadRequest, "Invalid profile data")
		return
	}

	profile.Created = time.Now()

	lm.mutex.Lock()
	lm.profiles[profile.Name] = &profile
	lm.mutex.Unlock()

	lm.writeJSONResponse(w, http.StatusCreated, profile)
}

func (lm *LogLevelManager) handleDeleteProfile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	// Prevent deletion of built-in profiles
	builtinProfiles := []string{"development", "production", "testing"}
	for _, builtin := range builtinProfiles {
		if name == builtin {
			lm.writeErrorResponse(w, http.StatusForbidden, "Cannot delete built-in profile")
			return
		}
	}

	lm.mutex.Lock()
	if _, exists := lm.profiles[name]; exists {
		delete(lm.profiles, name)
		lm.mutex.Unlock()
		lm.writeJSONResponse(w, http.StatusOK, map[string]string{
			"message": "Profile deleted successfully",
		})
	} else {
		lm.mutex.Unlock()
		lm.writeErrorResponse(w, http.StatusNotFound, "Profile not found")
	}
}

// Rule handlers

func (lm *LogLevelManager) handleGetRules(w http.ResponseWriter, r *http.Request) {
	rules := lm.GetRules()
	lm.writeJSONResponse(w, http.StatusOK, rules)
}

func (lm *LogLevelManager) handleGetRule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	rules := lm.GetRules()
	for _, rule := range rules {
		if rule.ID == id {
			lm.writeJSONResponse(w, http.StatusOK, rule)
			return
		}
	}

	lm.writeErrorResponse(w, http.StatusNotFound, "Rule not found")
}

func (lm *LogLevelManager) handleCreateRule(w http.ResponseWriter, r *http.Request) {
	var rule LogLevelRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		lm.writeErrorResponse(w, http.StatusBadRequest, "Invalid rule data")
		return
	}

	if err := lm.AddRule(rule); err != nil {
		lm.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	lm.writeJSONResponse(w, http.StatusCreated, rule)
}

func (lm *LogLevelManager) handleUpdateRule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var updatedRule LogLevelRule
	if err := json.NewDecoder(r.Body).Decode(&updatedRule); err != nil {
		lm.writeErrorResponse(w, http.StatusBadRequest, "Invalid rule data")
		return
	}

	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	for i, rule := range lm.rules {
		if rule.ID == id {
			updatedRule.ID = id
			updatedRule.Created = rule.Created
			lm.rules[i] = updatedRule

			// Clear rule cache
			lm.cacheMutex.Lock()
			lm.ruleCache = make(map[string]bool)
			lm.cacheMutex.Unlock()

			lm.writeJSONResponse(w, http.StatusOK, updatedRule)
			return
		}
	}

	lm.writeErrorResponse(w, http.StatusNotFound, "Rule not found")
}

func (lm *LogLevelManager) handleDeleteRule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := lm.RemoveRule(id); err != nil {
		lm.writeErrorResponse(w, http.StatusNotFound, err.Error())
		return
	}

	lm.writeJSONResponse(w, http.StatusOK, map[string]string{
		"message": "Rule deleted successfully",
	})
}

func (lm *LogLevelManager) handleEnableRule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	for i, rule := range lm.rules {
		if rule.ID == id {
			lm.rules[i].Enabled = true
			lm.writeJSONResponse(w, http.StatusOK, map[string]string{
				"message": "Rule enabled successfully",
			})
			return
		}
	}

	lm.writeErrorResponse(w, http.StatusNotFound, "Rule not found")
}

func (lm *LogLevelManager) handleDisableRule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	for i, rule := range lm.rules {
		if rule.ID == id {
			lm.rules[i].Enabled = false
			lm.writeJSONResponse(w, http.StatusOK, map[string]string{
				"message": "Rule disabled successfully",
			})
			return
		}
	}

	lm.writeErrorResponse(w, http.StatusNotFound, "Rule not found")
}

// Level handlers

func (lm *LogLevelManager) handleGetLevel(w http.ResponseWriter, r *http.Request) {
	level := lm.logger.GetLevel()
	lm.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"level":   level,
		"name":    rfc5424Names[level],
		"numeric": int(level),
	})
}

func (lm *LogLevelManager) handleSetLevel(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Level string `json:"level"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		lm.writeErrorResponse(w, http.StatusBadRequest, "Invalid request data")
		return
	}

	level, err := ParseRFC5424Severity(request.Level)
	if err != nil {
		lm.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	lm.logger.SetLevel(level)
	lm.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message": "Level set successfully",
		"level":   level,
		"name":    rfc5424Names[level],
	})
}

func (lm *LogLevelManager) handleGetModuleLevel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	module := vars["module"]

	// Get module level from current profile or config
	var level RFC5424Severity = lm.logger.GetLevel() // Default to global level

	if lm.currentProfile != nil {
		if moduleLevel, exists := lm.currentProfile.ModuleLevels[module]; exists {
			level = moduleLevel
		}
	}

	lm.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"module":  module,
		"level":   level,
		"name":    rfc5424Names[level],
		"numeric": int(level),
	})
}

func (lm *LogLevelManager) handleSetModuleLevel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	module := vars["module"]

	var request struct {
		Level string `json:"level"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		lm.writeErrorResponse(w, http.StatusBadRequest, "Invalid request data")
		return
	}

	level, err := ParseRFC5424Severity(request.Level)
	if err != nil {
		lm.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	lm.logger.SetModuleLevel(module, level)
	lm.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message": "Module level set successfully",
		"module":  module,
		"level":   level,
		"name":    rfc5424Names[level],
	})
}

// Metrics and status handlers

func (lm *LogLevelManager) handleGetMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := lm.GetMetrics()
	lm.writeJSONResponse(w, http.StatusOK, metrics)
}

func (lm *LogLevelManager) handleGetStatus(w http.ResponseWriter, r *http.Request) {
	rules := lm.GetRules()
	currentProfile := lm.GetCurrentProfile()
	metrics := lm.GetMetrics()

	enabledRules := 0
	for _, rule := range rules {
		if rule.Enabled {
			enabledRules++
		}
	}

	var profileName string
	if currentProfile != nil {
		profileName = currentProfile.Name
	}

	status := map[string]interface{}{
		"current_profile":   profileName,
		"global_level":      lm.logger.GetLevel(),
		"global_level_name": rfc5424Names[lm.logger.GetLevel()],
		"total_rules":       len(rules),
		"enabled_rules":     enabledRules,
		"metrics":           metrics,
		"uptime":            time.Since(time.Now().Add(-time.Hour)), // Placeholder
	}

	lm.writeJSONResponse(w, http.StatusOK, status)
}

func (lm *LogLevelManager) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"version":   "1.0.0",
	}
	lm.writeJSONResponse(w, http.StatusOK, health)
}

// Utility methods

func (lm *LogLevelManager) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		lm.logger.ErrorLevel(context.Background(), "Failed to encode JSON response",
			map[string]interface{}{"error": err.Error()})
	}
}

func (lm *LogLevelManager) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	errorResponse := map[string]interface{}{
		"error":     message,
		"timestamp": time.Now(),
		"status":    statusCode,
	}
	lm.writeJSONResponse(w, statusCode, errorResponse)
}

// Signal handling

func (lm *LogLevelManager) setupSignalHandling() {
	lm.signalChan = make(chan os.Signal, 1)
	signal.Notify(lm.signalChan, syscall.SIGUSR1, syscall.SIGUSR2, syscall.SIGHUP)

	go func() {
		for sig := range lm.signalChan {
			switch sig {
			case syscall.SIGUSR1:
				// Increase log level (more verbose)
				lm.adjustLogLevel(true)
			case syscall.SIGUSR2:
				// Decrease log level (less verbose)
				lm.adjustLogLevel(false)
			case syscall.SIGHUP:
				// Reload configuration / Apply default profile
				lm.ApplyProfile("development")
				lm.logger.InfoLevel(context.Background(), "Reloaded log configuration via SIGHUP")
			}
		}
	}()

	lm.logger.InfoLevel(context.Background(), "Signal handling initialized",
		map[string]interface{}{
			"signals": []string{"SIGUSR1", "SIGUSR2", "SIGHUP"},
		})
}

// adjustLogLevel adjusts the current log level up or down
func (lm *LogLevelManager) adjustLogLevel(increase bool) {
	currentLevel := lm.logger.GetLevel()
	var newLevel RFC5424Severity

	if increase {
		// Increase verbosity (lower numeric value)
		if currentLevel > SeverityEmergency {
			newLevel = currentLevel - 1
		} else {
			newLevel = currentLevel
		}
	} else {
		// Decrease verbosity (higher numeric value)
		if currentLevel < SeverityDebug {
			newLevel = currentLevel + 1
		} else {
			newLevel = currentLevel
		}
	}

	if newLevel != currentLevel {
		lm.logger.SetLevel(newLevel)
		lm.logger.InfoLevel(context.Background(), "Log level adjusted via signal",
			map[string]interface{}{
				"old_level": rfc5424Names[currentLevel],
				"new_level": rfc5424Names[newLevel],
				"direction": map[bool]string{true: "increase", false: "decrease"}[increase],
			})
	}
}

// CLI support functions

// SetLevelFromString sets log level from string (for CLI integration)
func (lm *LogLevelManager) SetLevelFromString(levelStr string) error {
	level, err := ParseRFC5424Severity(levelStr)
	if err != nil {
		return err
	}

	lm.logger.SetLevel(level)
	return nil
}

// SetModuleLevelFromString sets module log level from string (for CLI integration)
func (lm *LogLevelManager) SetModuleLevelFromString(module, levelStr string) error {
	level, err := ParseRFC5424Severity(levelStr)
	if err != nil {
		return err
	}

	lm.logger.SetModuleLevel(module, level)
	return nil
}

// ListProfiles returns profile names for CLI
func (lm *LogLevelManager) ListProfiles() []string {
	profiles := lm.GetProfiles()
	names := make([]string, 0, len(profiles))
	for name := range profiles {
		names = append(names, name)
	}
	return names
}

// GetRuleByID returns a rule by ID
func (lm *LogLevelManager) GetRuleByID(id string) (*LogLevelRule, error) {
	rules := lm.GetRules()
	for _, rule := range rules {
		if rule.ID == id {
			return &rule, nil
		}
	}
	return nil, fmt.Errorf("rule with ID '%s' not found", id)
}

// EnableRule enables a rule by ID
func (lm *LogLevelManager) EnableRule(id string) error {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	for i, rule := range lm.rules {
		if rule.ID == id {
			lm.rules[i].Enabled = true
			return nil
		}
	}

	return fmt.Errorf("rule with ID '%s' not found", id)
}

// DisableRule disables a rule by ID
func (lm *LogLevelManager) DisableRule(id string) error {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	for i, rule := range lm.rules {
		if rule.ID == id {
			lm.rules[i].Enabled = false
			return nil
		}
	}

	return fmt.Errorf("rule with ID '%s' not found", id)
}
