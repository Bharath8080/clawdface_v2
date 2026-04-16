import "server-only";

import { StackServerApp } from "@stackframe/stack";

export const stackServerApp = new StackServerApp({
  tokenStore: "nextjs-cookie",
  urls: {
    accountSettings: "/settings/profile",
    signIn: "/log-in",
    signUp: "/sign-up",
    passwordReset: "/reset-password",
  },
});
