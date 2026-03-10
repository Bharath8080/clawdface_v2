# Stateless Agent Architecture: agent.py

The `agent.py` file represents the "Brain" of the ClawdFace platform. It is engineered to be **100% stateless**, meaning it does not maintain its own database or configuration file. Instead, it "bootstraps" itself using information provided by the frontend in real-time.

---

## 💎 The "Mega-Token" Strategy

To remain stateless while supporting custom OpenClaw configurations (which require a URL, a Gateway Token, and a Session Key), the agent uses a **Mega-Token** strategy.

### 1. Metadata Injection (Frontend)
When the user clicks "Start Session", the frontend packs their config into a JSON object and embeds it into the **LiveKit Participant Metadata** before joining the room.

### 2. Configuration Recovery (Agent)
As soon as the agent joins the room, it iterates through participants and reads this metadata:
```python
# Unpacking from metadata
for p in ctx.room.remote_participants.values():
    if p.metadata:
        config = json.loads(p.metadata)
```

### 3. Token Packing
Since most LLM plugins (like OpenAI) only expect a single `api_key` string, we pack the entire configuration into one string separated by pipes:
`URL | GATEWAY_TOKEN | SESSION_KEY`

---

## 🛠️ The OpenClaw Flask Proxy

ClawdFace includes an internal **Flask Proxy** running inside the agent process (Port 8080). This proxy solves the "Stateless Bridge" problem:

1.  **Unpacking**: It receives the "Mega-Token" from the Agent's LLM plugin as a Bearer token.
2.  **Stateless Routing**: It splits the token back into the three required parts.
3.  **Forwarding**: It authenticates and forwards the request to the user's specific OpenClaw backend.

### Why a Proxy?
- **Compatibility**: Standard LLM plugins don't support custom headers like `x-openclaw-session-key`.
- **Statelessness**: The proxy doesn't save anything; it just acts as a real-time translator for the session.

---

## 📦 AI Stack Integration

The agent integrates the following best-in-class providers:

- **STT**: Deepgram Nova-3 (via `deepgram/nova-3`).
- **LLM**: Custom OpenAI compatible provider (OpenClaw) via the stateless proxy.
- **TTS**: ElevenLabs (model `eleven_flash_v2_5`) with specific voice tuning.
- **Avatar**: Trugen AI (real-time video avatar generation).

---

## 🏃 Running the Agent

### Development Mode
```bash
python agent.py dev
```
In dev mode, the agent joins the room and prints real-time logs of the "Mega-Token" unpacking and proxy activity.

### Environment Requirements
The agent requires the following keys set in your system or `.env` file:
- `LIVEKIT_URL`, `LIVEKIT_API_KEY`, `LIVEKIT_API_SECRET`
- `DEEPGRAM_API_KEY`
- `ELEVEN_API_KEY`
- `OPENAI_API_KEY` (Used for the OpenAI plugin wrapper)
- `TRUGEN_API_KEY` & `TRUGEN_AVATAR_ID`
