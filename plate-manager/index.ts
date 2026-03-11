import { log } from "./log";
import pkg from "./package.json";
import { Database } from "bun:sqlite";
import { sql_init } from "./sql/init";
import { mkdir } from "node:fs/promises";
import { authRouter } from "./accounts";
import "dotenv/config";
import {
  SERVICE_KEY,
  handlePingMessage,
  startPingLoop,
  stopPingLoop,
} from "./service";
type PingPacket = {
  type: "ping" | "pong";
};
import type { ServerWebSocket } from "bun";

await mkdir("data", { recursive: true });
export const db = new Database("data/plate.db", { create: true });

export let connectedServers: Record<string, ConnectedServer> = {};

log("Version " + pkg.version);

log("Initializing database...");
db.run("PRAGMA journal_mode = WAL;");
sql_init();
log("Database initialized successfully.");

log("Serving web server...");

export type ServerTypes = "db" | "kv";
export type ServerWSData = {
  type: ServerTypes;
  latency: number;
  id: string;
};
export type ConnectedServer = {
  type: ServerTypes;
  socket: ServerWebSocket<ServerWSData>;
  latency: number;
};

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
      return authRouter(req, url);
    }

    if (url.pathname.startsWith("/__service")) {
      const type = url.searchParams.get("t") || "";
      const id = url.searchParams.get("id") || "";
      const validTypes = ["db", "kv"];

      if (!validTypes.includes(type)) {
        return new Response("Invalid service type", { status: 400 });
      }

      const key = url.searchParams.get("k") || "";

      if (!key) {
        return new Response("Missing service key", { status: 400 });
      }

      if (key !== SERVICE_KEY) {
        return new Response("Invalid service key", { status: 403 });
      }

      if (!id) {
        return new Response("Missing service id", { status: 400 });
      }

      if (
        server.upgrade(req, {
          data: {
            type: type as ServerTypes,
            latency: 100,
            id,
          },
        })
      ) {
        return;
      }
      return new Response("Upgrade failed", { status: 500 });
    }

    return new Response("[404] Not Found", { status: 404 });
  },
  websocket: {
    data: {} as ServerWSData,
    open(ws) {
      const { type } = ws.data;
      ws.subscribe(type);
      log("New server connected:", type);

      startPingLoop(ws);
    },
    message(ws, message) {
      const rawMessage =
        typeof message === "string" ? message : message.toString();

      let parsedMessage: unknown;
      try {
        parsedMessage = JSON.parse(rawMessage);
      } catch {
        ws.sendText(
          '{"type":"error", "message":"I swear if you continue sending bad JSON imma terminate the socket", "short":"bad_json"}',
        );
        return;
      }

      if (
        typeof parsedMessage === "object" &&
        parsedMessage !== null &&
        "type" in parsedMessage &&
        (parsedMessage.type === "ping" || parsedMessage.type === "pong")
      ) {
        handlePingMessage(ws, parsedMessage as PingPacket);
      }
    },
    close(ws, code, reason) {
      stopPingLoop(ws);
      const { type } = ws.data;
      log("Server disconnected:", type, "Code:", code, "Reason:", reason);

      for (const [id, server] of Object.entries(connectedServers)) {
        if (server.socket === ws) {
          delete connectedServers[id];
          log("Removed server from connected servers:", id);
          break;
        }
      }
    },
  },
});

log("Server is running at http://localhost:" + server.port);
