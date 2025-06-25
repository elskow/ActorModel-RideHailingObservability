package actor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"actor-model-observability/internal/logging"

	"github.com/google/uuid"
)

// ActorState represents the current state of an actor
type ActorState string

const (
	ActorStateIdle       ActorState = "idle"
	ActorStateProcessing ActorState = "processing"
	ActorStateStopped    ActorState = "stopped"
	ActorStateError      ActorState = "error"
)

// Message represents a message that can be sent between actors
type Message interface {
	GetID() string
	GetType() string
	GetPayload() interface{}
	GetSender() string
	GetTimestamp() time.Time
}

// BaseMessage provides a basic implementation of Message
type BaseMessage struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"`
	Payload   interface{} `json:"payload"`
	Sender    string      `json:"sender"`
	Timestamp time.Time   `json:"timestamp"`
}

func NewBaseMessage(msgType string, payload interface{}, sender string) *BaseMessage {
	return &BaseMessage{
		ID:        uuid.New().String(),
		Type:      msgType,
		Payload:   payload,
		Sender:    sender,
		Timestamp: time.Now(),
	}
}

func (m *BaseMessage) GetID() string           { return m.ID }
func (m *BaseMessage) GetType() string         { return m.Type }
func (m *BaseMessage) GetPayload() interface{} { return m.Payload }
func (m *BaseMessage) GetSender() string       { return m.Sender }
func (m *BaseMessage) GetTimestamp() time.Time { return m.Timestamp }

// Actor represents the core actor interface
type Actor interface {
	GetID() string
	GetType() string
	GetState() ActorState
	Start(ctx context.Context) error
	Stop() error
	Send(message Message) error
	Receive() <-chan Message
	GetMetrics() ActorMetrics
}

// ActorMetrics holds performance metrics for an actor
type ActorMetrics struct {
	MessagesReceived   int64         `json:"messages_received"`
	MessagesProcessed  int64         `json:"messages_processed"`
	MessagesFailed     int64         `json:"messages_failed"`
	AverageProcessTime time.Duration `json:"average_process_time"`
	LastActivity       time.Time     `json:"last_activity"`
	Uptime             time.Duration `json:"uptime"`
	CurrentQueueSize   int           `json:"current_queue_size"`
}

// BaseActor provides a basic implementation of Actor
type BaseActor struct {
	id          string
	actorType   string
	state       ActorState
	mailbox     chan Message
	ctx         context.Context
	cancel      context.CancelFunc
	logger      *logging.Logger
	metrics     ActorMetrics
	metricsLock sync.RWMutex
	startTime   time.Time
	wg          sync.WaitGroup

	// Message handler function
	handler func(Message) error
}

// NewBaseActor creates a new base actor
func NewBaseActor(id, actorType string, mailboxSize int, handler func(Message) error) *BaseActor {
	if id == "" {
		id = uuid.New().String()
	}

	return &BaseActor{
		id:        id,
		actorType: actorType,
		state:     ActorStateIdle,
		mailbox:   make(chan Message, mailboxSize),
		logger:    logging.GetGlobalLogger().WithActor(id, actorType),
		handler:   handler,
		metrics: ActorMetrics{
			LastActivity: time.Now(),
		},
	}
}

func (a *BaseActor) GetID() string {
	return a.id
}

func (a *BaseActor) GetType() string {
	return a.actorType
}

func (a *BaseActor) GetState() ActorState {
	return a.state
}

func (a *BaseActor) Start(ctx context.Context) error {
	if a.state != ActorStateIdle {
		return fmt.Errorf("actor %s is already started or stopped", a.id)
	}

	a.ctx, a.cancel = context.WithCancel(ctx)
	a.startTime = time.Now()
	a.state = ActorStateProcessing

	a.wg.Add(1)
	go a.messageLoop()

	a.logger.Info("Actor started")
	return nil
}

func (a *BaseActor) Stop() error {
	if a.state == ActorStateStopped {
		return nil
	}

	a.state = ActorStateStopped
	if a.cancel != nil {
		a.cancel()
	}

	// Close mailbox to signal message loop to exit
	close(a.mailbox)

	// Wait for message loop to finish
	a.wg.Wait()

	a.logger.Info("Actor stopped")
	return nil
}

func (a *BaseActor) Send(message Message) error {
	if a.state == ActorStateStopped {
		return fmt.Errorf("actor %s is stopped", a.id)
	}

	select {
	case a.mailbox <- message:
		a.updateMetrics(func(m *ActorMetrics) {
			m.MessagesReceived++
			m.CurrentQueueSize = len(a.mailbox)
			m.LastActivity = time.Now()
		})
		return nil
	case <-a.ctx.Done():
		return fmt.Errorf("actor %s context cancelled", a.id)
	default:
		return fmt.Errorf("actor %s mailbox is full", a.id)
	}
}

func (a *BaseActor) Receive() <-chan Message {
	return a.mailbox
}

func (a *BaseActor) GetMetrics() ActorMetrics {
	a.metricsLock.RLock()
	defer a.metricsLock.RUnlock()

	metrics := a.metrics
	if !a.startTime.IsZero() {
		metrics.Uptime = time.Since(a.startTime)
	}
	metrics.CurrentQueueSize = len(a.mailbox)

	return metrics
}

func (a *BaseActor) messageLoop() {
	defer a.wg.Done()

	for {
		select {
		case message, ok := <-a.mailbox:
			if !ok {
				// Mailbox closed, exit loop
				return
			}

			a.processMessage(message)

		case <-a.ctx.Done():
			// Context cancelled, exit loop
			return
		}
	}
}

func (a *BaseActor) processMessage(message Message) {
	start := time.Now()

	a.logger.WithMessage(message.GetID(), message.GetType(), message.GetSender(), a.id).Debug("Processing message")

	var err error
	if a.handler != nil {
		err = a.handler(message)
	}

	processTime := time.Since(start)

	a.updateMetrics(func(m *ActorMetrics) {
		m.CurrentQueueSize = len(a.mailbox)
		m.LastActivity = time.Now()

		if err != nil {
			m.MessagesFailed++
			a.logger.WithError(err).Error("Message processing failed")
		} else {
			m.MessagesProcessed++
		}

		// Update average process time
		totalMessages := m.MessagesProcessed + m.MessagesFailed
		if totalMessages > 0 {
			m.AverageProcessTime = time.Duration(
				(int64(m.AverageProcessTime)*(totalMessages-1) + int64(processTime)) / totalMessages,
			)
		}
	})
}

func (a *BaseActor) updateMetrics(updater func(*ActorMetrics)) {
	a.metricsLock.Lock()
	defer a.metricsLock.Unlock()
	updater(&a.metrics)
}
