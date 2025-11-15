package realitymonitor

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"multiverse-core/internal/eventbus"
)

// Service represents the Reality Monitor service
type Service struct {
	eventBus     *eventbus.EventBus
	state        *State
	ctx          context.Context
	cancel       context.CancelFunc
}

// State holds the current state of the reality monitor
type State struct {
	Metrics map[string]*WorldMetrics
}

// WorldMetrics holds aggregated metrics for a world
type WorldMetrics struct {
	WorldID           string
	LastUpdated       time.Time
	SpatialIntegrity  float64
	KarmaEntropy      float64
	CoreResonance     float64
	AnomalyDetected   bool
	AnomalyType       string
	AnomalyTimestamp  time.Time
}

// NewService creates a new Reality Monitor service
func NewService(eventBus *eventbus.EventBus) *Service {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &Service{
		eventBus: eventBus,
		state: &State{
			Metrics: make(map[string]*WorldMetrics),
		},
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start starts the Reality Monitor service
func (s *Service) Start() error {
	log.Println("Starting Reality Monitor service...")
	
	// Subscribe to world metrics events
	go s.eventBus.Subscribe(s.ctx, "world.metrics.*", "reality-monitor-group", s.handleWorldMetricsEvent)
	
	// Subscribe to system events for anomaly detection
	go s.eventBus.Subscribe(s.ctx, "reality.anomaly.detected", "reality-monitor-group", s.handleAnomalyEvent)
	
	go s.run()
	
	log.Println("Reality Monitor service started successfully")
	return nil
}

// Stop stops the Reality Monitor service
func (s *Service) Stop() error {
	s.cancel()
	log.Println("Reality Monitor service stopped")
	return nil
}

// run runs the main loop of the service
func (s *Service) run() {
	ticker := time.NewTicker(30 * time.Second) // Check every 30 seconds
	defer ticker.Stop()
	
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.checkForAnomalies()
		}
	}
}

// handleWorldMetricsEvent handles incoming world metrics events
func (s *Service) handleWorldMetricsEvent(event eventbus.Event) {
	var metrics WorldMetrics
	
	// Convert payload to JSON bytes then back to map for proper parsing
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		log.Printf("Error marshaling event payload: %v", err)
		return
	}
	
	// Parse the event payload
	if err := json.Unmarshal(payloadBytes, &metrics); err != nil {
		log.Printf("Error parsing world metrics event: %v", err)
		return
	}
	
	// Update metrics in state
	s.state.Metrics[metrics.WorldID] = &metrics
	
	log.Printf("Updated metrics for world %s: spatial=%f, karma=%f, resonance=%f", 
		metrics.WorldID, metrics.SpatialIntegrity, metrics.KarmaEntropy, metrics.CoreResonance)
}

// handleAnomalyEvent handles anomaly detection events
func (s *Service) handleAnomalyEvent(event eventbus.Event) {
	log.Printf("Received anomaly detection event: %s", event.EventType)
	
	// Process anomaly event
	// This could trigger alerts, notifications, or other actions
}

// checkForAnomalies performs periodic anomaly detection
func (s *Service) checkForAnomalies() {
	log.Println("Checking for anomalies...")
	
	for worldID, metrics := range s.state.Metrics {
		if s.isAnomaly(metrics) {
			// Prepare anomaly data as map for payload
			anomalyData := map[string]interface{}{
				"world_id":     worldID,
				"anomaly_type": metrics.AnomalyType,
				"timestamp":    time.Now().Format(time.RFC3339),
			}
			
			// Publish anomaly detected event
			anomalyEvent := eventbus.Event{
				EventID:   "anomaly-" + time.Now().Format("20060102150405"),
				EventType: "reality.anomaly.detected",
				Source:    "reality-monitor",
				WorldID:   worldID,
				Payload:   anomalyData,
				Timestamp: time.Now(),
			}
			
			if err := s.eventBus.PublishSystemEvent(s.ctx, anomalyEvent); err != nil {
				log.Printf("Failed to publish anomaly event: %v", err)
			} else {
				log.Printf("Published anomaly detected event for world %s", worldID)
			}
		}
	}
}

// isAnomaly determines if metrics indicate an anomaly
func (s *Service) isAnomaly(metrics *WorldMetrics) bool {
	// Check for spatial integrity anomalies
	if metrics.SpatialIntegrity < 0.1 || metrics.SpatialIntegrity > 0.9 {
		metrics.AnomalyType = "spatial_integrity"
		metrics.AnomalyDetected = true
		metrics.AnomalyTimestamp = time.Now()
		return true
	}
	
	// Check for karma entropy anomalies
	if metrics.KarmaEntropy > 0.9 {
		metrics.AnomalyType = "karma_entropy"
		metrics.AnomalyDetected = true
		metrics.AnomalyTimestamp = time.Now()
		return true
	}
	
	// Check for core resonance anomalies
	if metrics.CoreResonance < 0.3 || metrics.CoreResonance > 1.0 {
		metrics.AnomalyType = "core_resonance"
		metrics.AnomalyDetected = true
		metrics.AnomalyTimestamp = time.Now()
		return true
	}
	
	// Reset anomaly flag if no issues
	metrics.AnomalyDetected = false
	metrics.AnomalyType = ""
	
	return false
}

// GetWorldMetrics returns metrics for a specific world
func (s *Service) GetWorldMetrics(worldID string) (*WorldMetrics, bool) {
	metrics, exists := s.state.Metrics[worldID]
	return metrics, exists
}

// GetAllMetrics returns all world metrics
func (s *Service) GetAllMetrics() map[string]*WorldMetrics {
	return s.state.Metrics
}