# Инструкция по отправке тестового события генерации мира

## Обзор

Эта инструкция описывает, как отправить тестовое событие для генерации мира в систему.

## Структура события

Тестовое событие должно соответствовать следующей структуре:

```json
{
  "event_id": "test-world-gen-001",
  "event_type": "world.generation.requested",
  "source": "test-client",
  "world_id": "test-world-001",
  "payload": {
    "seed": "test-seed-12345",
    "constraints": {
      "max_regions": 5,
      "max_cities": 3,
      "biomes": ["forest", "mountain", "plains"]
    }
  },
  "timestamp": "2025-11-09T20:42:00Z"
}
```

## Способы отправки события

### Способ 1: Через командную строку с использованием kafkacat

Если у вас установлен `kafkacat`, выполните команду:

```bash
echo '{
  "event_id": "test-world-gen-001",
  "event_type": "world.generation.requested",
  "source": "test-client",
  "world_id": "test-world-001",
  "payload": {
    "seed": "test-seed-12345",
    "constraints": {
      "max_regions": 5,
      "max_cities": 3,
      "biomes": ["forest", "mountain", "plains"]
    }
  },
  "timestamp": "2025-11-09T20:42:00Z"
}' | kafkacat -b redpanda:9092 -t system_events -P
```

### Способ 2: Через Python скрипт

Создайте файл `send_test_event.py`:

```python
#!/usr/bin/env python3
import json
import kafka

# Настройки подключения
bootstrap_servers = 'redpanda:9092'
topic = 'system_events'

# Тестовое событие
test_event = {
    "event_id": "test-world-gen-001",
    "event_type": "world.generation.requested",
    "source": "test-client",
    "world_id": "test-world-001",
    "payload": {
        "seed": "test-seed-12345",
        "constraints": {
            "max_regions": 5,
            "max_cities": 3,
            "biomes": ["forest", "mountain", "plains"]
        }
    },
    "timestamp": "2025-11-09T20:42:00Z"
}

# Отправка сообщения
producer = kafka.KafkaProducer(
    bootstrap_servers=bootstrap_servers,
    value_serializer=lambda v: json.dumps(v).encode('utf-8')
)

producer.send(topic, test_event)
producer.flush()
producer.close()

print("Тестовое событие отправлено успешно!")
```

### Способ 3: Через Go код

Создайте простой Go скрипт для отправки события:

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "time"

    "github.com/Shopify/sarama"
)

func main() {
    // Конфигурация Kafka
    config := sarama.NewConfig()
    config.Producer.RequiredAcks = sarama.WaitForAll
    config.Producer.Retry.Max = 10
    config.Producer.Return.Successes = true

    producer, err := sarama.NewSyncProducer([]string{"redpanda:9092"}, config)
    if err != nil {
        log.Fatalf("Ошибка создания продюсера: %v", err)
    }
    defer producer.Close()

    // Тестовое событие
    testEvent := map[string]interface{}{
        "event_id":   "test-world-gen-001",
        "event_type": "world.generation.requested",
        "source":     "test-client",
        "world_id":   "test-world-001",
        "payload": map[string]interface{}{
            "seed": "test-seed-12345",
            "constraints": map[string]interface{}{
                "max_regions": 5,
                "max_cities":  3,
                "biomes":      []string{"forest", "mountain", "plains"},
            },
        },
        "timestamp": time.Now().UTC().Format(time.RFC3339),
    }

    // Преобразование в JSON
    eventData, err := json.Marshal(testEvent)
    if err != nil {
        log.Fatalf("Ошибка marshaling события: %v", err)
    }

    // Отправка сообщения
    _, _, err = producer.SendMessage(&sarama.ProducerMessage{
        Topic: "system_events",
        Value: sarama.ByteEncoder(eventData),
    })

    if err != nil {
        log.Fatalf("Ошибка отправки сообщения: %v", err)
    }

    fmt.Println("Тестовое событие отправлено успешно!")
}
```

## Проверка результата

После отправки события проверьте:
1. Логи сервиса WorldGenerator
2. Публикацию событий `world.generated` и `entity.created` в теме `system_events`
3. Создание сущностей мира (регионы, города, воды)

## Решение проблем с парсингом JSON

При генерации мира Oracle может возвращать текст с JSON в середине, что может вызвать проблемы с парсингом. Если вы получаете ошибки парсинга JSON, убедитесь, что:

1. Oracle правильно форматирует ответ в JSON
2. Используйте тестовое семя "test-seed-12345" для проверки
3. Проверьте логи сервиса для получения деталей ошибок
4. В случае проблем с парсингом, можно временно отключить проверку JSON в коде для тестирования

## Необходимые зависимости

Для использования скриптов потребуются:
- `kafkacat` (для первого способа)
- `kafka-python` (для Python скрипта)
- `sarama` (для Go скрипта)
- Доступ к Redpanda/Kafka на порту 9092

## Дополнительные параметры

Можно модифицировать параметры тестового события:
- `seed`: уникальное семя для генерации
- `constraints`: ограничения для генерации мира
- `world_id`: идентификатор мира
- `source`: источник события