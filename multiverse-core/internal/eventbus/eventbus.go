package eventbus

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/segmentio/kafka-go"
)

type EventBus struct {
	writers map[string]*kafka.Writer
	brokers []string
}

func NewEventBus(brokers []string) *EventBus {
	topics := []string{
		TopicPlayerEvents,
		TopicWorldEvents,
		TopicGameEvents,
		TopicSystemEvents,
		TopicScopeManagement,
		TopicNarrativeOutput,
	}
	writers := make(map[string]*kafka.Writer)
	for _, topic := range topics {
		writers[topic] = &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		}
	}
	return &EventBus{
		writers: writers,
		brokers: brokers,
	}
}

func (eb *EventBus) Publish(ctx context.Context, topic string, event Event) error {
	if event.EventID == "" || event.EventType == "" || event.WorldID == "" {
		return fmt.Errorf("event missing required fields: event_id=%q, event_type=%q, world_id=%q",
			event.EventID, event.EventType, event.WorldID)
	}
	msg, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}
	return eb.writers[topic].WriteMessages(ctx, kafka.Message{
		Key:   []byte(event.WorldID),
		Value: msg,
	})
}

func (eb *EventBus) Subscribe(ctx context.Context, topic, groupID string, handler func(Event)) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  eb.brokers,
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})
	defer reader.Close()
	log.Printf("Subscribed to %s as %s", topic, groupID)
	for {
		m, err := reader.ReadMessage(ctx)
		if err != nil {
			select {
			case <-ctx.Done():
				log.Printf("Subscription to %s stopped: %v", topic, ctx.Err())
				return
			default:
				log.Printf("Read error on %s: %v", topic, err)
			}
			continue
		}
		var event Event
		if err := json.Unmarshal(m.Value, &event); err != nil {
			log.Printf("Parse error on %s key=%s: %v", topic, string(m.Key), err)
			continue
		}
		handler(event)
	}
}

func (eb *EventBus) Close() error {
	var errs []error
	for topic, writer := range eb.writers {
		if err := writer.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close writer for topic %s: %w", topic, err))
		}
	}
	if len(errs) > 0 {
		for _, err := range errs[1:] {
			log.Printf("Additional close error: %v", err)
		}
		return errs[0]
	}
	return nil
}

func (eb *EventBus) PublishPlayerEvent(ctx context.Context, event Event) error {
	return eb.Publish(ctx, TopicPlayerEvents, event)
}

func (eb *EventBus) PublishWorldEvent(ctx context.Context, event Event) error {
	return eb.Publish(ctx, TopicWorldEvents, event)
}

func (eb *EventBus) PublishGameEvent(ctx context.Context, event Event) error {
	return eb.Publish(ctx, TopicGameEvents, event)
}

func (eb *EventBus) PublishSystemEvent(ctx context.Context, event Event) error {
	return eb.Publish(ctx, TopicSystemEvents, event)
}

func (eb *EventBus) PublishNarrativeEvent(ctx context.Context, event Event) error {
	return eb.Publish(ctx, TopicNarrativeOutput, event)
}
