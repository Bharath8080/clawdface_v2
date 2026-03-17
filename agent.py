import os
import json
import requests
import asyncio
import threading
from flask import Flask, request, Response
from dotenv import load_dotenv
from livekit import agents
from livekit.agents import Agent, AgentServer, AgentSession
from livekit.plugins import elevenlabs, openai, trugen, groq, silero, deepgram
 
load_dotenv()
 
# --- OPENCLAW SESSION PROXY (Stateless / Mega-Token) ---
app = Flask(__name__)
 
@app.route('/v1/chat/completions', methods=['POST'])
def chat_proxy():
    try:
        data = request.get_json()
        messages = data.get('messages', [])
 
        # 1. Unpack "Mega-Token" from Authorization header (URL|TOKEN|KEY)
        auth_header = request.headers.get("Authorization", "")
        token_str = auth_header.replace("Bearer ", "")
       
        if "|" not in token_str:
             print("[PROXY] ✗ Invalid Mega-Token format")
             return {"error": "Missing x-openclaw-url. Check your Mega-Token."}, 400
 
        # Unpack: URL | TOKEN | SESSION_KEY
        parts = token_str.split("|")
        target_url = parts[0]
        gate_token = parts[1]
        sess_key   = parts[2]
 
        print(f"\n[PROXY] → {target_url}  session={sess_key}")
 
        # Filter to only send user messages to OpenClaw (it manages history)
        new_messages = []
        for msg in reversed(messages):
            if msg.get('role') == 'user':
                new_messages.insert(0, msg)
            else:
                break
 
        if not new_messages:
            return {"error": "No user message found"}, 400
 
        headers = {
            "Authorization": f"Bearer {gate_token}",
            "x-openclaw-session-key": sess_key,
            "x-openclaw-agent-id": "main",
            "ngrok-skip-browser-warning": "true"
        }
 
        resp = requests.post(
            f"{target_url}/v1/chat/completions",
            headers=headers,
            json={"model": "main", "messages": new_messages, "stream": data.get("stream", True)},
            stream=True,
            timeout=30
        )
 
        def generate():
            for chunk in resp.iter_content(chunk_size=1024):
                yield chunk
 
        return Response(generate(), resp.status_code, {"Content-Type": "text/event-stream"})
 
    except Exception as e:
        print(f"[PROXY] Error: {e}")
        return {"error": str(e)}, 500
 
import traceback
from livekit import api as lk_api

# ... (chat_proxy and run_proxy same as before) ...

@app.route('/join-meeting', methods=['POST'])
def join_meeting():
    try:
        data = request.get_json()

        meeting_url   = data.get("meetingUrl")
        openclaw_url  = data.get("openclawUrl")
        gateway_token = data.get("gatewayToken")
        session_key   = data.get("sessionKey")
        avatar_id     = data.get("avatarId", "1a640442")
        room_id       = data.get("roomId", f"recall-{os.urandom(4).hex()}")

        if not meeting_url:
            return {"error": "meetingUrl is required"}, 400
        if not openclaw_url:
            return {"error": "openclawUrl is required"}, 400
        if not gateway_token:
            return {"error": "gatewayToken is required"}, 400
        if not session_key:
            return {"error": "sessionKey is required"}, 400

        print(f"[JOIN] room={room_id} meeting={meeting_url}")

        # Step 1 — Create Recall bot
        recall_payload = {
            "meeting_url": meeting_url,
            "bot_name": "Lisa",
            "output_media": {
                "camera": {
                    "kind": "webpage",
                    "config": {
                        "url": (
                            f"{os.getenv('AVATAR_VIDEO_STREAM')}"
                            f"?room={room_id}"
                            f"&avatarId={avatar_id}"
                            f"&openclawUrl={openclaw_url}"
                            f"&gatewayToken={gateway_token}"
                            f"&sessionKey={session_key}"
                        )
                    }
                }
            },
            "recording_config": {
                "transcript": {
                    "provider": {
                        "assembly_ai": {}
                    }
                }
            },
            "real_time_endpoints": [
                {
                    "type": "webhook",
                    "url": f"{os.getenv('RECALL_WEBHOOK_URL')}/{room_id}"
                }
            ]
        }

        recall_resp = requests.post(
            f"{os.getenv('RECALL_API_URL')}",
            headers={
                "Authorization": f"Token {os.getenv('RECALL_API_TOKEN')}",
                "Content-Type": "application/json"
            },
            json=recall_payload,
            timeout=15
        )

        if recall_resp.status_code not in (200, 201):
            print(f"[JOIN] Recall failed: {recall_resp.status_code} {recall_resp.text}")
            return {"error": "Recall bot creation failed", "detail": recall_resp.text}, 500

        recall_bot_id = recall_resp.json().get("id", "unknown")
        print(f"[JOIN] Recall bot created: {recall_bot_id}")

        # Step 2 — Create LiveKit room + dispatch agent
        room_metadata = json.dumps({
            "openclawUrl":  openclaw_url,
            "gatewayToken": gateway_token,
            "sessionKey":   session_key,
            "avatarId":     avatar_id,
            "meetingUrl":   meeting_url,
        })

        async def setup_livekit():
            client = lk_api.LiveKitAPI(
                url=os.getenv("LIVEKIT_URL"),
                api_key=os.getenv("LIVEKIT_API_KEY"),
                api_secret=os.getenv("LIVEKIT_API_SECRET"),
            )
            await client.room.create_room(
                lk_api.CreateRoomRequest(
                    name=room_id,
                    metadata=room_metadata,
                )
            )
            print(f"[JOIN] LiveKit room created: {room_id}")

            await client.agent.create_job(
                lk_api.CreateJobRequest(
                    agent_name="my_agent",
                    room=lk_api.JobRoom(name=room_id),
                )
            )
            print(f"[JOIN] Agent dispatched: {room_id}")
            await client.aclose()

        asyncio.run(setup_livekit())

        return {
            "status": "ok",
            "roomId": room_id,
            "recallBotId": recall_bot_id,
            "avatarId": avatar_id,
            "meetingUrl": meeting_url
        }, 200

    except Exception as e:
        print(f"[JOIN] Error: {e}")
        traceback.print_exc()
        return {"error": str(e)}, 500

def run_proxy():
    print("--- OpenClaw Proxy Active (port 8080) ---")
    app.run(host='0.0.0.0', port=8080, debug=False, use_reloader=False)

threading.Thread(target=run_proxy, daemon=True).start()
 
# --- LIVEKIT AGENT ---
AGENT_INSTRUCTIONS = (
    "You are a helpful AI assistant. Keep responses to 2-4 short spoken sentences. "
    "Be conversational. Never use markdown, bullet points, or formatting."
)
 
class MyAgent(Agent):
    def __init__(self) -> None:
        super().__init__(instructions=AGENT_INSTRUCTIONS)
 
server = AgentServer()
 
@server.rtc_session()
async def my_agent(ctx: agents.JobContext):
    await ctx.connect()
   
    # 1. Get Config (Strictly from metadata passed from frontend localStorage)
    config = None
    for p in ctx.room.remote_participants.values():
        if p.metadata:
            try:
                config = json.loads(p.metadata)
                print(f"[SESSION] ✓ Config from participant metadata")
                break
            except: pass
   
    if not config:
        print(f"[SESSION] ✗ No config found in participant metadata")
        
    if not config and ctx.room.metadata:
        try:
            config = json.loads(ctx.room.metadata)
            print(f"[SESSION] ✓ Config from room metadata")
        except:
            pass
            
    if not config:
        print(f"[SESSION] ✗ No config found in participant or room metadata")
        return
 
    url    = config.get("openclawUrl", "")
    token  = config.get("gatewayToken", "")
    key    = config.get("sessionKey", "")
 
    if not url or not token or not key:
        print(f"[SESSION] ✗ Incomplete config: {config}")
        return
 
    # 2. Determine Voice ID based on Avatar Gender
    # Male: Kevin, Jason, Sameer, Mike, Johnny, Aman, Alex, Amir, Akbar
    # Female: Jessica, Cathy, Sofia, Lucy, Kiara, Jennifer, Priya, Chloe, Lisa, Allie, Misha
    avatar_id = config.get("avatarId") or os.getenv("TRUGEN_AVATAR_ID")
    if not avatar_id:
        avatar_id = "1a640442" # Default to Lisa

    male_ids = {
        "182b03e8", "05a001fc", "be5b2ce0", "03ae0187", 
        "1fa504ff", "0f160301", "13550375", "48d778c9", "18c4043e"
    }
    
    # Female: aura-2-andromeda-en, Male: aura-2-aries-en
    voice_id = "aura-2-aries-en" if avatar_id in male_ids else "aura-2-andromeda-en"
    print(f"[SESSION] Using Avatar ID: {avatar_id}, Voice ID: {voice_id}")

    # 3. MEGA-TOKEN: Pack everything into the api_key for the proxy
    mega_token = f"{url}|{token}|{key}"

    openclaw_llm = openai.LLM(
        model="main",
        base_url="http://localhost:8080/v1",
        api_key=mega_token,
    )

    # 4. Simple AgentSession setup
    session = AgentSession(
        stt=groq.STT(model="whisper-large-v3-turbo"),
        vad=silero.VAD.load(),
        llm=openclaw_llm,
        tts=deepgram.TTS(
            model=voice_id,
        ),
    )
   
    trugen_avatar = trugen.AvatarSession(avatar_id=avatar_id)
    await trugen_avatar.start(session, room=ctx.room)
 
    await session.start(room=ctx.room, agent=MyAgent())
    session.say("Hello! I'm ready to chat.")
 
if __name__ == "__main__":
    import sys
    if len(sys.argv) > 1 and sys.argv[1] == "download-files":
        # This is used by the Dockerfile to pre-download models (e.g. Silero)
        print("Pre-downloading models...")
        silero.VAD.load()
        sys.exit(0)
    agents.cli.run_app(server)