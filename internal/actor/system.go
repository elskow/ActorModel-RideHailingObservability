package actor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"actor-model-observability/internal/logging"
)

// SupervisionStrategy defines how to handle actor failures
type SupervisionStrategy string

const (
	SupervisionRestart SupervisionStrategy = "restart"
	SupervisionStop    SupervisionStrategy = "stop"
	SupervisionIgnore  SupervisionStrategy = "ignore"
)

// ActorRef represents a reference to an actor
type ActorRef struct {
	ID       string
	Type     string
	Actor    Actor
	Strategy SupervisionStrategy
}

// SystemMetrics holds metrics for the entire actor system
type SystemMetrics struct {
	TotalActors       int           `json:"total_actors"`
	ActiveActors      int           `json:"active_actors"`
	TotalMessages     int64         `json:"total_messages"`
	MessagesPerSecond float64       `json:"messages_per_second"`
	AverageLatency    time.Duration `json:"average_latency"`
	SystemUptime      time.Duration `json:"system_uptime"`
	LastMetricsUpdate time.Time     `json:"last_metrics_update"`
}

// ActorSystem manages a collection of actors
type ActorSystem struct {
	name         string
	actors       map[string]*ActorRef
	actorsMutex  sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
	logger       *logging.Logger
	metrics      SystemMetrics
	metricsLock  sync.RWMutex
	startTime    time.Time
	wg           sync.WaitGroup
	started      bool
	startedMutex sync.RWMutex

	// Event handlers
	onActorStarted func(actorID string)
	onActorStopped func(actorID string)
	onActorFailed  func(actorID string, err error)
	onMessage      func(from, to, messageType string)
}

// NewActorSystem creates a new actor system
func NewActorSystem(name string) *ActorSystem {
	return &ActorSystem{
		name:   name,
		actors: make(map[string]*ActorRef),
		logger: logging.GetGlobalLogger().WithComponent("actor_system").WithField("system", name),
		metrics: SystemMetrics{
			LastMetricsUpdate: time.Now(),
		},
	}
}

// Start initializes and starts the actor system
func (s *ActorSystem) Start(ctx context.Context) error {
	s.startedMutex.Lock()
	defer s.startedMutex.Unlock()

	if s.started {
		s.logger.Warn("Actor system is already started")
		return nil
	}

	s.ctx, s.cancel = context.WithCancel(ctx)
	s.startTime = time.Now()
	s.started = true

	// Start metrics collection goroutine
	s.wg.Add(1)
	go s.metricsCollector()

	s.logger.Info("Actor system started")
	return nil
}

// IsStarted returns whether the actor system is currently started
func (s *ActorSystem) IsStarted() bool {
	s.startedMutex.RLock()
	defer s.startedMutex.RUnlock()
	return s.started
}

// Stop gracefully shuts down the actor system
func (s *ActorSystem) Stop() error {
	s.startedMutex.Lock()
	defer s.startedMutex.Unlock()

	if !s.started {
		s.logger.Warn("Actor system is already stopped")
		return nil
	}

	s.logger.Info("Stopping actor system")

	// Cancel context
	if s.cancel != nil {
		s.cancel()
	}

	// Stop all actors
	s.actorsMutex.Lock()
	for _, actorRef := range s.actors {
		if err := actorRef.Actor.Stop(); err != nil {
			s.logger.WithError(err).Error("Failed to stop actor", "actor_id", actorRef.ID)
		}
	}
	s.actorsMutex.Unlock()

	// Wait for all goroutines to finish
	s.wg.Wait()

	s.started = false
	s.logger.Info("Actor system stopped")
	return nil
}

// SpawnActor creates and starts a new actor
func (s *ActorSystem) SpawnActor(actorType, actorID string, mailboxSize int, handler func(Message) error, strategy SupervisionStrategy) (*ActorRef, error) {
	if actorID == "" {
		return nil, fmt.Errorf("actor ID cannot be empty")
	}

	s.actorsMutex.Lock()
	defer s.actorsMutex.Unlock()

	// Check if actor already exists
	if _, exists := s.actors[actorID]; exists {
		return nil, fmt.Errorf("actor with ID %s already exists", actorID)
	}

	// Create new actor
	actor := NewBaseActor(actorID, actorType, mailboxSize, handler)
	actorRef := &ActorRef{
		ID:       actorID,
		Type:     actorType,
		Actor:    actor,
		Strategy: strategy,
	}

	// Start the actor
	if err := actor.Start(s.ctx); err != nil {
		return nil, fmt.Errorf("failed to start actor %s: %w", actorID, err)
	}

	// Add to actors map
	s.actors[actorID] = actorRef

	// Update metrics
	s.updateMetrics(func(m *SystemMetrics) {
		m.TotalActors++
		m.ActiveActors++
	})

	// Trigger event handler
	if s.onActorStarted != nil {
		s.onActorStarted(actorID)
	}

	s.logger.WithFields(logging.Fields{
		"actor_id":   actorID,
		"actor_type": actorType,
	}).Info("Actor spawned")

	return actorRef, nil
}

// StopActor stops and removes an actor from the system
func (s *ActorSystem) StopActor(actorID string) error {
	s.actorsMutex.Lock()
	defer s.actorsMutex.Unlock()

	actorRef, exists := s.actors[actorID]
	if !exists {
		return fmt.Errorf("actor %s not found", actorID)
	}

	// Stop the actor
	if err := actorRef.Actor.Stop(); err != nil {
		return fmt.Errorf("failed to stop actor %s: %w", actorID, err)
	}

	// Remove from actors map
	delete(s.actors, actorID)

	// Update metrics
	s.updateMetrics(func(m *SystemMetrics) {
		m.ActiveActors--
	})

	// Trigger event handler
	if s.onActorStopped != nil {
		s.onActorStopped(actorID)
	}

	s.logger.WithField("actor_id", actorID).Info("Actor stopped")
	return nil
}

// GetActor returns an actor reference by ID
func (s *ActorSystem) GetActor(actorID string) (*ActorRef, error) {
	s.actorsMutex.RLock()
	defer s.actorsMutex.RUnlock()

	actorRef, exists := s.actors[actorID]
	if !exists {
		return nil, fmt.Errorf("actor %s not found", actorID)
	}

	return actorRef, nil
}

// SendMessage sends a message to an actor
func (s *ActorSystem) SendMessage(toActorID string, message Message) error {
	actorRef, err := s.GetActor(toActorID)
	if err != nil {
		return err
	}

	if err := actorRef.Actor.Send(message); err != nil {
		return fmt.Errorf("failed to send message to actor %s: %w", toActorID, err)
	}

	// Update metrics
	s.updateMetrics(func(m *SystemMetrics) {
		m.TotalMessages++
	})

	// Trigger event handler
	if s.onMessage != nil {
		s.onMessage(message.GetSender(), toActorID, message.GetType())
	}

	return nil
}

// BroadcastMessage sends a message to all actors of a specific type
func (s *ActorSystem) BroadcastMessage(actorType string, message Message) error {
	s.actorsMutex.RLock()
	var targetActors []*ActorRef
	for _, actorRef := range s.actors {
		if actorRef.Type == actorType {
			targetActors = append(targetActors, actorRef)
		}
	}
	s.actorsMutex.RUnlock()

	if len(targetActors) == 0 {
		return fmt.Errorf("no actors of type %s found", actorType)
	}

	var errors []error
	for _, actorRef := range targetActors {
		if err := actorRef.Actor.Send(message); err != nil {
			errors = append(errors, fmt.Errorf("failed to send to %s: %w", actorRef.ID, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("broadcast failed for some actors: %v", errors)
	}

	// Update metrics
	s.updateMetrics(func(m *SystemMetrics) {
		m.TotalMessages += int64(len(targetActors))
	})

	return nil
}

// ListActors returns a list of all actor references
func (s *ActorSystem) ListActors() []*ActorRef {
	s.actorsMutex.RLock()
	defer s.actorsMutex.RUnlock()

	var actors []*ActorRef
	for _, actorRef := range s.actors {
		actors = append(actors, actorRef)
	}

	return actors
}

// GetMetrics returns current system metrics
func (s *ActorSystem) GetMetrics() SystemMetrics {
	s.metricsLock.RLock()
	defer s.metricsLock.RUnlock()

	metrics := s.metrics
	if !s.startTime.IsZero() {
		metrics.SystemUptime = time.Since(s.startTime)
	}

	return metrics
}

// SetEventHandlers sets event handlers for system events
func (s *ActorSystem) SetEventHandlers(
	onActorStarted func(actorID string),
	onActorStopped func(actorID string),
	onActorFailed func(actorID string, err error),
	onMessage func(from, to, messageType string),
) {
	s.onActorStarted = onActorStarted
	s.onActorStopped = onActorStopped
	s.onActorFailed = onActorFailed
	s.onMessage = onMessage
}

// metricsCollector runs in a goroutine to collect system metrics
func (s *ActorSystem) metricsCollector() {
	defer s.wg.Done()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	lastMessageCount := int64(0)
	lastUpdate := time.Now()

	for {
		select {
		case <-ticker.C:
			now := time.Now()
			duration := now.Sub(lastUpdate)

			s.updateMetrics(func(m *SystemMetrics) {
				// Calculate messages per second
				messageDiff := m.TotalMessages - lastMessageCount
				if duration.Seconds() > 0 {
					m.MessagesPerSecond = float64(messageDiff) / duration.Seconds()
				}

				// Calculate average latency from all actors
				s.actorsMutex.RLock()
				var totalLatency time.Duration
				activeCount := 0
				for _, actorRef := range s.actors {
					if actorRef.Actor.GetState() == ActorStateProcessing {
						actorMetrics := actorRef.Actor.GetMetrics()
						totalLatency += actorMetrics.AverageProcessTime
						activeCount++
					}
				}
				s.actorsMutex.RUnlock()

				if activeCount > 0 {
					m.AverageLatency = totalLatency / time.Duration(activeCount)
				}

				m.LastMetricsUpdate = now
			})

			lastMessageCount = s.metrics.TotalMessages
			lastUpdate = now

		case <-s.ctx.Done():
			return
		}
	}
}

func (s *ActorSystem) updateMetrics(updater func(*SystemMetrics)) {
	s.metricsLock.Lock()
	defer s.metricsLock.Unlock()
	updater(&s.metrics)
}
