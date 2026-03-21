// Cloudflare Worker resolve + redirect example
const LINK_BASE = "https://link.example.com";
const PLATE_ID = "1";

export default {
  async fetch(request) {
    const url = new URL(request.url);
    const parts = url.pathname.replace(/^\/+/, "").split("/");
    // Expects /u/:id/:tail*
    if (parts[0] !== "u" || !parts[1]) {
      return new Response("Not found", { status: 404 });
    }
    const id = parts[1];
    const tail = parts.slice(2).join("/");
    const resolveUrl = `${LINK_BASE}/${PLATE_ID}/resolve/${id}${tail ? `/${tail}` : ""}${url.search}`;

    const resolved = await fetch(resolveUrl, {
      method: "GET",
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

    return Response.redirect(destination, 307);
  },
};
