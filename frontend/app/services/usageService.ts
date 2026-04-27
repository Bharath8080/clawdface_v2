export interface UsagePayload {
  conversation_id: string;
  job_id: string;
  status: string;
  usage: {
    total_duration: number;
    stt?: {
      audio_duration: number;
    };
    llm?: {
      prompt_tokens: number;
      completion_tokens: number;
    };
    tts?: {
      characters_count: number;
    };
  };
}

export const sendUsageData = async (payload: UsagePayload): Promise<{ success: boolean; error: string | null }> => {
  try {
    const baseUrl = process.env.NEXT_PUBLIC_BASE_URL || "http://localhost:3077";
    const response = await fetch(`${baseUrl}/v1/usage`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(payload),
    });

    if (response.ok) {
      return { success: true, error: null };
    } else {
      const errorText = await response.text();
      return { success: false, error: errorText || "Failed to send usage data" };
    }
  } catch (err: unknown) {
    return { success: false, error: err instanceof Error ? err.message : "Unknown error" };
  }
};
