export type ApiKey = {
  created: string;
  description: string;
  expire_at: string;
  id: string;
  key_hash: string;
  name: string;
  is_default: boolean;
  workspace_id: string;
};

export type ApiKeyResponse = ApiKey[];

export type CreateApiKeyRequest = {
  user_id?: string;
  name?: string;
  description?: string;
  is_default?: boolean;
  workspace_id?: string;
};

export const getAllApiKeys = async (
  userId: string,
  token: string
): Promise<ApiKeyResponse | null> => {
  try {
    const response = await fetch(
      `${process.env.NEXT_PUBLIC_BASE_URL}/v1/apikey?user_id=${userId}`,
      {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      }
    );
    if (response.ok) {
      const data = (await response.json()) as ApiKeyResponse;
      return data;
    } else {
      return null;
    }
  } catch (error) {
    console.error("Error fetching API keys:", error);
    return null;
  }
};

export const createApiKey = async (
  token: string,
  body: CreateApiKeyRequest
): Promise<ApiKey | null> => {
  try {
    const response = await fetch(
      `${process.env.NEXT_PUBLIC_BASE_URL}/v1/apikey`,
      {
        headers: {
          Authorization: `Bearer ${token}`,
          "Content-Type": "application/json",
        },
        method: "POST",
        body: JSON.stringify(body),
      }
    );
    if (response.ok) {
      const data = (await response.json()) as ApiKey;
      return data;
    } else {
      const err = await response.json().catch(() => ({}));
      console.error("createApiKey failed:", response.status, err);
      return null;
    }
  } catch (error) {
    console.error("Error creating API key:", error);
    return null;
  }
};

/**
 * Fetches the user's workspace ID from the backend.
 * Required when creating an API key so it is scoped to the correct workspace.
 */
const getUserWorkspaceId = async (
  userId: string,
  accessToken: string
): Promise<string | null> => {
  try {
    const response = await fetch(
      `${process.env.NEXT_PUBLIC_BASE_URL}/v1/user/${userId}`,
      {
        headers: { Authorization: `Bearer ${accessToken}` },
      }
    );
    if (response.ok) {
      const data = await response.json();
      return data?.organizations?.[0]?.workspaces?.[0]?.id ?? null;
    }
    return null;
  } catch {
    return null;
  }
};

/**
 * Ensures a valid default API key exists for the user.
 *
 * - Always fetches the key list from the backend (no stale-cache issues).
 * - If no default key exists, fetches the workspace ID first, then creates one.
 * - Stores the key_hash in localStorage as "defaultApiKey" for use by other services.
 * - Clears any previously cached key before starting so an invalid cached key is never reused.
 */
export const initDefaultApiKey = async (
  userId: string,
  accessToken: string
): Promise<string | null> => {
  try {
    // Always clear any previously cached key to avoid reusing an invalid one
    localStorage.removeItem("defaultApiKey");

    const apiKeys = await getAllApiKeys(userId, accessToken);
    const defaultKey = apiKeys?.find((k) => k.is_default);

    if (defaultKey) {
      localStorage.setItem("defaultApiKey", defaultKey.key_hash);
      return defaultKey.key_hash;
    }

    // No default key yet — fetch workspace ID first, then create one
    const workspaceId = await getUserWorkspaceId(userId, accessToken);

    const newKey = await createApiKey(accessToken, {
      user_id: userId,
      name: "Default Key",
      description: "Default Key for ClawdFace",
      is_default: true,
      ...(workspaceId ? { workspace_id: workspaceId } : {}),
    });

    if (newKey) {
      localStorage.setItem("defaultApiKey", newKey.key_hash);
      return newKey.key_hash;
    }

    console.warn("Could not create a default API key. Subscription features will be unavailable.");
    return null;
  } catch (error) {
    console.error("Error initializing default API key:", error);
    return null;
  }
};

/** Clears the cached API key (call on sign-out). */
export const clearApiKey = () => {
  localStorage.removeItem("defaultApiKey");
};
