export const getUserService = async (id: string) => {
  try {
    console.log("Fetching user with ID:", id);
    const response = await fetch(`/v1/user/${id}`);
    if (response.ok) {
      const userdata = await response.json();
      return { ...userdata, profileImageUrl: null };
    } else {
      return null;
    }
  } catch (error) {
    console.error("Error getting user:", error);
    return null;
  }
};

export const getUserServiceServer = async (id: string) => {
  try {
    const response = await fetch(
      `${process.env.NEXT_PUBLIC_BASE_URL}/v1/user/${id}`
    );
    if (response.ok) {
      const userdata = await response.json();
      return { ...userdata, profileImageUrl: null };
    } else {
      return null;
    }
  } catch (error) {
    console.error("Error getting user:", error);
    return null;
  }
};
