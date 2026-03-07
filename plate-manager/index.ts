import { log } from "./log";
import pkg from "./package.json";
import { Database } from "bun:sqlite";
import { sql_init } from "./sql/init";
import { mkdir } from "node:fs/promises";

await mkdir("data", { recursive: true });
export const db = new Database("data/plate.db", { create: true });

console.log("[plate::manager] Version " + pkg.version);

log("Initializing database...");
db.run("PRAGMA journal_mode = WAL;");
sql_init();
log("Database initialized successfully.");

log("Serving web server...");

const server = Bun.serve({
  fetch(req) {
    const url = new URL(req.url);

    return new Response("404 Not Found", { status: 404 });
  },
});

log("Server is running at http://localhost:" + server.port);
