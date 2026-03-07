import { log } from "./log";
import pkg from "./package.json";
import { Database } from "bun:sqlite";
import { sql_init } from "./sql/init";
import { mkdir } from "node:fs/promises";

await mkdir("data", { recursive: true });
export const db = new Database("data/plate.db", { create: true });

log("Version " + pkg.version);

log("Initializing database...");
db.run("PRAGMA journal_mode = WAL;");
sql_init();
log("Database initialized successfully.");

log("Serving web server...");

const server = Bun.serve({
  port: process.env.PORT ? parseInt(process.env.PORT) : 3200,
  fetch(req) {
    const url = new URL(req.url);
    log("INBOUND", req.method, url.pathname);
    if (url.pathname === "/") {
      const now = new Date();
      const timestamp = `${String(now.getHours()).padStart(2, "0")}:${String(now.getMinutes()).padStart(2, "0")}:${String(now.getSeconds()).padStart(2, "0")}.${String(now.getMilliseconds()).padStart(3, "0")}`;

      return new Response(
        "Plate Manager v" + pkg.version + " (" + timestamp + ")",
        {
          status: 200,
        },
      );
    }

    if (url.pathname.startsWith("/auth")) {
    }

    return new Response("[404] Not Found", { status: 404 });
  },
});

log("Server is running at http://localhost:" + server.port);
