import { NextRequest, NextResponse } from "next/server";

export async function POST(req: NextRequest) {
  try {
    const body = await req.json();
    const { url, token, sessionKey } = body;

    if (!url || !token) {
      return NextResponse.json({ error: "Missing required parameters" }, { status: 400 });
    }

    const cleanUrl = url.replace(/\/$/, "");
    const targetUrl = `${cleanUrl}/v1/chat/completions`;

    console.log(`DOCTOR_CHECK: Attempting validated POST to ${targetUrl}`);

    // Perform a 1-token dummy completion as the ultimate health check
    const response = await fetch(targetUrl, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Authorization": `Bearer ${token}`,
        "x-openclaw-session-key": sessionKey || "",
        "x-openclaw-agent-id": "main",
        "ngrok-skip-browser-warning": "true"
      },
      body: JSON.stringify({
        model: "openclaw",
        messages: [{ role: "user", content: "ping" }],
        max_tokens: 1,
        stream: true 
      }),
      cache: 'no-store',
      // @ts-ignore
      // INCREASED TIMEOUT: Local LLMs can be slow to initialize (cold start)
      signal: AbortSignal.timeout(60000) 
    });

    console.log(`DOCTOR_CHECK: Response status ${response.status}`);

    return NextResponse.json({
      status: response.status,
      ok: response.ok
    });
  } catch (error: any) {
    console.error("DOCTOR_API_ERROR:", error.message);
    
    if (error.name === 'TimeoutError' || error.message?.includes('timeout')) {
      return NextResponse.json({ 
        error: "Gateway connection timed out. Your local LLM is taking too long to respond (>60s). This usually happens during a cold start or when system resources are low." 
      }, { status: 504 });
    }
    
    return NextResponse.json({ error: "Could not reach gateway (is it running?)" }, { status: 502 });
  }
}
