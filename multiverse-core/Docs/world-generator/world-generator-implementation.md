# 🛠️ Реализация расширенного генератора мира

## Обзор изменений

Расширенный генератор мира будет включать следующие новые возможности:
- Генерация регионов с различными биомами
- Генерация водных объектов (реки, моря, озера)
- Генерация городов с базовыми характеристиками
- Публикация событий для интеграции с другими сервисами

## Структура данных

### Географическая структура мира

```go
// WorldGeography представляет полную географическую структуру мира
type WorldGeography struct {
    Core        string        `json:"core"`
    Ontology    WorldOntology `json:"ontology"`
    Geography   Geography     `json:"geography"`
    Mythology   string        `json:"mythology"`
}

// WorldOntology представляет онтологию культивации мира
type WorldOntology struct {
    Carriers  []string `json:"carriers"`
    Paths     []string `json:"paths"`
    Forbidden []string `json:"forbidden"`
}

// Geography представляет географическую структуру мира
type Geography struct {
    Regions     []Region     `json:"regions"`
    WaterBodies []WaterBody  `json:"water_bodies"`
    Cities      []City       `json:"cities"`
}

// Region представляет регион мира
type Region struct {
    Name        string  `json:"name"`
    Biome       string  `json:"biome"`
    Coordinates Point   `json:"coordinates"`
    Size        float64 `json:"size"`
}

// WaterBody представляет водный объект
type WaterBody struct {
    Name        string  `json:"name"`
    Type        string  `json:"type"` // river, sea, lake
    Coordinates Point   `json:"coordinates"`
    Size        float64 `json:"size"`
}

// City представляет город
type City struct {
    Name       string  `json:"name"`
    Population int     `json:"population"`
    Type       string  `json:"type"` // major, minor
    Location   Location `json:"location"`
}

// Location представляет местоположение
type Location struct {
    Region      string `json:"region"`
    Coordinates Point  `json:"coordinates"`
}

// Point представляет точку координат
type Point struct {
    X float64 `json:"x"`
    Y float64 `json:"y"`
}
```

## Основные методы

### generateEnhancedWorldDetails

```go
// generateEnhancedWorldDetails генерирует расширенные детали мира через Ascension Oracle
func (wg *WorldGenerator) generateEnhancedWorldDetails(ctx context.Context, worldID, worldSeed string) {
    prompt := fmt.Sprintf(`
Создай детали мира с семенем "%s".

Требуется сгенерировать:
1. Ядро Мира (1-2 предложения)
2. Онтологию культивации (носители, пути, запреты)
3. Географию:
   - 3-5 регионов с уникальными биомами (леса, горы, поля, пустыни, болота)
   - Водные объекты (2-4 реки, 1-2 моря, 1-3 озера)
   - 2-4 города с основными характеристиками
4. Мифологию (краткий миф)

Верни строго в JSON:
{
  "core": "string",
  "ontology": { 
    "carriers": ["string"], 
    "paths": ["string"], 
    "forbidden": ["string"] 
  },
  "geography": {
    "regions": [
      {
        "name": "string",
        "biome": "string",
        "coordinates": {"x": 0.0, "y": 0.0},
        "size": 0.0
      }
    ],
    "water_bodies": [
      {
        "name": "string",
        "type": "river|sea|lake",
        "coordinates": {"x": 0.0, "y": 0.0},
        "size": 0.0
      }
    ],
    "cities": [
      {
        "name": "string",
        "population": 0,
        "type": "major|minor",
        "location": {
          "region": "string",
          "coordinates": {"x": 0.0, "y": 0.0}
        }
      }
    ]
  },
  "mythology": "string"
}
`, worldSeed)

    resp, err := CallOracle(ctx, prompt)
    if err != nil {
        log.Printf("Oracle world details failed: %v", err)
        return
    }

    // Парсинг ответа
    var geography WorldGeography
    if err := json.Unmarshal([]byte(resp.Narrative), &geography); err != nil {
        log.Printf("Failed to parse geography: %v", err)
        return
    }

    // Создание сущностей и публикация событий
    wg.createGeographicEntities(ctx, worldID, geography)
}
```

### createGeographicEntities

```go
// createGeographicEntities создает сущности для географических объектов
func (wg *WorldGenerator) createGeographicEntities(ctx context.Context, worldID string, geography WorldGeography) {
    // Создание регионов
    for _, region := range geography.Geography.Regions {
        wg.createRegionEntity(ctx, worldID, region)
    }

    // Создание водных объектов
    for _, water := range geography.Geography.WaterBodies {
        wg.createWaterEntity(ctx, worldID, water)
    }

    // Создание городов
    for _, city := range geography.Geography.Cities {
        wg.createCityEntity(ctx, worldID, city)
    }

    // Публикация события о завершении генерации географии
    wg.publishGeographyGeneratedEvent(ctx, worldID, geography)
}
```

### createRegionEntity

```go
// createRegionEntity создает сущность региона
func (wg *WorldGenerator) createRegionEntity(ctx context.Context, worldID string, region Region) {
    regionID := "region-" + uuid.New().String()[:8]
    
    regionEvent := eventbus.Event{
        EventID:   "region-create-" + uuid.New().String()[:8],
        EventType: "entity.created",
        Source:    "world-generator",
        WorldID:   worldID,
        Payload: map[string]interface{}{
            "entity_id":   regionID,
            "entity_type": "region",
            "payload": map[string]interface{}{
                "name":        region.Name,
                "biome":       region.Biome,
                "coordinates": region.Coordinates,
                "size":        region.Size,
            },
        },
        Timestamp: time.Now(),
    }
    
    wg.bus.Publish(ctx, eventbus.TopicSystemEvents, regionEvent)
    log.Printf("Created region entity: %s (%s)", region.Name, region.Biome)
}
```

### createWaterEntity

```go
// createWaterEntity создает сущность водного объекта
func (wg *WorldGenerator) createWaterEntity(ctx context.Context, worldID string, water WaterBody) {
    waterID := "water-" + uuid.New().String()[:8]
    
    waterEvent := eventbus.Event{
        EventID:   "water-create-" + uuid.New().String()[:8],
        EventType: "entity.created",
        Source:    "world-generator",
        WorldID:   worldID,
        Payload: map[string]interface{}{
            "entity_id":   waterID,
            "entity_type": "water_body",
            "payload": map[string]interface{}{
                "name":        water.Name,
                "type":        water.Type,
                "coordinates": water.Coordinates,
                "size":        water.Size,
            },
        },
        Timestamp: time.Now(),
    }
    
    wg.bus.Publish(ctx, eventbus.TopicSystemEvents, waterEvent)
    log.Printf("Created water entity: %s (%s)", water.Name, water.Type)
}
```

### createCityEntity

```go
// createCityEntity создает сущность города
func (wg *WorldGenerator) createCityEntity(ctx context.Context, worldID string, city City) {
    cityID := "city-" + uuid.New().String()[:8]
    
    cityEvent := eventbus.Event{
        EventID:   "city-create-" + uuid.New().String()[:8],
        EventType: "entity.created",
        Source:    "world-generator",
        WorldID:   worldID,
        Payload: map[string]interface{}{
            "entity_id":   cityID,
            "entity_type": "city",
            "payload": map[string]interface{}{
                "name":       city.Name,
                "population": city.Population,
                "type":       city.Type,
                "location":   city.Location,
            },
        },
        Timestamp: time.Now(),
    }
    
    wg.bus.Publish(ctx, eventbus.TopicSystemEvents, cityEvent)
    log.Printf("Created city entity: %s (population: %d)", city.Name, city.Population)
}
```

### publishGeographyGeneratedEvent

```go
// publishGeographyGeneratedEvent публикует событие о завершении генерации географии
func (wg *WorldGenerator) publishGeographyGeneratedEvent(ctx context.Context, worldID string, geography WorldGeography) {
    geographyEvent := eventbus.Event{
        EventID:   "geography-generated-" + uuid.New().String()[:8],
        EventType: "world.geography.generated",
        Source:    "world-generator",
        WorldID:   worldID,
        Payload: map[string]interface{}{
            "world_id":   worldID,
            "regions":    len(geography.Geography.Regions),
            "water_bodies": len(geography.Geography.WaterBodies),
            "cities":     len(geography.Geography.Cities),
        },
        Timestamp: time.Now(),
    }
    
    wg.bus.Publish(ctx, eventbus.TopicSystemEvents, geographyEvent)
    log.Printf("Published geography generated event for world: %s", worldID)
}
```

## Интеграция с другими сервисами

### События для CityGovernor

CityGovernor будет подписываться на события `entity.created` с типами `city` для получения информации о новых городах и их управлении.

### События для EntityManager

EntityManager будет обрабатывать события `entity.created` для всех типов географических объектов (регионы, города, водные объекты) для их хранения и управления.

### События для BanOfWorld

BanOfWorld будет использовать события `world.geography.generated` и `entity.created` для мониторинга целостности мира.

## Мониторинг и метрики

Новые метрики для отслеживания:
- Количество сгенерированных регионов
- Количество сгенерированных водных объектов
- Количество сгенерированных городов
- Время генерации географической структуры

## Тестирование

### Unit-тесты

```go
func TestGenerateEnhancedWorldDetails(t *testing.T) {
    // Тест генерации расширенных деталей мира
}

func TestCreateGeographicEntities(t *testing.T) {
    // Тест создания географических сущностей
}

func TestCreateRegionEntity(t *testing.T) {
    // Тест создания сущности региона
}
```

### Интеграционные тесты

```go
func TestWorldGeneratorIntegration(t *testing.T) {
    // Интеграционный тест всего процесса генерации
}
```

## Развертывание

### Обновление Dockerfile

```dockerfile
# Обновление зависимостей и сборка сервиса
FROM golang:1.25 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o world-generator cmd/world-generator/main.go
```

### Обновление docker-compose.yml

```yaml
services:
  world-generator:
    build: .
    command: ./world-generator
    environment:
      - ORACLE_URL=http://ascension-oracle:8080
      - KAFKA_BROKERS=redpanda:9092
      - MINIO_ENDPOINT=minio:9000
    depends_on:
      - redpanda
      - minio
```

## Заключение


## Тестирование

Для тестирования функциональности генерации мира можно использовать следующие подходы:

1. **Автоматическое тестирование** - отправка тестовых событий через Kafka
2. **Ручное тестирование** - использование скриптов для отправки событий
3. **Интеграционное тестирование** - проверка взаимодействия с другими сервисами

См. подробности в [руководстве по тестированию](../docs/world-generator/world_generation_testing_guide.md) и [инструкции по отправке тестовых событий](../docs/world-generator/world_generation_test_script.md).

## Заключение

Реализация расширенного генератора мира позволит создавать полноценные игровые миры с детализированной географической структурой, что значительно улучшит игровой опыт и обеспечит более глубокую интеграцию с другими сервисами системы.
Реализация расширенного генератора мира позволит создавать полноценные игровые миры с детализированной географической структурой, что значительно улучшит игровой опыт и обеспечит более глубокую интеграцию с другими сервисами системы.