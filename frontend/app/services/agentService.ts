export interface AgentBot {
  id: string;
  agent_name: string;
  agent_system_prompt: string;
  email: string;
  config: {
    openclaw_url: string;
    gateway_token: string;
    session_key: string;
    thinking_enabled: boolean;
    thinking_delay: number;
  };
  tools: Record<string, any>;
  avatars: Array<{ avatar_key_id: string }>;
  is_active: boolean;
  is_public: boolean;
  type: string;
  add_on: Array<{ type: string }>;
  created_at?: string;
}

export interface CreateAgentPayload {
  agent_name: string;
  agent_system_prompt: string;
  email: string;
  config: Record<string, any>;
  tools: Record<string, any>;
  avatars: any;
  is_active: boolean;
  is_public: boolean;
  type: string;
  add_on: Array<{ type: string }>;
  knowledge_base?: Array<{ id: string; name: string }>;
  record?: boolean;
}

export const createAgent = async (
  apiKey: string,
  body: CreateAgentPayload
): Promise<{ data: any; status?: number; error: string | null }> => {
  try {
    const response = await fetch(`${process.env.NEXT_PUBLIC_BASE_URL}/v1/agent/`, {
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
      const errorText = await response.json();
      return {
        data: null,
        status: response.status,
        error: `${errorText?.error || errorText?.message || "An error occurred"}`,
      };
    }
  } catch (err: unknown) {
    const errorMessage = err instanceof Error ? err.message : "Unknown error";
    console.error("Error creating agent:", err);
    return { data: null, error: errorMessage };
  }
};

export const updateAgent = async (
  apiKey: string,
  id: string,
  body: Partial<CreateAgentPayload>
): Promise<{ data: any; status?: number; error: string | null }> => {
  try {
    const response = await fetch(`${process.env.NEXT_PUBLIC_BASE_URL}/v1/agent/${id}`, {
      method: "PUT",
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
      const errorText = await response.json();
      return {
        data: null,
        status: response.status,
        error: `${errorText?.error || errorText?.message || "An error occurred"}`,
      };
    }
  } catch (err: unknown) {
    const errorMessage = err instanceof Error ? err.message : "Unknown error";
    console.error("Error updating agent:", err);
    return { data: null, error: errorMessage };
  }
};

export const deleteAgent = async (
  apiKey: string,
  id: string
): Promise<{ error: string | null }> => {
  try {
    const response = await fetch(`${process.env.NEXT_PUBLIC_BASE_URL}/v1/agent/${id}`, {
      method: "DELETE",
      headers: { "X-API-Key": apiKey },
    });
    if (response.ok) return { error: null };
    const errorText = await response.json();
    return { error: `${errorText?.error || errorText?.message || "An error occurred"}` };
  } catch (err: unknown) {
    return { error: err instanceof Error ? err.message : "Unknown error" };
  }
};

export const getAgents = async (
  apiKey: string
): Promise<{ data: AgentBot[] | null; status?: number; error: string | null }> => {
  try {
    const response = await fetch(`${process.env.NEXT_PUBLIC_BASE_URL}/v1/agent/`, {
      method: "GET",
      headers: { "X-API-Key": apiKey },
    });
    if (response.ok) {
      const data = await response.json();
      return { data: data?.filter((el: any) => el?.is_active) ?? [], error: null };
    } else {
      const errorText = await response.json();
      return {
        data: null,
        status: response.status,
        error: `${errorText?.error || errorText?.message || "An error occurred"}`,
      };
    }
  } catch (err: unknown) {
    const errorMessage = err instanceof Error ? err.message : "Unknown error";
    console.error("Error fetching agents:", err);
    return { data: null, error: errorMessage };
  }
};
