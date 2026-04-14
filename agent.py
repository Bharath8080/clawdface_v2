import os
import json
import asyncio
import base64
import requests
import typing
import logging
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
    "1fa504ff", "0f160301", "13550375", "18c4043e", "48d778c9"
}

DEFAULT_AVATAR_ID = "1a640442"

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
                return cfg, "email_dispatch"
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
        email = room_id if "@" in room_id else f"{room_id}@agent.truhire.ai"
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
    def __init__(self, *, groq_llm: llm.LLM | None = None, enable_thinking: bool = True, thinking_delay: float = 3.0, **kwargs) -> None:
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
    
    async def on_user_turn_completed(self, turn_ctx, new_message):
        """Custom handling: stream LLM, buffer if thinking needed, then release."""
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
        
        # Build chat context
        chat_ctx = turn_ctx.copy()
        chat_ctx.items.append(new_message)
        
        # Create events for coordination
        thinking_started = asyncio.Event()
        thinking_done = asyncio.Event()
        first_chunk_received = asyncio.Event()
        
        # Start LLM stream
        logger.info("[STREAM] Starting LLM stream...")
        # Access the llm provider (ensure it's the standard LLM for .chat())
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
                    
                    # Track first text chunk timing and cancel thinking timer if it's fast
                    if not first_chunk_received.is_set():
                        first_chunk_received.set()
                        elapsed = asyncio.get_event_loop().time() - stream_start_time
                        logger.info(f"[STREAM] First text chunk after {elapsed:.2f}s")
                        
                        # Fast response: cancel timer, don't play thinking
                        if not thinking_started.is_set():
                            logger.info("[STREAM] Fast response, cancelling thinking timer")
                            thinking_task.cancel()
                            try:
                                await thinking_task
                            except asyncio.CancelledError:
                                pass
                            thinking_done.set()  # Signal that "thinking" is done (skipped)
                        else:
                            # Thinking is playing - let it finish, just buffer the response
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
            except RuntimeError as e:
                if "AgentSession is closing" in str(e):
                    logger.info("[STREAM] Session closing, skipping speak")
                else:
                    raise
        else:
            logger.warning("[STREAM] Empty response, nothing to speak")
    
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
            
            # Only play if not already cancelled
            started_event.set()
            
            # Generate dynamic waiting message using Groq
            waiting_message = await self._generate_waiting_message()
            
            logger.info(f"[THINKING] Timer fired, saying: '{waiting_message}'")
            self._thinking_played = True
            # No interruptions - let the full message play for better UX
            # Play the filler phrase (non-interruptible to ensure it finishes)
            try:
                await self.session.say(waiting_message, allow_interruptions=False)
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
            # Fallback to static if Groq not available
            fallbacks = [
                "Let me think about that for a moment.",
                "Hmm, give me a second to consider this.",
                "Just a moment while I work through that.",
            ]
            return random.choice(fallbacks)
        
        try:
            # Create prompt for Groq to generate waiting message
            user_query = self._last_user_message.content or ""
            
            waiting_prompt = (
                "You are a helpful AI assistant. Generate a brief, natural-sounding "
                "waiting phrase that acknowledges the SPECIFIC TOPIC from the user's question. "
                "You MUST reference the actual subject they asked about.\n\n"
                f"User's question: '{user_query}'\n\n"
                "RULES:\n"
                "1. MUST mention the specific topic from their question\n"
                "2. 3-8 words maximum\n"
                "3. Sound natural and conversational\n"
                "4. NEVER use generic phrases like 'Let me think about that'\n"
                "5. NEVER use 'that' - use the actual topic instead\n\n"
                "GOOD examples (notice how they mention the topic):\n"
                "- User asks about chess → 'Let me think about chess strategies...'\n"
                "- User asks about weather → 'Checking the weather forecast...'\n"
                "- User asks about Python → 'Hmm, let me consider Python approaches...'\n"
                "- User asks about cooking → 'Thinking about cooking techniques...'\n"
                "- User asks about history → 'Exploring historical facts...'\n\n"
                "BAD examples (too generic, don't use):\n"
                "- 'Let me think about that' ❌\n"
                "- 'One moment please' ❌\n"
                "- 'Give me a second' ❌\n\n"
                "Now generate YOUR context-aware waiting phrase (mention the topic!):"
            )
            
            # Quick Groq call for instant response
            from livekit.agents import llm as llm_types
            chat_ctx = llm_types.ChatContext()
            chat_ctx.add_message(role="user", content=waiting_prompt)
            
            response_text = ""
            async for chunk in self._groq_llm.chat(chat_ctx=chat_ctx):
                text = self._extract_text_from_chunk(chunk)
                if text:
                    response_text += text
            
            # Clean up the response
            waiting_msg = response_text.strip().strip('"').strip("'")
            
            # Ensure it's not too long
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
    # Dynamic thinking settings resolved from metadata below

    await ctx.connect()

    config, connection_type = {}, "unknown"
    for _ in range(30):
        config, connection_type = resolve_config(ctx)
        if config: break
        await asyncio.sleep(0.5)

    logger.info(f"[METADATA] Resolution Completed | Mode: {connection_type.upper()}")
    logger.info(f"[METADATA] Room Metadata: {ctx.room.metadata[:300]}")
    logger.info(f"[METADATA] Job Metadata: {ctx.job.metadata[:300]}")

    if not config:
        logger.error(f"[SESSION] ✗ Failed to resolve config for room {ctx.room.name}")
        return

    url = config.get("openclawUrl", "").strip()
    token = config.get("gatewayToken", "")
    key = config.get("sessionKey", "")
    avatar_id = config.get("avatarId") or os.getenv("TRUGEN_AVATAR_ID") or DEFAULT_AVATAR_ID
    voice_id = "CwhRBWXzGAHq8TQ4Fs17" if avatar_id in MALE_AVATAR_IDS else "FGY2WhTYpPnrIDTdsKH5"
    
    # Resolve dynamic thinking settings check both snake_case and camelCase
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

    session = AgentSession(
        stt=stt_provider,
        vad=vad_provider,
        llm=llm,
        tts=elevenlabs.TTS(voice_id=voice_id, model="eleven_flash_v2_5"),
        conn_options=SessionConnectOptions(
            llm_conn_options=APIConnectOptions(timeout=300.0, max_retry=0)
        ),
        # preemptive_generation=False: We manually control LLM timing to avoid
        # double requests. We stream the LLM response but buffer it until
        # after "let me think" plays, then release it.
        preemptive_generation=False
    )

    # Note: All thinking logic is now handled inside MyAgent class
    # to enable single LLM request with streaming/buffering approach

    # Groq LLM for instant waiting messages
    groq_api_key = os.getenv("GROQ_API_KEY") or "MISSING_KEY"
    groq_llm = openai.LLM(
        api_key=groq_api_key,
        base_url="https://api.groq.com/openai/v1",
        model="openai/gpt-oss-120b",  # Fast Groq model for instant waiting messages
    )

    if connection_type in ("email_dispatch", "recall") or config.get("recallBotId"):
        room_opts = RoomOptions(close_on_disconnect=False)
        logger.info("[SESSION] Meeting mode active: close_on_disconnect=False")
    else:
        room_opts = NOT_GIVEN

    try:
        trugen_avatar = trugen.AvatarSession(avatar_id=avatar_id)
        await trugen_avatar.start(session, room=ctx.room)
        
        # In LiveKit 1.x, we use the high-level Agent properties correctly.
        agent = MyAgent(
            llm=llm,
            stt=stt_provider,
            tts=elevenlabs.TTS(voice_id=voice_id, model="eleven_flash_v2_5"),
            vad=vad_provider,
            groq_llm=groq_llm, 
            enable_thinking=ENABLE_LET_ME_THINK, 
            thinking_delay=THINKING_DELAY
        )
        
        await session.start(agent, room=ctx.room, room_options=room_opts)
        await agent.session.say("Hello! Let's get started.")
    except Exception as e:
        logger.error(f"[SESSION] ✗ Fatal error: {e}")
        raise


if __name__ == "__main__":
    import sys
    if len(sys.argv) > 1 and sys.argv[1] == "download-files":
        silero.VAD.load()
        sys.exit(0)

    cli.run_app(server)