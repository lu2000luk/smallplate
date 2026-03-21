// Vite middleware resolve + redirect example
import type { Plugin } from "vite";

const LINK_BASE = process.env.PLATE_LINK_BASE || "https://link.example.com";
const PLATE_ID = process.env.PLATE_ID || "1";

export function linkRedirectPlugin(): Plugin {
  return {
    name: "link-redirect-plugin",
    configureServer(server) {
      server.middlewares.use(async (req, res, next) => {
        const url = req.url || "/";
        if (!url.startsWith("/u/")) return next();

        const [path, query = ""] = url.split("?");
        const parts = path.split("/").filter(Boolean); // [u,id,...tail]
        const id = parts[1];
        const tail = parts.slice(2).join("/");
        if (!id) return next();

        const resolveUrl = `${LINK_BASE}/${PLATE_ID}/resolve/${id}${tail ? `/${tail}` : ""}${query ? `?${query}` : ""}`;
        try {
          const resolved = await fetch(resolveUrl, {
            headers: { "Accept": "application/json" },
          });
          if (!resolved.ok) {
            res.statusCode = resolved.status;
            res.end("Link not found");
            return;
          }
          const body = await resolved.json();
          const destination = body?.data?.destination;
          if (!destination) {
            res.statusCode = 502;
            res.end("Invalid resolve payload");
            return;
          }
          res.statusCode = 307;
          res.setHeader("Location", destination);
          res.end();
        } catch {
          res.statusCode = 502;
          res.end("Resolve request failed");
        }
      });
    },
  };
}
