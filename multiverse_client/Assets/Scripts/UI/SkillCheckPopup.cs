using UnityEngine;
using TMPro;
using System.Collections;

public class SkillCheckPopup : MonoBehaviour
{
    public TMP_Text label, resultText;
    public GameObject success, failure;

    void OnEnable() => EventGateway.OnEventReceived += HandleEvent;
    void OnDisable() => EventGateway.OnEventReceived -= HandleEvent;

    void HandleEvent(Event ev)
    {
        if (ev.event_type == "player.skill_check.request")
        {
            gameObject.SetActive(true);
            label.text = $"Проверка: <b>{ev.payload.skill}</b>\nМодификатор: {ev.payload.modifier:+0;-#}";
            StartCoroutine(DelayedRollCheck(ev, 1.0f)); // ✅
        }
    }

    IEnumerator DelayedRollCheck(Event ev, float delay)
    {
        yield return new WaitForSeconds(delay);
        RollCheck(ev);
    }

    void RollCheck(Event ev)
    {
        int dice = Random.Range(1, 21);
        int total = dice + (int)ev.payload.modifier;
        bool success = total >= 15;

        resultText.text = $"{dice} <color=gray>+ {ev.payload.modifier}</color> = <b>{total}</b>";
        this.success.SetActive(success);
        this.failure.SetActive(!success);

        var response = new Event
        {
            event_type = "player.skill_check.result",
            payload = new EventPayload { skill = ev.payload.skill, result = success ? "success" : "failure" }
        };
        EventGateway.Instance.SendEvent(response);

        Invoke(nameof(Hide), 2f);
    }

    void Hide() => gameObject.SetActive(false);
}