# Тестовое событие для генерации мира

## Описание
Тестовое событие для проверки функциональности генерации мира в системе.

## Структура события

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

## Параметры

| Параметр | Описание | Пример |
|----------|----------|--------|
| event_id | Уникальный идентификатор события | test-world-gen-001 |
| event_type | Тип события | world.generation.requested |
| source | Источник события | test-client |
| world_id | Идентификатор мира | test-world-001 |
| payload.seed | Семя для генерации мира | test-seed-12345 |
| payload.constraints | Ограничения для генерации | JSON объект с ограничениями |

## Цель теста
Проверить обработку события генерации мира сервисом WorldGenerator и корректную генерацию структуры мира.

## Как использовать
1. Отправьте это событие в тему `system_events` через Event Bus
2. Наблюдайте за публикацией событий `world.generated` и `entity.created`
3. Проверьте создание сущностей мира (регионы, города, воды)