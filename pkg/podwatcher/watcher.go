package podwatcher

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

// Watcher watches for pod changes using Kubernetes watch API
type Watcher struct {
	clientset      *kubernetes.Clientset
	namespace      string
	labelSelector  string
	backendPort    int
	updateInterval time.Duration
	useWatch       bool

	mu            sync.RWMutex
	lastBackends  []string
	backendsChan  chan []string
	errorChan     chan error
	stopChan      chan struct{}
}

// New creates a new pod watcher
func New(clientset *kubernetes.Clientset, namespace, labelSelector string, backendPort int, updateInterval time.Duration, useWatch bool) *Watcher {
	return &Watcher{
		clientset:      clientset,
		namespace:      namespace,
		labelSelector:  labelSelector,
		backendPort:    backendPort,
		updateInterval: updateInterval,
		useWatch:       useWatch,
		backendsChan:   make(chan []string, 10),
		errorChan:      make(chan error, 10),
		stopChan:       make(chan struct{}),
	}
}

// Watch starts watching for pod changes
func (w *Watcher) Watch(ctx context.Context) (<-chan []string, <-chan error) {
	if w.useWatch {
		go w.watchWithAPI(ctx)
	} else {
		go w.watchWithPolling(ctx)
	}
	return w.backendsChan, w.errorChan
}

// watchWithAPI uses Kubernetes watch API for real-time updates
func (w *Watcher) watchWithAPI(ctx context.Context) {
	slog.Info("Starting pod watcher with Kubernetes watch API",
		"namespace", w.namespace,
		"labelSelector", w.labelSelector)

	for {
		select {
		case <-ctx.Done():
			slog.Info("Pod watcher stopped")
			close(w.backendsChan)
			close(w.errorChan)
			return
		default:
			if err := w.runWatch(ctx); err != nil {
				slog.Error("Watch error, restarting", "error", err)
				w.errorChan <- err
				time.Sleep(5 * time.Second)
			}
		}
	}
}

// runWatch runs a single watch session
func (w *Watcher) runWatch(ctx context.Context) error {
	watcher, err := w.clientset.CoreV1().Pods(w.namespace).Watch(ctx, metav1.ListOptions{
		LabelSelector: w.labelSelector,
	})
	if err != nil {
		return fmt.Errorf("failed to start watch: %w", err)
	}
	defer watcher.Stop()

	// Get initial list
	if err := w.updateBackendList(ctx); err != nil {
		return fmt.Errorf("failed to get initial pod list: %w", err)
	}

	// Watch for changes
	for {
		select {
		case <-ctx.Done():
			return nil
		case event, ok := <-watcher.ResultChan():
			if !ok {
				return fmt.Errorf("watch channel closed")
			}
			w.handleWatchEvent(ctx, event)
		}
	}
}

// handleWatchEvent processes a watch event
func (w *Watcher) handleWatchEvent(ctx context.Context, event watch.Event) {
	switch event.Type {
	case watch.Added, watch.Modified, watch.Deleted:
		if err := w.updateBackendList(ctx); err != nil {
			slog.Error("Failed to update backend list", "error", err)
			w.errorChan <- err
		}
	case watch.Error:
		slog.Error("Received error event from watch")
	}
}

// watchWithPolling uses traditional polling for pod discovery
func (w *Watcher) watchWithPolling(ctx context.Context) {
	slog.Info("Starting pod watcher with polling",
		"namespace", w.namespace,
		"labelSelector", w.labelSelector,
		"interval", w.updateInterval)

	ticker := time.NewTicker(w.updateInterval)
	defer ticker.Stop()

	// Initial update
	if err := w.updateBackendList(ctx); err != nil {
		slog.Error("Failed initial pod discovery", "error", err)
		w.errorChan <- err
	}

	for {
		select {
		case <-ctx.Done():
			slog.Info("Pod watcher stopped")
			close(w.backendsChan)
			close(w.errorChan)
			return
		case <-ticker.C:
			if err := w.updateBackendList(ctx); err != nil {
				slog.Error("Failed to update backend list", "error", err)
				w.errorChan <- err
			}
		}
	}
}

// updateBackendList fetches the current pod list and updates backends if changed
func (w *Watcher) updateBackendList(ctx context.Context) error {
	pods, err := w.clientset.CoreV1().Pods(w.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: w.labelSelector,
	})
	if err != nil {
		return fmt.Errorf("failed to list pods: %w", err)
	}

	backends := w.extractBackends(pods.Items)

	w.mu.Lock()
	defer w.mu.Unlock()

	// Sort for comparison
	sort.Strings(backends)
	sort.Strings(w.lastBackends)

	// Check if backends changed
	if !equal(backends, w.lastBackends) {
		slog.Info("Pod backends changed",
			"old_count", len(w.lastBackends),
			"new_count", len(backends))

		w.lastBackends = backends

		// Send to channel (non-blocking)
		select {
		case w.backendsChan <- backends:
		default:
			slog.Warn("Backend channel full, skipping update")
		}
	}

	return nil
}

// extractBackends extracts backend addresses from pod list
func (w *Watcher) extractBackends(pods []corev1.Pod) []string {
	var backends []string
	for _, pod := range pods {
		if pod.Status.Phase == corev1.PodRunning && pod.Status.PodIP != "" {
			backend := fmt.Sprintf("%s:%d", pod.Status.PodIP, w.backendPort)
			backends = append(backends, backend)
		}
	}
	return backends
}

// Close stops the watcher
func (w *Watcher) Close() error {
	close(w.stopChan)
	return nil
}

// equal compares two string slices (assumes both are sorted)
func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
