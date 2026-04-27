export interface AvatarItem {
  id: string;        // avatar_key_id — used as avatarId in sessions
  name: string;      // avatar_name
  image: string;     // display_picture URL
}

interface AvatarAPIResponse {
  id: string;
  avatar_key_id: string;
  avatar_name: string;
  display_picture: string;
  image_url: string;
  gender: string;
  default_prompt: string;
}

export const fetchAvatars = async (
  apiKey: string
): Promise<{ data: AvatarItem[] | null; status?: number; error: string | null }> => {
  try {
    const response = await fetch(`${process.env.NEXT_PUBLIC_BASE_URL}/v1/public/avatar`, {
      headers: { "X-API-Key": apiKey },
      method: "GET",
    });
    if (response.ok) {
      const data = (await response.json()) as AvatarAPIResponse[];
      return {
        data: data.map((a) => ({
          id: a.avatar_key_id,
          name: a.avatar_name,
          image: a.display_picture,
        })),
        error: null,
      };
    } else {
      const errorText = await response.text();
      let errorMessage = "An error occurred";
      try {
        const errorJson = JSON.parse(errorText);
        errorMessage = errorJson.error || errorJson.message || errorText;
      } catch {
        errorMessage = errorText || "An error occurred";
      }
      return {
        data: null,
        status: response.status,
        error: errorMessage,
      };
    }
  } catch (err: unknown) {
    const errorMessage = err instanceof Error ? err.message : "Unknown error";
    console.error("Error fetching avatars:", err);
    return { data: null, error: errorMessage };
  }
};
