package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	"golang.org/x/exp/slices"
	"log/slog"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {
	// Initialize structured logger
	handler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)

	labels := os.Getenv("POD_LABELS")
	if labels == "" {
		slog.Error("POD_LABELS environment variable is not set")
		os.Exit(1)
	}

	traefikAPIURL := os.Getenv("TRAEFIK_API_URL")
	if traefikAPIURL == "" {
		slog.Error("TRAEFIK_API_URL environment variable is not set")
		os.Exit(1)
	}

	namespace := os.Getenv("POD_NAMESPACE")
	if namespace == "" {
		namespaceBytes, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
		if err != nil {
			slog.Error("Error reading namespace", "error", err)
			os.Exit(1)
		}
		namespace = strings.TrimSpace(string(namespaceBytes))
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		slog.Error("Error building in-cluster config", "error", err)
		os.Exit(1)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		slog.Error("Error building Kubernetes clientset", "error", err)
		os.Exit(1)
	}

	// Use signal.NotifyContext for graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Get update interval from environment variable
	updateIntervalStr := os.Getenv("UPDATE_INTERVAL")
	if updateIntervalStr == "" {
		updateIntervalStr = "1s"
	}
	updateInterval, err := time.ParseDuration(updateIntervalStr)
	if err != nil {
		slog.Error("Invalid UPDATE_INTERVAL, using default 1s", "error", err)
		updateInterval = time.Second
	}

	ticker := time.NewTicker(updateInterval)
	defer ticker.Stop()

	var lastBackends []string

	for {
		select {
		case <-ctx.Done():
			slog.Info("Received termination signal, exiting...")
			return
		case <-ticker.C:
			if err := updateTraefikBackends(ctx, clientset, labels, namespace, traefikAPIURL, &lastBackends); err != nil {
				slog.Error("Error updating Traefik backends", "error", err)
			}
		}
	}
}

func updateTraefikBackends(ctx context.Context, clientset *kubernetes.Clientset, labels, namespace, traefikAPIURL string, lastBackends *[]string) error {
	pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labels,
	})

	if err != nil {
		return fmt.Errorf("error listing pods: %w", err)
	}

	var backends []string
	for _, pod := range pods.Items {
		if pod.Status.Phase == "Running" && pod.Status.PodIP != "" {
			backend := fmt.Sprintf("%s:3333", pod.Status.PodIP)
			backends = append(backends, backend)
		}
	}

	sort.Strings(backends)
	sort.Strings(*lastBackends)

	// Use slices.Equal for comparison
	if slices.Equal(backends, *lastBackends) {
		// Backends have not changed
		return nil
	}

	if len(backends) == 0 {
		slog.Info("No running pods found with the specified labels")
	} else {
		slog.Info("Discovered pods", "count", len(backends))
	}

	if err := updateTraefikConfig(backends, traefikAPIURL); err != nil {
		return fmt.Errorf("error updating Traefik config: %w", err)
	}

	*lastBackends = backends
	return nil
}

func updateTraefikConfig(backends []string, traefikAPIURL string) error {
	// Build servers slice outside the map for readability
	servers := []map[string]string{}
	for _, backend := range backends {
		servers = append(servers, map[string]string{
			"address": backend,
		})
	}

	// Use 'any' instead of 'interface{}' for readability
config := map[string]any{
    "tcp": map[string]any{
        "routers": map[string]any{
            "relay-router": map[string]any{
                "entryPoints": []string{"tcp"},
                "rule":        "HostSNI(`*`)",
                "service":     "relay-service",
            },
        },
        "services": map[string]any{
            "relay-service": map[string]any{
                "loadBalancer": map[string]any{
                    "method":   "leastconn",
                    "servers": servers,
                },
            },
        },
    },
}

	jsonData, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("error marshaling Traefik config to JSON: %w", err)
	}

	req, err := http.NewRequest("PUT", traefikAPIURL, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("error creating HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Include response body in error for better diagnostics
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("received unexpected status code from Traefik API: %d, body: %s", resp.StatusCode, string(body))
	}

	slog.Info("Updated Traefik configuration via API", "url", traefikAPIURL)
	return nil
}
