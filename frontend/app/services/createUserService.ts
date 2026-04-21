interface CreateUser {
  id?: string;
  email?: string;
  first_name?: string;
  last_name?: string;
  company?: string;
  password_hash?: string;
  state?: string;
  country?: string;
  address?: string;
  two_factor_enabled?: boolean;
  profile_Image?: string;
  subscribe_to_mailer?: boolean;
}

export const createUserService = async (userDetails: CreateUser) => {
  try {
    const response = await fetch(`/v1/user`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(userDetails),
    });
    if (response.ok) {
      return true;
    }
    return false;
  } catch (error) {
    console.error("Error creating user:", error);
    return false;
  }
};

export const createUserServiceServer = async (userDetails: CreateUser) => {
  try {
    const response = await fetch(
      `${process.env.NEXT_PUBLIC_BASE_URL}/v1/user`,
      {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(userDetails),
      },
    );
    if (response.ok) {
      return true;
    }
    return false;
  } catch (error) {
    console.error("Error creating user:", error);
    return false;
  }
};

export const updateUserService = async (
  userId: string,
  userDetails: CreateUser,
) => {
  try {
    const response = await fetch(
      `${process.env.NEXT_PUBLIC_BASE_URL}/v1/user/${userId}`,
      {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(userDetails),
      },
    );
    if (response.ok) {
      return true;
    }
    return false;
  } catch (error) {
    console.error("Error UPDATING user:", error);
    return false;
  }
};
