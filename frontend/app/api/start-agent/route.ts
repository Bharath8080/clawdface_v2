import { NextResponse } from 'next/server';
import { RoomServiceClient, AgentDispatchClient } from 'livekit-server-sdk';

function generateRoomId(): string {
  const now = new Date();
  const format = now.toISOString().slice(0, 19).replace(/:/g, '-');
  return `room-${format}`;
}

function generateSessionKey(): string {
  const now = new Date();
  const format = now.toISOString().slice(0, 19).replace(/:/g, '-');
  return `session-${format}`;
}

export async function POST(request: Request) {
  try {
    const body = await request.json();
    const {
      email,
      meetingUrl,
      startTime,
      roomId: requestedRoomId,
      userName,
      userId,
    } = body;

    if (!email) {
      return NextResponse.json({ error: 'Missing email' }, { status: 400 });
    }

    const roomName   = requestedRoomId || generateRoomId();
    const sessionKey = generateSessionKey();
    const isExternalMeeting = !!meetingUrl;

    const BACKEND_BASE_URL = process.env.NEXT_PUBLIC_BASE_URL || 'https://api.clawdface.ai';

    const conversationRes = await fetch(`${BACKEND_BASE_URL}/v1/public/conversation/byemail`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        email,
        roomId:     roomName,
        meetingURL: meetingUrl  || '',
        startTime:  startTime   || new Date().toISOString(),
        userName:   userName    || '',
        userId:     userId      || email,
        context:    { text: '' },
        metadata:   { active: 'true' },
      }),
    });

    if (!conversationRes.ok) {
      const errText = await conversationRes.text().catch(() => '');
      console.error('[start-agent] byemail error:', errText);
      return NextResponse.json(
        { error: `Failed to fetch agent details (${conversationRes.status})` },
        { status: conversationRes.status }
      );
    }

    const agentData = await conversationRes.json();

    if (!agentData) {
      return NextResponse.json({ error: 'Agent not found' }, { status: 404 });
    }

    const conversationId = agentData.conversationId || '';
    const avatarId       = agentData.id             || '';
    const agentName      = email.split('@')[0]       || agentData.name || 'AI Assistant';
    const openclawUrl    = agentData.openclaw_url   || '';
    const gatewayToken   = agentData.gateway_token  || '';

    const API_KEY     = process.env.LIVEKIT_API_KEY;
    const API_SECRET  = process.env.LIVEKIT_API_SECRET;
    const LIVEKIT_URL = process.env.LIVEKIT_URL;

    if (!LIVEKIT_URL || !API_KEY || !API_SECRET) {
      return NextResponse.json({ error: 'LiveKit configuration is missing' }, { status: 500 });
    }

    const roomService    = new RoomServiceClient(LIVEKIT_URL, API_KEY, API_SECRET);
    const dispatchClient = new AgentDispatchClient(LIVEKIT_URL, API_KEY, API_SECRET);

    await roomService.createRoom({
      name:            roomName,
      emptyTimeout:    10 * 60,
      maxParticipants: 10,
    });

    const metadata = JSON.stringify({
      openclawUrl,
      gatewayToken,
      sessionKey,
      avatarId,
      name:            agentName,
      agentName,
      meetingUrl:      meetingUrl     || '',
      recallBotId:     '',
      roomName,
      conversation_id: conversationId,
      user_email:      email,
      connection_type: isExternalMeeting ? 'email_dispatch' : 'website',
      max_call_duration: agentData.max_call_duration,
      thinking_delay:    agentData.thinking_delay,
      thinking_enabled:  agentData.thinking_enabled,
    });

    await dispatchClient.createDispatch(roomName, 'clawdface', { metadata });
    console.log(`[start-agent] ✓ Dispatched → room=${roomName} | avatarId=${avatarId} | convId=${conversationId}`);

    const baseAppUrl = process.env.NEXT_PUBLIC_APP_URL || 'http://localhost:3000';
    const videoUrl = `${baseAppUrl}/avatar?room=${encodeURIComponent(roomName)}&avatarId=${avatarId}&openclawUrl=${encodeURIComponent(openclawUrl)}&gatewayToken=${gatewayToken}&sessionKey=${sessionKey}&conversationId=${encodeURIComponent(conversationId)}&connection_type=${isExternalMeeting ? 'email_dispatch' : 'website'}`;

    if (isExternalMeeting) {
      await fetch(`${BACKEND_BASE_URL}/v1/ext/recall-trigger`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          meeting_url: meetingUrl,
          room_name: roomName,
          conversation_id: conversationId,
          video_url: videoUrl,
        }),
      }).catch(err => console.error('[start-agent] Recall trigger failed:', err));
    }

    return NextResponse.json({
      videoUrl,
      userEmail:      email,
      agentName,
      avatarId,
      roomName,
      sessionKey,
      conversationId,
      max_call_duration: agentData.max_call_duration,
      thinking_delay:    agentData.thinking_delay,
      thinking_enabled:  agentData.thinking_enabled,
    });

  } catch (error: unknown) {
    const message = error instanceof Error ? error.message : String(error);
    console.error('[start-agent] Unhandled error:', error);
    return NextResponse.json({ error: message }, { status: 500 });
  }
}