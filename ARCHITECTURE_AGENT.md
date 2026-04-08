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

## 🛠️ Logic Flow

ClawdFace is designed to be a direct bridge between decentralized configurations and stateless agents:

1.  **Unpacking**: The agent receives the "Mega-Token" from the Participant Metadata.
2.  **Stateless Routing**: The agent splits the token back into the three required parts (URL, Token, Key).
3.  **Forwarding**: The agent's LLM plugin authenticates and forwards the request directly to the user's specific OpenClaw backend.

### Key Benefits
- **Compatibility**: The "Mega-Token" strategy allows standard LLM plugins to support custom headers like `x-openclaw-session-key` by unpacking them just-in-time.
- **Statelessness**: No user data is saved on the server; the agent just acts as a real-time translator for the session.

---

## 📦 AI Stack Integration

The agent integrates the following best-in-class providers:

- **TTS**: ElevenLabs (model `eleven_flash_v2_5`) with **Dynamic Voice Selection**.
- **Voice Logic**: The agent identifies the user's selected avatar and switches the ElevenLabs voice ID based on gender (e.g., Male: `CwhRBWXzGAHq8TQ4Fs17`, Female: `FGY2WhTYpPnrIDTdsKH5`).
- **Avatar**: Trugen AI (real-time video avatar generation).

---

## 🏃 Running the Agent

### Development Mode
```bash
python agent.py dev
```
In dev mode, the agent joins the room and prints real-time logs of the "Mega-Token" unpacking activity.

### Environment Requirements
The agent requires the following keys set in your system or `.env` file:
- `LIVEKIT_URL=...`
- `LIVEKIT_API_KEY=...`
- `LIVEKIT_API_SECRET=...`
- `DEEPGRAM_API_KEY=...`
- `ELEVEN_API_KEY=...`
- `ELEVEN_VOICE_ID_MALE=...`
- `ELEVEN_VOICE_ID_FEMALE=...`
- `OPENAI_API_KEY=...`
- `TRUGEN_API_KEY=...`
- `TRUGEN_AVATAR_ID=...`
