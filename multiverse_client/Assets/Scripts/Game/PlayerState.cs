using System.Collections.Generic;
using UnityEngine;

public static class PlayerState
{
    // Кэш: entity_id → имя (из entity.updated)
    public static Dictionary<string, string> KnownEntities = new()
    {
        ["player-777"] = "Кайн",
        ["npc-elder"] = "Старец",
        ["group-wanderers"] = "Отряд Странников"
    };

    // Локальное состояние (для UI и проверок)
    public static float Fatigue = 0.85f;
    public static float Injuries = 0.6f;
}