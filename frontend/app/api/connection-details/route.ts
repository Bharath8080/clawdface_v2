import { AccessToken, AccessTokenOptions } from "livekit-server-sdk";
import { RoomAgentDispatch, RoomConfiguration } from "@livekit/protocol";
import { NextResponse } from "next/server";

const API_KEY     = process.env.LIVEKIT_API_KEY!;
const API_SECRET  = process.env.LIVEKIT_API_SECRET!;
const LIVEKIT_URL = process.env.LIVEKIT_URL!;

export const revalidate = 0;

export type ConnectionDetails = {
  serverUrl: string;
  roomName: string;
  participantName: string;
  participantToken: string;
};

export async function GET(req: Request) {
  const { searchParams } = new URL(req.url);
  const config = {
    openclawUrl:  searchParams.get("openclawUrl") || undefined,
    gatewayToken: searchParams.get("gatewayToken") || undefined,
    sessionKey:   searchParams.get("sessionKey") || undefined,
    avatarId:     searchParams.get("avatarId") || undefined,
    roomName:     searchParams.get("room") || undefined,
    meetingUrl:   searchParams.get("meetingUrl") || undefined,
    connection_type: searchParams.get("connection_type") || undefined,
    enable_thinking: searchParams.get("enable_thinking") || searchParams.get("thinkingEnabled") || undefined,
    thinking_delay:  searchParams.get("thinking_delay") || searchParams.get("thinkingDelay") || undefined,
  };
  return handleConnection(config);
}

export async function POST(req: Request) {
  const body = await req.json().catch(() => ({}));
  return handleConnection(body);
}


async function handleConnection(config: {
  openclawUrl?: string;
  gatewayToken?: string;
  sessionKey?: string;
  avatarId?: string;
  roomName?: string;
  meetingUrl?: string;
  connection_type?: string;
  enable_thinking?: string;
  thinking_delay?: string;
  conversation_id?: string;
  job_id?: string;
}) {
  try {
    if (!LIVEKIT_URL) throw new Error("LIVEKIT_URL is not defined");
    if (!API_KEY)     throw new Error("LIVEKIT_API_KEY is not defined");
    if (!API_SECRET)  throw new Error("LIVEKIT_API_SECRET is not defined");

    const participantIdentity = `user_${Math.floor(Math.random() * 10_000)}`;
    const roomName = config.roomName || `clawdface_room_${Math.floor(Math.random() * 10_000)}`;

    // Participant token metadata — carries agent config for resolve_config()
    const participantMetadata = JSON.stringify({
      openclawUrl:  config.openclawUrl  || "",
      gatewayToken: config.gatewayToken || "",
      sessionKey:   config.sessionKey   || "",
      avatarId:     config.avatarId     || "",
      meetingUrl:   config.meetingUrl   || "",
      connection_type: config.connection_type || "website",
      enable_thinking: config.enable_thinking || "true",
      thinking_delay:  config.thinking_delay  || "5.0",
      conversation_id: config.conversation_id || "",
      job_id:          config.job_id          || "",
    });

    // Agent dispatch metadata — this becomes ctx.job.metadata in agent.py.
    // MUST contain conversation_id and job_id so usage reporting uses the
    // correct DB-registered UUIDs, not a freshly generated fallback.
    // MUST contain connection_type so resolve_config() routes correctly:
    //   "website" → Deepgram STT  |  "email_dispatch" → Recall STT
    const agentMetadata = JSON.stringify({
      openclawUrl:     config.openclawUrl     || "",
      gatewayToken:    config.gatewayToken    || "",
      sessionKey:      config.sessionKey      || "",
      avatarId:        config.avatarId        || "",
      enable_thinking: config.enable_thinking || "true",
      thinking_delay:  config.thinking_delay  || "5.0",
      connection_type: config.connection_type || "website",
      conversation_id: config.conversation_id || "",
      job_id:          config.job_id          || "",
    });

    console.log(`[connection-details] Room: ${roomName}`);
    console.log(`[connection-details] Session Key: ${config.sessionKey || "(default)"}`);
    console.log(`[connection-details] conv_id → agent: ${config.conversation_id || "(none)"}`);

    const participantToken = await createParticipantToken(
      { identity: participantIdentity, metadata: participantMetadata },
      roomName,
      agentMetadata
    );

    return NextResponse.json(
      {
        serverUrl: LIVEKIT_URL,
        roomName,
        participantToken,
        participantName: participantIdentity,
      } as ConnectionDetails,
      { headers: { "Cache-Control": "no-store" } }
    );
  } catch (error) {
    if (error instanceof Error) {
      console.error("[connection-details]", error.message);
      return new NextResponse(error.message, { status: 500 });
    }
  }
}

function createParticipantToken(
  userInfo: AccessTokenOptions & { metadata?: string },
  roomName: string,
  agentMetadata: string = ""
) {
  const at = new AccessToken(API_KEY, API_SECRET, {
    identity: userInfo.identity,
    ttl: "15m",
  });
  at.metadata = userInfo.metadata || "";
  at.addGrant({
    room: roomName,
    roomJoin: true,
    canPublish: true,
    canPublishData: true,
    canSubscribe: true,
  });

  // Dispatch the named agent and embed metadata into the job so agent.py
  // receives conversation_id/job_id in ctx.job.metadata.
  at.roomConfig = new RoomConfiguration({
    agents: [
      new RoomAgentDispatch({
        agentName: 'clawdface',
        metadata: agentMetadata,
      }),
    ],
  });

  return at.toJwt();
}

