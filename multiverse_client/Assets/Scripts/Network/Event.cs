[System.Serializable]
public class Event
{
    public string event_id;
    public string event_type;
    public double timestamp;
    public string source;
    public string target;
    public string world_id;
    public string scope_id;
    public EventPayload payload;
}

[System.Serializable]
public class EventPayload
{
    public string text;
    public string skill;
    public float modifier;
    public string emotion;
    public string[] mentions;
    public string result;
}