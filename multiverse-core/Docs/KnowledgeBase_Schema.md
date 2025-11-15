# 🧠 KnowledgeBase Schema

> `KnowledgeBase` — **только факты**, никаких намерений или интерпретаций.  
> Сохраняется в MinIO как `application/json; charset=utf-8`.

---

## 📦 Структура снапшота

    {
      "scope_id": "player:123",
      "world_time": 1763136000000,
      "entities": {
        "player:123": {
          "id": "player:123",
          "type": "player",
          "state": { "hp": 45, "location": "alley" },
          "parameters": { "perception": 0.3, "fear": 0.6 },
          "last_updated": 1763135995000
        }
      },
      "canon": [
        "Бессмертие нарушает Закон Сохранения Души"
      ],
      "event_log": [
        {
          "event_id": "evt-abc123",
          "timestamp": 1763135990000,
          "type": "player.moved",
          "source": "player:123",
          "target": "alley"
        }
      ],
      "last_mood": ["tense", "sudden"],
      "metadata": {
        "snapshot_version": "v1.1",
        "gm_config_hash": "sha256:abc123...",
        "created_at": 1763136000100
      }
    }

---

## 📌 Поля

| Поле | Описание |
|------|----------|
| `scope_id` | ID области |
| `entities` | Состояния и параметры сущностей |
| `canon` | Факты мира (законы, история) |
| `event_log` | Последние 100 событий |
| `last_mood` | Атмосфера от LLM (факт для continuity) |
| `metadata` | Версия, хэш конфига, время создания |