// Next.js (App Router) resolve + redirect example
// File: app/u/[...parts]/route.ts
import { NextRequest, NextResponse } from "next/server";

const LINK_BASE = process.env.PLATE_LINK_BASE || "https://link.example.com";
const PLATE_ID = process.env.PLATE_ID || "1";

export async function GET(req: NextRequest, ctx: { params: { parts: string[] } }) {
  const parts = ctx.params.parts || [];
  const [id, ...tail] = parts;
  if (!id) return NextResponse.json({ error: "missing id" }, { status: 400 });

  const query = req.nextUrl.searchParams.toString();
  const resolvePath = `/${PLATE_ID}/resolve/${id}${tail.length ? `/${tail.join("/")}` : ""}${query ? `?${query}` : ""}`;
  const response = await fetch(`${LINK_BASE}${resolvePath}`, {
    headers: { "Accept": "application/json" },
    cache: "no-store",
  });

  if (!response.ok) {
    return NextResponse.json({ error: "link_not_resolved" }, { status: response.status });
  }

  const body = await response.json();
  const destination: string | undefined = body?.data?.destination;
  if (!destination) {
    return NextResponse.json({ error: "missing_destination" }, { status: 502 });
  }

  return NextResponse.redirect(destination, { status: 307 });
}
