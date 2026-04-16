import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";
import { stackServerApp } from "./stack";

export async function middleware(req: NextRequest) {
  const user = await stackServerApp.getUser({ or: "return-null" });

  if (!user) return NextResponse.next();

  const { pathname, searchParams } = req.nextUrl;

  if (pathname !== "/log-in") {
    return NextResponse.next();
  }

  const redirectTo = searchParams.get("redirectTo");

  if (redirectTo && redirectTo.startsWith("/")) {
    return NextResponse.redirect(new URL(redirectTo, req.url));
  }

  return NextResponse.redirect(new URL("/dashboard", req.url));
}

export const config = {
  matcher: "/log-in",
};
