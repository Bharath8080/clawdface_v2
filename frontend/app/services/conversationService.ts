export interface ConversationPayload {
  agentId: string;
  userName: string;
  userId: string;
  mode: string;
  context: { text: string };
  metadata: { active: string };
}

export interface ConversationItem {
  id: string;
  agentId?: string;
  userName?: string;
  userId?: string;
  status?: string;
  created_at?: string;
  context?: { text?: string };
  metadata?: Record<string, any>;
}

export const getConversations = async (
  apiKey: string
): Promise<{ data: ConversationItem[] | null; error: string | null }> => {
  try {
    const response = await fetch(`${process.env.NEXT_PUBLIC_BASE_URL}/v1/conversation/`, {
      method: "GET",
      headers: { "X-API-Key": apiKey },
    });
    if (response.ok) {
      const data = await response.json();
      return { data: Array.isArray(data) ? data : data?.results ?? [], error: null };
    } else {
      const errorText = await response.json().catch(() => ({}));
      return { data: null, error: errorText?.error || errorText?.message || "An error occurred" };
    }
  } catch (err: unknown) {
    return { data: null, error: err instanceof Error ? err.message : "Unknown error" };
  }
};

export const getConversationById = async (
  apiKey: string,
  conversationId: string
): Promise<{ data: any | null; error: string | null }> => {
  try {
    const response = await fetch(
      `${process.env.NEXT_PUBLIC_BASE_URL}/v1/conversation/${conversationId}`,
      {
        method: "GET",
        headers: { "X-API-Key": apiKey },
      }
    );
    if (response.ok) {
      const data = await response.json();
      return { data, error: null };
    } else {
      const errorText = await response.json().catch(() => ({}));
      return { data: null, error: errorText?.error || errorText?.message || "An error occurred" };
    }
  } catch (err: unknown) {
    return { data: null, error: err instanceof Error ? err.message : "Unknown error" };
  }
};

export const createConversation = async (
  apiKey: string,
  body: ConversationPayload
): Promise<{ data: any; error: string | null }> => {
  try {
    const response = await fetch(`${process.env.NEXT_PUBLIC_BASE_URL}/v1/conversation`, {
      method: "POST",
      headers: {
        "X-API-Key": apiKey,
        "Content-Type": "application/json",
      },
      body: JSON.stringify(body),
    });
    if (response.ok) {
      const data = await response.json();
      return { data, error: null };
    } else {
      const errorText = await response.json().catch(() => ({}));
      return { data: null, error: errorText?.error || errorText?.message || "An error occurred" };
    }
  } catch (err: unknown) {
    return { data: null, error: err instanceof Error ? err.message : "Unknown error" };
  }
};
