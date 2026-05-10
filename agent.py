import os
import json
import asyncio
import base64
import requests
import typing
import logging
import boto3
import datetime
from dotenv import load_dotenv
from livekit import agents
from livekit.agents import Agent, AgentServer, AgentSession, cli, metrics, APIConnectOptions, StopResponse, llm
from livekit.agents.telemetry import set_tracer_provider
from livekit.agents.voice import MetricsCollectedEvent
from livekit.agents.voice.agent_session import SessionConnectOptions
from livekit.plugins import elevenlabs, openai, trugen, silero, deepgram

# OTEL for Langfuse
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.exporter.otlp.proto.http.trace_exporter import OTLPSpanExporter
from opentelemetry.sdk.trace.export import BatchSpanProcessor

import time
import random
import websockets
from livekit.agents import stt
from livekit.agents.utils import AudioBuffer
from livekit.agents.types import NOT_GIVEN, NotGivenOr
from livekit.agents.voice.room_io import RoomOptions

load_dotenv()
logger = logging.getLogger("trugen-agent")

# ---------------------------------------------------------------------------
# LOG FILTERING
# ---------------------------------------------------------------------------
class SilencePaddingFilter(logging.Filter):
    """
    Suppresses noisy LiveKit warnings (silence padding, input samples) for all sessions.
    """
    def filter(self, record):
        msg = record.getMessage()
        if "skipping silence padding" in msg or "Input is shorter by" in msg:
            return False
        return True

# Apply filter to the noisy LiveKit logger
logging.getLogger("livekit.agents").addFilter(SilencePaddingFilter())


# ---------------------------------------------------------------------------
# CONFIG RESOLUTION
# ---------------------------------------------------------------------------
EMAIL_BOT_AVATAR_MAP = {
    "amansbot": "0f160301", "jasonbot": "182b03e8", "sameerbot": "05a001fc",
    "mikebot": "be5b2ce0", "johnnybot": "03ae0187", "amanbot": "0f160301",
    "alexbot": "13550375", "amirbot": "48d778c9",
    "jessicabot": "1a640442", "lisasbot": "1a640442",
    "lisabot": "1a640442", "cathybot": "1a640442", "sofiabot": "1a640442",
    "lucybot": "1a640442", "kiarabot": "1a640442", "jenniferbot": "1a640442",
    "priyabot": "1a640442", "chloebot": "1a640442", "mishabot": "1a640442",
    "alliebot": "1a640442"
}

MALE_AVATAR_IDS = {
    "182b03e8", "05a001fc", "be5b2ce0", "03ae0187",
    "1fa504ff", "0f160301", "13550375", "18c4043e", "48d778c9",
    "60a0926a", "5daa73d5", "2b130585"
}


# ---------------------------------------------------------------------------
# BACKEND USAGE REPORTING
# ---------------------------------------------------------------------------
async def post_backend_usage(usage: dict):
    """Async POST session usage summary to BACKEND_BASE_URL.
    Uses aiohttp with no timeout — fully dynamic, API responds however long it needs.
    Runs as a detached background task so it never blocks LiveKit shutdown.
    """
    base_url = os.getenv("BACKEND_BASE_URL", "").rstrip("/")
    if not base_url:
        logger.warning("[USAGE] BACKEND_BASE_URL not set, skipping backend usage post.")
        return
    endpoint = f"{base_url}/v1/usage"
    import aiohttp
    try:
        # total=None disables all timeouts — fully dynamic, no ConnectTimeout ever.
        async with aiohttp.ClientSession(timeout=aiohttp.ClientTimeout(total=None)) as session:
            async with session.post(endpoint, json=usage) as resp:
                resp.raise_for_status()
                logger.info(f"✅ [USAGE] Backend usage posted successfully for {usage.get('conversation_id')} → {endpoint} | {resp.status}")
    except Exception as e:
        logger.error(f"❌ [USAGE] Error posting usage for {usage.get('conversation_id')}: {e}")


async def register_conversation(config: dict, conversation_id: str):
    """Registers a new conversation with the backend."""
    base_url = os.getenv("FRONTEND_URL", "").rstrip("/")
    if not base_url:
        logger.warning("[REGISTER] FRONTEND_URL not set, cannot register conversation.")
        return
    
    endpoint = f"{base_url}/api/conversations/register"
    payload = {
        "conversation_id": conversation_id,
        "user_email": config.get("user_email") or config.get("userEmail") or config.get("ownerEmail") or config.get("email") or "system@clawdface.ai",
        "bot_name": config.get("name") or "AI Assistant",
        "bot_avatar": config.get("avatarId") or "0f160301",
        "agent_id": config.get("agentId") or ""
    }
    
    logger.debug(f"[REGISTER] Sending payload: {json.dumps(payload)}")
    
    import aiohttp
    try:
        async with aiohttp.ClientSession() as session:
            async with session.post(endpoint, json=payload, timeout=aiohttp.ClientTimeout(total=10)) as resp:
                if resp.status == 200:
                    logger.info(f"✅ [REGISTER] Session registered successfully | ID: {conversation_id}")
                else:
                    text = await resp.text()
                    logger.error(f"❌ [REGISTER] Failed to register ID: {conversation_id} | Status: {resp.status} | Error: {text}")
    except Exception as e:
        logger.error(f"[REGISTER] Error during registration: {e}")


# ---------------------------------------------------------------------------
# TRANSCRIPT S3 UPLOAD
# ---------------------------------------------------------------------------
def push_transcript_to_s3(transcript_data: list, conversation_id: str):
    """Upload conversation transcript to S3 in the backend-compatible format."""
    try:
        # Support both legacy and backend bucket env names
        bucket_name = os.getenv("BUCKET_NAME") or os.getenv("AWS_BUCKET_ADDITIONAL") or os.getenv("AWS_BUCKET")
        aws_access_key_id = os.getenv("AWS_ACCESS_KEY_ID")
        aws_secret_access_key = os.getenv("AWS_SECRET_ACCESS_KEY")
        region = os.getenv("REGION") or os.getenv("AWS_REGION") or "us-east-2"

        if not bucket_name:
            logger.error("[TRANSCRIPT] BUCKET_NAME environment variable is not set, cannot upload transcript")
            return

        if not aws_access_key_id or not aws_secret_access_key:
            logger.error("[TRANSCRIPT] AWS credentials not set, cannot upload transcript")
            return

        # Transform transcript to backend-compatible format
        # Backend expects: { items: [{ id, type, role, content[], metrics, ... }] }
        items = []
        for idx, entry in enumerate(transcript_data):
            content = entry.get("content", "")
            if isinstance(content, list):
                content_list = [str(c) for c in content]
            elif content is None:
                content_list = []
            else:
                content_list = [str(content)]

            message_ts = entry.get("message_timestamp")
            metrics = {"started_speaking_at": message_ts} if message_ts else {}

            items.append({
                "id": entry.get("id") or f"{conversation_id}-{message_ts or idx}-{idx}",
                "type": entry.get("type", "message"),
                "role": entry.get("role", "assistant"),
                "content": content_list,
                "metrics": metrics,
                "name": entry.get("name", ""),
                "arguments": entry.get("arguments", ""),
                "output": entry.get("output", ""),
                "is_error": bool(entry.get("is_error", False)),
                "extra": entry.get("extra", None),
                "call_id": entry.get("call_id", ""),
            })

        # S3 path matching backend expectation: egress/{conversation_id}/transcript.json
        s3_key = f"egress/{conversation_id}/transcript.json"

        # Create S3 client
        s3_client = boto3.client(
            "s3",
            aws_access_key_id=aws_access_key_id,
            aws_secret_access_key=aws_secret_access_key,
            region_name=region,
        )

        # Convert transcript to JSON
        transcript_json = json.dumps({"items": items}, indent=2)

        # Upload to S3
        s3_client.put_object(
            Bucket=bucket_name,
            Key=s3_key,
            Body=transcript_json.encode("utf-8"),
            ContentType="application/json",
        )

        logger.info(f"[TRANSCRIPT] ✓ Uploaded to S3: s3://{bucket_name}/{s3_key} ({len(items)} messages)")

    except Exception as e:
        logger.error(f"[TRANSCRIPT] ✗ Failed to upload transcript to S3: {e}", exc_info=True)


# ---------------------------------------------------------------------------
# RECALL.AI STT — connects to relay with room_id in URL
# ---------------------------------------------------------------------------
class RecallAIDirectSTT(stt.STT):
    def __init__(self, ctx: agents.JobContext, recall_bot_id: str = "", room_id: str = ""):
        super().__init__(capabilities=stt.STTCapabilities(streaming=True, interim_results=True))
        self._ctx = ctx
        self._recall_bot_id = recall_bot_id
        self._room_id = room_id  # ← LiveKit room name, used in relay URL

    @property
    def provider(self) -> str: return "recall-ai-direct"

    async def _recognize_impl(self, buffer: AudioBuffer, *, language: NotGivenOr[str] = NOT_GIVEN, conn_options: APIConnectOptions = APIConnectOptions()) -> stt.SpeechEvent:
        return stt.SpeechEvent(type=stt.SpeechEventType.START_OF_SPEECH, alternatives=[])

    def stream(self, *, language: NotGivenOr[str] = NOT_GIVEN, conn_options: APIConnectOptions = APIConnectOptions()) -> "RecallSpeechStream":
        logger.info(f"[STT] Creating new RecallSpeechStream... bot_id={self._recall_bot_id}")
        return RecallSpeechStream(
            stt=self,
            conn_options=conn_options,
            ctx=self._ctx,
            recall_bot_id=self._recall_bot_id,
            room_id=self._room_id,
        )


class RecallSpeechStream(stt.SpeechStream):
    def __init__(self, *, stt: RecallAIDirectSTT, conn_options: APIConnectOptions,
                 ctx: agents.JobContext, recall_bot_id: str = "", room_id: str = "") -> None:
        super().__init__(stt=stt, conn_options=conn_options)
        self._ctx = ctx
        self._recall_bot_id = recall_bot_id
        self._room_id = room_id
        self._speaking = False

    def _emit_final(self, text: str):
        if not text: return
        if not self._speaking:
            self._event_ch.send_nowait(stt.SpeechEvent(type=stt.SpeechEventType.START_OF_SPEECH, alternatives=[]))
            self._speaking = True
        self._event_ch.send_nowait(stt.SpeechEvent(type=stt.SpeechEventType.END_OF_SPEECH, alternatives=[stt.SpeechData(text=text, language="en")]))  # type: ignore
        self._event_ch.send_nowait(stt.SpeechEvent(type=stt.SpeechEventType.FINAL_TRANSCRIPT, alternatives=[stt.SpeechData(text=text, language="en")]))  # type: ignore
        self._speaking = False

    def _emit_interim(self, text: str):
        if not text: return
        if not self._speaking:
            self._event_ch.send_nowait(stt.SpeechEvent(type=stt.SpeechEventType.START_OF_SPEECH, alternatives=[]))
            self._speaking = True
        self._event_ch.send_nowait(stt.SpeechEvent(type=stt.SpeechEventType.INTERIM_TRANSCRIPT, alternatives=[stt.SpeechData(text=text, language="en")]))  # type: ignore

    async def _run(self) -> None:
        base_url = os.getenv("EXTERNAL_MEETINGS_WS_URL", "").strip()
        if not base_url:
            logger.warning("[RECALL] EXTERNAL_MEETINGS_WS_URL is not set!")
        retry_delay = 2

        # -----------------------------------------------------------------------
        # RELAY ROUTING: Include room_id (and bot_id) as query params so the relay
        # can match this agent WS connection to incoming Recall.ai events.
        # Recall.ai connects to the relay with ?room_id=... in its endpoint URL;
        # the relay routes by matching the same param on the agent side.
        # We also send message-based registration as a secondary/fallback protocol.
        # -----------------------------------------------------------------------
        room_id = self._room_id or self._ctx.room.name
        logger.info(f"[RECALL] Base URL: {base_url}")

        while True:
            try:
                logger.info(f"[RECALL] Initializing WebSocket connection to {base_url}...")
                async with websockets.connect(
                    base_url,
                    ping_interval=20,
                    ping_timeout=20,
                    open_timeout=15,
                ) as ws:
                    logger.info("[RECALL] Handshake successful. Sending registration messages...")

                    # 1. Register Room ID
                    reg_room = {"type": "set_lk_room_id", "data": room_id}
                    logger.info(f"[RECALL] Sending Room Registration → {reg_room}")
                    await ws.send(json.dumps(reg_room))

                    # 2. Register Bot ID (optional)
                    if self._recall_bot_id:
                        reg_bot = {"type": "set_bot_id", "data": self._recall_bot_id}
                        logger.info(f"[RECALL] Sending Bot Registration → {reg_bot}")
                        await ws.send(json.dumps(reg_bot))
                    else:
                        logger.warning("[RECALL] No bot_id provided for registration. Routing might depend solely on room_id.")

                    logger.info(f"[RECALL] ✓ Handshake and registration successful. Room: {room_id}")
                    retry_delay = 2

                    while True:
                        try:
                            raw = await asyncio.wait_for(ws.recv(), timeout=30.0)
                        except asyncio.TimeoutError:
                            logger.debug("[RECALL] keepalive ping")
                            await ws.ping()
                            continue

                        msg = json.loads(raw)
                        event = msg.get("event")
                        logger.debug(f"[RECALL] Incoming event: {event} | Payload: {raw}")

                        if event in ("transcript.data", "transcript.partial_data"):
                            # Robust parsing: try nested 'data.data' then fall back to 'data'
                            inner_data = msg.get("data", {})
                            if isinstance(inner_data, dict) and "data" in inner_data:
                                inner_data = inner_data.get("data", {})

                            words = inner_data.get("words", []) if isinstance(inner_data, dict) else []
                            text = " ".join(
                                w.get("text", "") for w in words
                                if isinstance(w, dict) and w.get("text")
                            ).strip()

                            if text:
                                if event == "transcript.data":
                                    participant = inner_data.get("participant", {})
                                    speaker = participant.get("name", "Unknown") if isinstance(participant, dict) else "Unknown"
                                    logger.debug(f"[RECALL] FINAL | {speaker}: {text}")
                                    self._emit_final(f"{speaker}: {text}")
                                else:
                                    logger.debug(f"[RECALL] PARTIAL: {text}")
                                    self._emit_interim(text)

                        elif event == "participant_events.join":
                            participant = msg.get("data", {}).get("data", {}).get("participant", {})
                            name = participant.get("name", "Unknown") if isinstance(participant, dict) else "Unknown"
                            logger.info(f"[RECALL] ✓ Participant joined: {name}")

                        elif event == "participant_events.leave":
                            participant = msg.get("data", {}).get("data", {}).get("participant", {})
                            name = participant.get("name", "Unknown") if isinstance(participant, dict) else "Unknown"
                            logger.info(f"[RECALL] Participant left: {name}")

                        else:
                            # Log ANY unrecognised event so we can debug relay message format
                            logger.info(f"[RECALL] Unknown event type: {event} | raw: {raw[:200]}")

            except websockets.InvalidHandshake as e:
                logger.error(f"[RECALL] ✗ Protocol error during handshake: {e}. Check if the URL is a valid WS/WSS endpoint. Retrying...")
                await asyncio.sleep(retry_delay)
                retry_delay = min(retry_delay * 2, 30)
            except websockets.ConnectionClosed as e:
                logger.warning(f"[RECALL] ! Connection closed unexpectedly: {e} (code={e.code}). Attempting reconnect in {retry_delay}s...")
                await asyncio.sleep(retry_delay)
                retry_delay = min(retry_delay * 2, 30)
            except Exception as e:
                logger.error(f"[RECALL] ✗ Unexpected error in relay loop: {e}. Full context follows:", exc_info=True)
                await asyncio.sleep(retry_delay)
                retry_delay = min(retry_delay * 2, 30)


def resolve_config(ctx: agents.JobContext) -> tuple[dict, str]:
    # 1. Job Metadata (Highest Priority for Dispatches)
    try:
        if ctx.job and ctx.job.metadata:
            cfg = json.loads(ctx.job.metadata)
            if cfg.get("recallBotId"):
                return cfg, "recall"
            if cfg.get("openclawUrl"):
                explicit_type = cfg.get("connection_type", "")
                return cfg, explicit_type if explicit_type else "email_dispatch"
    except Exception as e:
        logger.debug(f"[CONFIG] Failed to parse job metadata: {e}")

    # 2. Room Metadata
    try:
        if ctx.room.metadata:
            cfg = json.loads(ctx.room.metadata)
            if cfg.get("recallBotId"):
                 return cfg, "recall"
            if cfg.get("openclawUrl"):
                return cfg, "room_metadata"
    except Exception as e:
        logger.debug(f"[CONFIG] Failed to parse room metadata: {e}")

    # 3. Participant Metadata
    for p in ctx.room.remote_participants.values():
        try:
            if p.metadata:
                cfg = json.loads(p.metadata)
                if cfg.get("openclawUrl"):
                    type = "url_share" if str(cfg.get("sessionKey", "")).startswith("session-") else "website"
                    if cfg.get("recallBotId"): type = "recall"
                    return cfg, type
        except: pass

    # 4. Backend Dynamic Lookup (by Room Name/Email)
    room_id = ctx.room.name
    if room_id and not room_id.startswith("room-"):
        email = room_id if "@" in room_id else f"{room_id}@agent.clawdface.ai"
        try:
            base_url = os.getenv("FRONTEND_URL", "").rstrip("/")
            if base_url:
                logger.info(f"[CONFIG] Fetching sync config for {email}...")
                resp = requests.get(f"{base_url}/api/agents/config?email={email}", timeout=5)
                if resp.status_code == 200:
                    cfg = resp.json()
                    if cfg.get("openclawUrl"):
                        mode = "recall" if cfg.get("recallBotId") else "email_dispatch"
                        return cfg, mode
        except Exception as e:
            logger.debug(f"[CONFIG] Lookup error for {email}: {e}")

    return {}, "unknown"


def setup_langfuse(metadata: dict):
    pub = os.getenv("LANGFUSE_PUBLIC_KEY")
    sec = os.getenv("LANGFUSE_SECRET_KEY")
    host = os.getenv("LANGFUSE_HOST") or os.getenv("LANGFUSE_BASE_URL")

    if not all([pub, sec, host]):
        print("[LANGFUSE] Missing credentials, tracing disabled.")
        return None

    auth = base64.b64encode(f"{pub}:{sec}".encode()).decode()
    safe_host = str(host).rstrip("/")
    os.environ["OTEL_EXPORTER_OTLP_ENDPOINT"] = f"{safe_host}/api/public/otel"
    os.environ["OTEL_EXPORTER_OTLP_HEADERS"] = f"Authorization=Basic {auth}"

    tp = TracerProvider()
    tp.add_span_processor(BatchSpanProcessor(OTLPSpanExporter()))
    set_tracer_provider(tp, metadata=metadata)
    return tp

class MyAgent(Agent):
    def __init__(self, *, groq_llm: llm.LLM | None = None, enable_thinking: bool = True, thinking_delay: float = 3.0, conversation_id: str = "", **kwargs) -> None:
        super().__init__(
        instructions=(
            "You are a helpful AI assistant. Keep responses to 2-4 short spoken sentences. "
            "Be conversational. Never use markdown, bullet points, or formatting."
        ),
            **kwargs
        )
        self._groq_llm = groq_llm  # Groq for instant waiting messages
        self._enable_thinking = enable_thinking
        self._thinking_delay = thinking_delay
        # Prevent concurrent turn processing
        self._turn_lock = asyncio.Lock()
        # Track if thinking played for current turn
        self._thinking_played = False
        self._response_buffer = []
        self._last_user_message = None  # Store for context-aware waiting
        self._shutting_down = False # Guard against post-disconnect transients
        
        # ── TRANSCRIPT COLLECTION ──────────────────────────────────────────
        self._conversation_id = conversation_id
        self._transcript = []  # List of transcript entries
        self._transcript_lock = asyncio.Lock()  # Thread-safe transcript updates
    
    async def on_user_turn_completed(self, turn_ctx, new_message):
        """Custom handling: stream LLM, buffer if thinking needed, then release."""
        if self._shutting_down:
            logger.info("[STREAM] Agent is shutting down, ignoring user turn")
            return

        if not self._enable_thinking or not self._llm:
            return  # Normal flow if thinking disabled
            
        # Prevent overlapping turns
        if self._turn_lock.locked():
            logger.info("[STREAM] Previous turn still active, skipping")
            return
            
        async with self._turn_lock:
            await self._process_turn(turn_ctx, new_message)
        
        # Raise StopResponse to prevent base Agent class from generating duplicate response
        raise StopResponse()
    
    async def _process_turn(self, turn_ctx, new_message):
        """Process a single user turn with streaming LLM."""
        # Reset state for new turn
        self._thinking_played = False
        self._response_buffer = []
        
        # Store user message for context-aware waiting
        self._last_user_message = new_message
        
        # ── CAPTURE USER MESSAGE IN TRANSCRIPT ─────────────────────────────
        await self._add_to_transcript(
            role="user",
            content=new_message.content if hasattr(new_message, 'content') else str(new_message),
            message_type="message"
        )
        
        # Build chat context
        chat_ctx = turn_ctx.copy()
        chat_ctx.items.append(new_message)
        
        # Create events for coordination
        thinking_started = asyncio.Event()
        thinking_done = asyncio.Event()
        first_chunk_received = asyncio.Event()
        
        # Start LLM stream
        logger.info("[STREAM] Starting LLM stream...")
        from livekit.agents import llm as llm_lib
        llm_provider = self.llm
        if not isinstance(llm_provider, llm_lib.LLM):
             logger.error(f"[STREAM] LLM provider not supported or not found: {type(llm_provider)}")
             return
             
        llm_stream = llm_provider.chat(chat_ctx=chat_ctx)
        stream_start_time = asyncio.get_event_loop().time()
        
        # Start thinking timer
        thinking_task = asyncio.create_task(
            self._thinking_timer(thinking_started, thinking_done)
        )
        
        # Process LLM stream
        chunk_count = 0
        
        try:
            async for chunk in llm_stream:
                chunk_count += 1
                
                # Extract text
                text = self._extract_text_from_chunk(chunk)
                if text:
                    self._response_buffer.append(text)
                    
                    if not first_chunk_received.is_set():
                        first_chunk_received.set()
                        elapsed = asyncio.get_event_loop().time() - stream_start_time
                        logger.info(f"[STREAM] First text chunk after {elapsed:.2f}s")
                        
                        if not thinking_started.is_set():
                            logger.info("[STREAM] Fast response, cancelling thinking timer")
                            thinking_task.cancel()
                            try:
                                await thinking_task
                            except asyncio.CancelledError:
                                pass
                            thinking_done.set()
                        else:
                            logger.info("[STREAM] Response arrived during thinking - will wait for it to finish")
                    
        except Exception as e:
            logger.error(f"[STREAM] LLM stream error: {e}")
            thinking_task.cancel()
            return
        
        # Wait for thinking to complete (always wait for full phrase)
        try:
            await thinking_done.wait()
        except Exception:
            pass  # Ignore if cancelled
        
        # Speak main response (with error handling for session closing)
        full_text = "".join(self._response_buffer)
        if full_text.strip():
            logger.info(f"[STREAM] Speaking {len(full_text)} chars ({chunk_count} chunks)")
            try:
                await self.session.say(full_text)
                
                # ── CAPTURE AGENT RESPONSE IN TRANSCRIPT ───────────────────
                await self._add_to_transcript(
                    role="assistant",
                    content=full_text,
                    message_type="message"
                )
            except RuntimeError as e:
                if "AgentSession is closing" in str(e):
                    logger.info("[STREAM] Session closing, skipping speak")
                else:
                    raise
        else:
            logger.warning("[STREAM] Empty response, nothing to speak")
    
    async def _on_shutdown(self) -> None:
        """Cleanup session resources."""
        self._shutting_down = True
        logger.info("[SESSION] Shutting down agent")
    
    async def _add_to_transcript(self, role: str, content: str, message_type: str = "message"):
        """Add a message to the transcript in a thread-safe manner."""
        async with self._transcript_lock:
            timestamp_unix = int(time.time())
            timestamp_iso = datetime.datetime.now(datetime.timezone.utc).isoformat()
            
            entry = {
                "timestamp": timestamp_iso,
                "role": role,
                "content": content,
                "type": message_type,
                "message_timestamp": timestamp_unix,
                "conversation_id": self._conversation_id,
            }
            
            self._transcript.append(entry)
            logger.debug(f"[TRANSCRIPT] Added {role} message: {content[:50]}...")
    
    def get_transcript(self) -> list[dict]:
        """Return the collected transcript."""
        return self._transcript.copy()
    
    def _extract_text_from_chunk(self, chunk):
        """Extract text from LLM chunk."""
        text = ""
        
        # Try OpenAI-style chunks
        if hasattr(chunk, 'choices') and chunk.choices:
            delta = chunk.choices[0].delta
            if hasattr(delta, 'content') and delta.content:
                text = delta.content
        # Try LiveKit ChatChunk style
        elif hasattr(chunk, 'delta') and chunk.delta:
            if hasattr(chunk.delta, 'content') and chunk.delta.content:
                text = chunk.delta.content
        # Try direct string
        elif isinstance(chunk, str):
            text = chunk
            
        return text
    
    async def _thinking_timer(self, started_event, done_event):
        """Timer that fires after delay to generate dynamic waiting message via Groq."""
        try:
            await asyncio.sleep(self._thinking_delay)
            
            started_event.set()
            
            # Generate dynamic waiting message using Groq
            waiting_message = await self._generate_waiting_message()
            
            logger.info(f"[THINKING] Timer fired, saying: '{waiting_message}'")
            self._thinking_played = True
            try:
                await self.session.say(waiting_message, allow_interruptions=False)
                # ── Capture filler message in transcript ──
                await self._add_to_transcript(
                    role="assistant",
                    content=waiting_message,
                    message_type="message"
                )
                logger.info("[THINKING] Done")
            except RuntimeError as e:
                if "AgentSession is closing" in str(e):
                    logger.info("[THINKING] Session closing, skipping say()")
                else:
                    raise
            
        except asyncio.CancelledError:
            logger.info("[THINKING] Timer cancelled (fast response or interrupted)")
        finally:
            done_event.set()
    
    async def _generate_waiting_message(self) -> str:
        """Generate context-aware waiting message using Groq LLM."""
        if not self._groq_llm or not self._last_user_message:
            fallbacks = [
                "Let me think about that for a moment.",
                "Hmm, give me a second to consider this.",
                "Just a moment while I work through that.",
            ]
            return random.choice(fallbacks)
        
        try:
            user_query = self._last_user_message.content or ""
            
            waiting_prompt = (
                "Generate a brief context-aware filler phrase while the main response is being prepared.\n\n"
                f"User message: '{user_query}'\n\n"
                "RULES:\n"
                "1. First, identify the TRUE intent behind the message — not the literal words\n"
                "2. 3-8 words, always ends with '...'\n"
                "3. NEVER extract filler words as topics — words like 'fine', 'okay', 'yeah', 'good', "
                "'all', 'now', 'here', 'it', 'things' are NOT topics\n"
                "4. NEVER say 'that' — name the actual subject\n\n"
                "INTENT DETECTION LOGIC — pick the right pattern:\n"
                "- Message expresses a STATUS or FEELING (fine, okay, good, not great, tired, happy) "
                "→ Acknowledge the emotion, bridge to response: 'Glad to hear, putting a reply together...'\n"
                "- Message is a GREETING (hi, hello, hey, good morning) "
                "→ Warm setup phrase: 'Getting things ready for you...'\n"
                "- Message is APPRECIATION or COMPLIMENT (thanks, you're great, awesome) "
                "→ Light acknowledgment: 'Happy to help, thinking ahead...'\n"
                "- Message is AGREEMENT or ACKNOWLEDGMENT (okay cool, makes sense, got it, sure) "
                "→ Move forward naturally: 'Good, figuring out the next step...'\n"
                "- Message is UNCERTAINTY or NEGATIVE (not sure, not really, I don't know, not good) "
                "→ Supportive bridge: 'No worries, thinking it through...'\n"
                "- Message is a clear TASK or QUESTION (what is X, help me with Y, fix Z) "
                "→ Reference the specific topic: 'Looking into [topic] for you...'\n\n"
                "GOOD examples:\n"
                "   - 'It is all okay now here' → 'Glad things are sorted, thinking ahead...'\n"
                "   - 'Everything is fine now' → 'Good to know, putting a reply together...'\n"
                "   - 'Yeah all good' → 'Nice, working on a response...'\n"
                "   - 'Not really doing great' → 'No worries, thinking it through...'\n"
                "   - 'Haha yeah makes sense' → 'Glad it clicked, figuring out more...'\n"
                "   - 'Hi there' → 'Setting things up for you...'\n"
                "   - 'Thanks a lot!' → 'Happy to help, thinking ahead...'\n"
                "   - 'Fix my Python code' → 'Scanning your Python code...'\n"
                "   - 'Plan a trip to Rome' → 'Planning your Rome itinerary...'\n"
                "   - 'What's the weather?' → 'Checking the weather forecast...'\n\n"
                "BAD examples (never do this):\n"
                "   - 'Looking into your fine status...' ❌\n"
                "   - 'Processing your okay...' ❌\n"
                "   - 'Checking your yeah...' ❌\n"
                "   - 'Thinking about your here...' ❌\n"
                "   - 'One moment please' ❌\n"
                "   - 'Let me think about that' ❌\n\n"
                "Respond with ONLY the filler phrase. No labels, no explanation."
            )
            
            from livekit.agents import llm as llm_types
            chat_ctx = llm_types.ChatContext()
            chat_ctx.add_message(role="user", content=waiting_prompt)
            
            response_text = ""
            async for chunk in self._groq_llm.chat(chat_ctx=chat_ctx):
                text = self._extract_text_from_chunk(chunk)
                if text:
                    response_text += text
            
            waiting_msg = response_text.strip().strip('"').strip("'")
            
            if len(waiting_msg) > 80:
                waiting_msg = waiting_msg[:77] + "..."
            
            if waiting_msg:
                logger.info(f"[GROQ] Generated waiting message: '{waiting_msg}'")
                return waiting_msg
            else:
                raise ValueError("Empty Groq response")
                
        except Exception as e:
            logger.warning(f"[GROQ] Failed to generate waiting message: {e}, using fallback")
            fallbacks = [
                "Let me think about that...",
                "Hmm, give me a moment...",
                "One second while I check...",
            ]
            return random.choice(fallbacks)


server = AgentServer()

@server.rtc_session(agent_name="clawdface")
async def my_agent(ctx: agents.JobContext):
    await ctx.connect()

    # ── IDs: resolve once here, reuse everywhere ──────────────────────────
    _raw_job_meta: dict = {}
    try:
        if ctx.job and ctx.job.metadata:
            _raw_job_meta = json.loads(ctx.job.metadata)
    except Exception:
        pass

    # ─────────────────────────────────────────────────────────────────────
    # Generate and Register conversation session
    # ─────────────────────────────────────────────────────────────────────
    config, connection_type = {}, "unknown"
    for _ in range(5):
        config, connection_type = resolve_config(ctx)
        if config: break
        await asyncio.sleep(0.5)

    if not config:
        logger.error(f"[SESSION] ✗ Failed to resolve config for room {ctx.room.name}")
        return

    # Resolve Conversation ID from dispatched metadata (always provided by route.ts)
    job_id = ctx.job.id
    conversation_id = (
        config.get("conversation_id")
        or config.get("conversationId")
        or ""
    )

    if not conversation_id:
        logger.warning(
            f"[SESSION] ⚠️  No conversation_id in metadata for job {job_id} — "
            "session usage/transcripts will not be linked. This is likely a stray or headless session."
        )
        return

    logger.info(f"🔗 [SESSION] conversation_id={conversation_id} | job_id={job_id}")

    url = config.get("openclawUrl", "").strip()
    token = config.get("gatewayToken", "")
    key = config.get("sessionKey", "")
    avatar_id = config.get("avatarId") or os.getenv("TRUGEN_AVATAR_ID")
    voice_id = "CwhRBWXzGAHq8TQ4Fs17" if avatar_id in MALE_AVATAR_IDS else "FGY2WhTYpPnrIDTdsKH5"
    
    # Resolve dynamic thinking settings — check both snake_case and camelCase
    thinking_enabled_val = config.get("enable_thinking") or config.get("thinking_enabled") or "true"
    ENABLE_LET_ME_THINK = str(thinking_enabled_val).lower() == "true"
    
    thinking_delay_val = config.get("thinking_delay") or config.get("thinkingDelay") or "5.0"
    try:
        THINKING_DELAY = float(thinking_delay_val)
    except (ValueError, TypeError):
        THINKING_DELAY = 5.0
    
    logger.info(f"[CONFIG] Thinking: {ENABLE_LET_ME_THINK} | Delay: {THINKING_DELAY}s")

    tp = setup_langfuse({
        "langfuse.session.id": ctx.room.name,
        "langfuse.user.id": key,
        "langfuse.tags": json.dumps(["production", f"source:{connection_type}", f"avatar:{avatar_id}"]),
        "connection.type": connection_type,
        "avatar.id": avatar_id,
        "room.name": ctx.room.name
    })

    if tp:
        async def _flush_sync() -> None:
            tp.force_flush()
        ctx.add_shutdown_callback(_flush_sync)

    import openai as _openai
    llm = openai.LLM(
        model="openclaw",
        client=_openai.AsyncOpenAI(
            base_url=f"{url}/v1",
            api_key=token,
            default_headers={
                "x-openclaw-session-key": key,
                "x-openclaw-agent-id": "main",
                "ngrok-skip-browser-warning": "true"
            },
            timeout=None,
            max_retries=0
        )
    )

    vad_provider = silero.VAD.load()
    if connection_type in ("email_dispatch", "recall") or config.get("recallBotId"):
        recall_bot_id = config.get("recallBotId", "") or ""
        livekit_room_id = config.get("roomId") or ctx.room.name
        logger.info(f"[SESSION] Start: {connection_type} | Mode: RECALL | Bot: {recall_bot_id or 'none'}")
        logger.info(f"[STT] Meeting mode → RecallAIDirectSTT | room_sid={livekit_room_id}")
        stt_provider = RecallAIDirectSTT(ctx=ctx, recall_bot_id=recall_bot_id, room_id=livekit_room_id)
    else:
        logger.info(f"[SESSION] Start: {connection_type} | Mode: STANDARD")
        logger.info(f"[STT] Standard mode → Deepgram STTv2 (Flux)")
        stt_provider = deepgram.STTv2(
            model="flux-general-en",
            eager_eot_threshold=0.4,
        )

    # ── METRICS / USAGE COLLECTION ─────────────────────────────────────────
    usage_collector = metrics.UsageCollector()

    # ── START SESSION ──────────────────────────────────────────────────────
    try:
        if not avatar_id:
            logger.error("[SESSION] ✗ No avatar_id resolved. Cannot start TruGen session.")
            return

        trugen_avatar = trugen.AvatarSession(avatar_id=avatar_id)

        session = AgentSession(
            stt=stt_provider,
            vad=vad_provider,
            llm=llm,
            tts=elevenlabs.TTS(voice_id=voice_id, model="eleven_flash_v2_5"),
            conn_options=SessionConnectOptions(
                llm_conn_options=APIConnectOptions(timeout=300.0, max_retry=0)
            ),
            preemptive_generation=False
        )

        @session.on("metrics_collected")
        def _on_metrics_collected(ev: MetricsCollectedEvent):
            metrics.log_metrics(ev.metrics)
            usage_collector.collect(ev.metrics)

        await trugen_avatar.start(session, room=ctx.room)

        # Groq LLM for instant waiting messages
        groq_api_key = os.getenv("GROQ_API_KEY") or "MISSING_KEY"
        groq_llm = openai.LLM(
            api_key=groq_api_key,
            base_url="https://api.groq.com/openai/v1",
            model="openai/gpt-oss-120b",
        )

        if connection_type in ("email_dispatch", "recall") or config.get("recallBotId"):
            room_opts = RoomOptions(close_on_disconnect=False)
            logger.info("[SESSION] Meeting mode active: close_on_disconnect=False")
        else:
            room_opts = NOT_GIVEN

        agent = MyAgent(
            llm=llm,
            stt=stt_provider,
            tts=elevenlabs.TTS(voice_id=voice_id, model="eleven_flash_v2_5"),
            vad=vad_provider,
            groq_llm=groq_llm,
            enable_thinking=ENABLE_LET_ME_THINK,
            thinking_delay=THINKING_DELAY,
            conversation_id=conversation_id  # Pass conversation_id for transcript tracking
        )

        ctx.add_shutdown_callback(agent._on_shutdown)
        await session.start(agent, room=ctx.room, room_options=room_opts)
        await agent.session.say("Hello! Let's get started.")
        # ── Capture greeting in transcript ──
        await agent._add_to_transcript(
            role="assistant",
            content="Hello! Let's get started.",
            message_type="message"
        )

        # ── USAGE REPORTING ────────────────────────────────────────────────
        # We fire the POST as a fully detached asyncio task the moment the
        # session closes — BEFORE LiveKit's shutdown sequence begins.
        # This means the HTTP request runs independently and is never subject
        # to LiveKit's 8-10s process kill window. No shutdown callback needed.
        session_start_time = time.time()

        @session.on("close")
        def _on_session_close():
            # Capture references to module-level functions in closure
            _push_transcript = push_transcript_to_s3
            _post_usage = post_backend_usage
            
            async def _send_usage_and_transcript():
                try:
                    summary = usage_collector.get_summary()
                    total_duration = time.time() - session_start_time
                    backend_payload = {
                        "conversation_id": conversation_id,
                        "job_id": job_id,
                        "status": "COMPLETED",
                        "usage": {
                            "total_duration": total_duration,
                            "stt": {
                                "audio_duration": summary.stt_audio_duration,
                            },
                            "tts": {
                                "characters_count": summary.tts_characters_count,
                            },
                        },
                    }
                    logger.info(f"📦 [USAGE] Backend payload: {json.dumps(backend_payload, default=str)}")
                    await _post_usage(backend_payload)
                except Exception:
                    import traceback
                    logger.error("[USAGE] Failed to post backend usage\n%s", traceback.format_exc())

                # ── UPLOAD TRANSCRIPT TO S3 ────────────────────────────────
                try:
                    transcript = agent.get_transcript()
                    if transcript:
                        logger.info(f"[TRANSCRIPT] Uploading {len(transcript)} messages to S3...")
                        _push_transcript(transcript, conversation_id)
                    else:
                        logger.warning("[TRANSCRIPT] No transcript data to upload")
                except Exception:
                    import traceback
                    logger.error("[TRANSCRIPT] Failed to upload transcript\n%s", traceback.format_exc())

            # Detached task — runs freely, not tied to shutdown window
            asyncio.ensure_future(_send_usage_and_transcript())

    except Exception as e:
        error_msg = str(e)
        if "404" in error_msg or "Avatar not found" in error_msg:
             logger.error(f"[SESSION] ✗ TruGen Avatar Error: Avatar ID '{avatar_id}' not found or invalid. Please check your configuration.")
        else:
             logger.error(f"[SESSION] ✗ Failed to start agent session: {e}")
        return


if __name__ == "__main__":
    import sys
    if len(sys.argv) > 1 and sys.argv[1] == "download-files":
        silero.VAD.load()
        sys.exit(0)

    cli.run_app(server)