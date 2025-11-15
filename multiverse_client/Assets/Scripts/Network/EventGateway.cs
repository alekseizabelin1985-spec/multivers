using System;
using System.Net.WebSockets;
using System.Text;
using System.Threading;
using System.Threading.Tasks;
using UnityEngine;

public class EventGateway : MonoBehaviour
{
    public static EventGateway Instance;
    public static event Action<Event> OnEventReceived;

    private ClientWebSocket ws;
    private string baseUrl = "ws://localhost:8088/ws/events";
    private CancellationTokenSource cts;

    void Awake() => Instance = this;

    async void Start()
    {
        cts = new CancellationTokenSource();
        await ConnectAndListen();
    }

    void OnDestroy() => cts?.Cancel();

    async Task ConnectAndListen()
    {
        ws = new ClientWebSocket();
        var url = $"{baseUrl}?player_id={PlayerPrefs.GetString("player_id", "player-777")}";
        await ws.ConnectAsync(new Uri(url), cts.Token);

        var buffer = new byte[8192];
        while (ws.State == WebSocketState.Open && !cts.Token.IsCancellationRequested)
        {
            try
            {
                var result = await ws.ReceiveAsync(buffer, cts.Token);
                var json = Encoding.UTF8.GetString(buffer, 0, result.Count);
                var ev = JsonUtility.FromJson<Event>(json);
                OnEventReceived?.Invoke(ev);
            }
            catch (OperationCanceledException) { break; }
            catch (Exception ex) { Debug.LogError($"WebSocket error: {ex}"); }
        }
    }

    public void SendEvent(Event ev)
    {
        ev.event_id = Guid.NewGuid().ToString("N").Substring(0, 8);
        var json = JsonUtility.ToJson(ev);
        ws.SendAsync(new ArraySegment<byte>(Encoding.UTF8.GetBytes(json)),
                     WebSocketMessageType.Text, true, cts.Token);
    }
}