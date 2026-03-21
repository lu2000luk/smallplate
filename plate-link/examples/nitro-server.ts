// Nitro/H3 resolve + redirect example
import { defineEventHandler, getRequestURL, sendRedirect } from "h3";

const LINK_BASE = process.env.PLATE_LINK_BASE || "https://link.example.com";
const PLATE_ID = process.env.PLATE_ID || "1";

export default defineEventHandler(async (event) => {
  const url = getRequestURL(event);
  if (!url.pathname.startsWith("/u/")) {
    return;
  }

  const parts = url.pathname.split("/").filter(Boolean);
  const id = parts[1];
  const tail = parts.slice(2).join("/");
  if (!id) return;

  const resolveUrl = `${LINK_BASE}/${PLATE_ID}/resolve/${id}${tail ? `/${tail}` : ""}${url.search}`;
  const resolved = await fetch(resolveUrl, {
    headers: { "Accept": "application/json" },
  });

  if (!resolved.ok) {
    return new Response("Link not found", { status: resolved.status });
  }

  const body = await resolved.json();
  const destination = body?.data?.destination;
  if (!destination) {
    return new Response("Invalid resolve payload", { status: 502 });
  }

  return sendRedirect(event, destination, 307);
});
