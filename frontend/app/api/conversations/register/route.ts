import { NextResponse } from 'next/server';

export async function POST(request: Request) {
  try {
    const body = await request.json();
    const { conversation_id, user_email } = body;

    // Logging for visibility, but NO database storage as requested.
    console.log(`[API] Registration received (Stateless Mode) | ID: ${conversation_id} | Email: ${user_email}`);

    return NextResponse.json({ 
      success: true, 
      conversation_id: conversation_id,
      message: "Registration accepted (stateless)" 
    });
  } catch (error: any) {
    console.error('[API] Registration parse error:', error.message);
    // Still return 200 to prevent blocking the agent/frontend
    return NextResponse.json({ success: true });
  }
}
