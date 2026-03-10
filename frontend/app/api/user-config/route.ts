import { NextResponse } from "next/server";
import { supabase } from "@/lib/supabase";

export async function GET(req: Request) {
  const { searchParams } = new URL(req.url);
  const email = searchParams.get("email");

  if (!email) {
    return NextResponse.json({ error: "Email is required" }, { status: 400 });
  }

  try {
    const { data, error } = await supabase
      .from("user_configs")
      .select("openclaw_url, gateway_token, session_key")
      .eq("email", email.toLowerCase())
      .single();

    if (error && error.code !== "PGRST116") {
      console.error("[user-config] Error fetching config:", error);
      return NextResponse.json({ error: "Failed to read config" }, { status: 500 });
    }

    if (!data) {
      return NextResponse.json({});
    }

    // Map DB fields back to frontend expected camelCase if necessary, 
    // or keep them as is if the frontend matches. 
    // The previous JSON structure was likely: { openclawUrl, gatewayToken, sessionKey }
    // Let's check the previous code or just return a compatible object.
    return NextResponse.json({
      openclawUrl: data.openclaw_url,
      gatewayToken: data.gateway_token,
      sessionKey: data.session_key
    });
  } catch (error: any) {
    return NextResponse.json({ error: "Unexpected error" }, { status: 500 });
  }
}

export async function POST(req: Request) {
  try {
    const body = await req.json();
    const { email, config } = body;

    if (!email) {
      return NextResponse.json({ error: "Email is required" }, { status: 400 });
    }

    const { error } = await supabase
      .from("user_configs")
      .upsert({
        email: email.toLowerCase(),
        openclaw_url: config.openclawUrl,
        gateway_token: config.gatewayToken,
        session_key: config.sessionKey
      });

    if (error) {
      console.error("[user-config] Error saving config:", error);
      return NextResponse.json({ error: "Failed to save config" }, { status: 500 });
    }

    return NextResponse.json({ success: true });
  } catch (error) {
    return NextResponse.json({ error: "Failed to save config" }, { status: 500 });
  }
}
