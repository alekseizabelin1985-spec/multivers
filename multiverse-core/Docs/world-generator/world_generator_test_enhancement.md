# Улучшение тестов для WorldGenerator

## Текущее состояние тестов

В файле `services/worldgenerator/generator_test.go` существуют базовые тесты:
- Тест структур данных
- Тест создания генератора
- Тест генерации расширенных деталей мира (заглушка)

## Рекомендации по улучшению

### 1. Добавить тест для обработки события генерации мира

```go
func TestWorldGeneratorHandleEvent(t *testing.T) {
    // Create a mock event bus
    mockBus := &eventbus.EventBus{}
    
    // Create world generator
    gen := NewWorldGenerator(mockBus)
    
    // Create a test event for world generation
    testEvent := eventbus.Event{
        EventID:   "test-world-gen-001",
        EventType: "world.generation.requested",
        Source:    "test-client",
        WorldID:   "test-world-001",
        Payload: map[string]interface{}{
            "seed": "test-seed-12345",
        },
        Timestamp: time.Now(),
    }
    
    // Test that the handler processes the event correctly
    // This would require mocking the event bus publishing behavior
    // and verifying that the right events are published
    gen.HandleEvent(testEvent)
    
    // Verify that appropriate events were published
    // (This would require more sophisticated mocking)
}
```

### 2. Добавить тест с валидацией параметров

```go
func TestWorldGeneratorHandleEventWithInvalidSeed(t *testing.T) {
    mockBus := &eventbus.EventBus{}
    gen := NewWorldGenerator(mockBus)
    
    // Test event with invalid seed
    testEvent := eventbus.Event{
        EventID:   "test-world-gen-invalid",
        EventType: "world.generation.requested",
        Source:    "test-client",
        WorldID:   "test-world-001",
        Payload: map[string]interface{}{
            "seed": "", // Empty seed should be rejected
        },
        Timestamp: time.Now(),
    }
    
    // Should not process event with empty seed
    gen.HandleEvent(testEvent)
    
    // Verify that no world was generated
    // (This would require mocking the event bus)
}
```

### 3. Добавить тест с ограничениями

```go
func TestWorldGeneratorHandleEventWithConstraints(t *testing.T) {
    mockBus := &eventbus.EventBus{}
    gen := NewWorldGenerator(mockBus)
    
    // Test event with constraints
    testEvent := eventbus.Event{
        EventID:   "test-world-gen-constraints",
        EventType: "world.generation.requested",
        Source:    "test-client",
        WorldID:   "test-world-001",
        Payload: map[string]interface{}{
            "seed": "test-seed-67890",
            "constraints": map[string]interface{}{
                "max_regions": 5,
                "max_cities": 3,
                "biomes": []string{"forest", "mountain"},
            },
        },
        Timestamp: time.Now(),
    }
    
    // Test processing with constraints
    gen.HandleEvent(testEvent)
    
    // Verify that constraints were handled properly
    // (Would require more advanced mocking)
}
```

## Рекомендации по реализации

1. Для полноценного тестирования необходимо создать моки для:
   - eventbus.EventBus
   - ArchivistClient
   - Oracle вызовов

2. Можно использовать библиотеку testify/mock или встроенные возможности Go для мокинга

3. Тесты должны проверять:
   - Правильную обработку события "world.generation.requested"
   - Валидацию параметров
   - Публикацию правильных событий в шину
   - Обработку ошибок