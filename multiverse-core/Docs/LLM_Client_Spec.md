# 🤖 LLM Client Specification

> **LLM Client — универсальный HTTP-адаптер к `/v1/chat/completions`.**  
> Совместим с Aliyun DashScope, Ollama, vLLM, LMSYS и любыми OpenAI-совместимыми API.

---

## 🔌 Конфигурация

| Переменная | Обязательная | Пример (Aliyun DashScope) | Примечание |
|-----------|--------------|----------------------------|------------|
| `ORACLE_URL` | ✅ | `https://dashscope-intl.aliyuncs.com/compatible-mode/v1/chat/completions` | URL endpoint |
| `ORACLE_MODEL` | ✅ | `qwen-plus` | Или `qwen3-max`, `qwen-turbo` и т.д. |
| `ORACLE_API_KEY` | ✅ | `sk-4659b9ed72ba489a81244ba02659b3de` | Ключ авторизации |
| `ORACLE_TIMEOUT_MS` | ❌ | `10000` | Таймаут запроса (мс) |
| `ORACLE_MAX_TOKENS` | ❌ | `1024` | Ограничение длины ответа |

> 🔹 Все параметры — **только через переменные окружения**.  
> 🔹 GM **никогда не хранит API-ключи в коде или конфигах**.

---

## 📥 Запрос: `NarrativeRequest` → `ChatCompletion`

GM формирует HTTP-запрос:

    POST /compatible-mode/v1/chat/completions
    Host: dashscope-intl.aliyuncs.com
    Authorization: Bearer sk-4659b9ed72ba489a81244ba02659b3de
    Content-Type: application/json

    {
      "model": "qwen-plus",
      "messages": [
        { "role": "system", "content": "Ты — Повествователь Мира..." },
        { "role": "user", "content": "### ЗАДАЧА\nПодумай..." }
      ],
      "temperature": 0.7,
      "max_tokens": 1024,
      "response_format": { "type": "json_object" }
    }

> ✅ Aliyun DashScope **поддерживает** `response_format: { "type": "json_object" }` для Qwen-серии.

---

## 📤 Ответ: `ChatCompletion` → `NarrativeResponse`

Клиент ожидает:

    {
      "choices": [{
        "message": {
          "content": "{\n  \"narrative\": \"Внезапно...\",\n  \"new_events\": [...]\n}"
        }
      }]
    }

→ Парсит `content` как JSON → валидирует → передаёт GM.

---

## 🧪 Пример: запуск GM с вашими параметрами

```bash
export ORACLE_URL="https://dashscope-intl.aliyuncs.com/compatible-mode/v1/chat/completions"
export ORACLE_MODEL="qwen-plus"
export ORACLE_API_KEY="sk-4659b9ed72ba489a81244ba02659b3de"
export ORACLE_TIMEOUT_MS=10000

./narrative-orchestrator \
  --scope_id="player:123" \
  --config="configs/gm_player.yaml"