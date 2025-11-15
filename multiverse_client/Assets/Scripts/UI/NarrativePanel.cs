using TMPro;
using UnityEngine;
using UnityEngine.Playables;

public class NarrativePanel : MonoBehaviour
{
    public TMP_Text narrativeText;
    [SerializeField] private TextMeshProUGUI textToUse;
    private Coroutine currentTyping;

    void OnEnable() => EventGateway.OnEventReceived += HandleEvent;
    void OnDisable() => EventGateway.OnEventReceived -= HandleEvent;

    void HandleEvent(Event ev)
    {
        if (ev.event_type == "narrative.description")
        {
            if (currentTyping != null) StopCoroutine(currentTyping);
            var resolvedText = ResolveEntityNames(ev.payload.text);
            currentTyping = StartCoroutine(TypeText(resolvedText, ev.payload.emotion));
        }
    }

    string ResolveEntityNames(string text)
    {
        // Подстановка: "${player-777}" → "Кайн"
        foreach (var id in PlayerState.KnownEntities.Keys)
        {
            var name = PlayerState.KnownEntities[id];
            text = text.Replace($"${{{id}}}", $"<color=#ffcc00>{name}</color>");
        }
        return text;
    }

    System.Collections.IEnumerator TypeText(string text, string emotion)
    {
        narrativeText.text = "";
        float delay = emotion == "dread" ? 0.04f : 0.02f;
        narrativeText.color = new Color(narrativeText.color.r, narrativeText.color.g, narrativeText.color.b, 1);

        foreach (char c in text)
        {
            narrativeText.text += c;
            PlaySoundForChar(c, emotion);
            yield return new WaitForSeconds(delay);
        }

        // Эффекты по эмоции
        if (emotion == "dread") StartCoroutine(ShakeText());

        yield return new WaitForSeconds(2f);
        StartCoroutine(FadeOutText(0.5f, narrativeText));
    }

    private System.Collections.IEnumerator FadeInText(float timeSpeed, TextMeshProUGUI text)
    {
        text.color = new Color(text.color.r, text.color.g, text.color.b, 0);
        while (text.color.a < 1.0f)
        {
            text.color = new Color(text.color.r, text.color.g, text.color.b, text.color.a + (Time.deltaTime * timeSpeed));
            yield return null;
        }
    }
    private System.Collections.IEnumerator FadeOutText(float timeSpeed, TMP_Text text)
    {
        text.color = new Color(text.color.r, text.color.g, text.color.b, 1);
        while (text.color.a > 0.0f)
        {
            text.color = new Color(text.color.r, text.color.g, text.color.b, text.color.a - (Time.deltaTime * timeSpeed));
            yield return null;
        }
        text.text = "";
    }

    void PlaySoundForChar(char c, string emotion)
    {
        if (emotion == "anger" && c == '!') AudioSource.PlayClipAtPoint(Resources.Load<AudioClip>("Sounds/sword_clash"), Vector3.zero);
        if (emotion == "whisper" && char.IsLetter(c)) AudioSource.PlayClipAtPoint(Resources.Load<AudioClip>("Sounds/whisper"), Vector3.zero, 0.3f);
    }

    System.Collections.IEnumerator ShakeText()
    {
        var orig = narrativeText.rectTransform.anchoredPosition;
        for (int i = 0; i < 10; i++)
        {
            narrativeText.rectTransform.anchoredPosition += new Vector2(Random.Range(-2f, 2f), 0);
            yield return new WaitForSeconds(0.05f);
        }
        narrativeText.rectTransform.anchoredPosition = orig;
    }
}