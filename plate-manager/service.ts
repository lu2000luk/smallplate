import type { ServerWebSocket } from "bun";
import { error, log } from "./log";
import type { ServerWSData } from ".";

type ManagedServerWebSocket = ServerWebSocket<ServerWSData>;
type PingTimer = ReturnType<typeof setTimeout>;
type PingMessage = {
  type: "ping" | "pong";
  time?: number;
};

function getPrivateServiceKey() {
  let key = process.env.SERVICE_KEY;
  if (!key) {
    let reccomendedKey = "";
    const chars =
      "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";
    for (let i = 0; i < 64; i++) {
      reccomendedKey += chars.charAt(Math.floor(Math.random() * chars.length));
    }
    error(
      "SERVICE_KEY is not set in environment variables. It is necessary for connecting servers together. Please set it to a secure random string and configure it to other servers aswell. Random reccomended key: " +
        reccomendedKey,
    );
    process.exit(1);
  }
  return key;
}

export const SERVICE_KEY = getPrivateServiceKey();

const pingTimeouts = new Map<ManagedServerWebSocket, PingTimer>();
const pingIntervals = new Map<ManagedServerWebSocket, PingTimer>();

function setPingTimeout(ws: ManagedServerWebSocket) {
  const existing = pingTimeouts.get(ws);
  if (existing) clearTimeout(existing);

  const timeout = setTimeout(() => {
    log("Server did not respond to ping, disconnecting...");
    stopPingLoop(ws);
    ws.close();
  }, 60_000);

  pingTimeouts.set(ws, timeout);
}

function sendPing(ws: ManagedServerWebSocket) {
  if (ws.readyState !== WebSocket.OPEN) {
    stopPingLoop(ws);
    return;
  }

  ws.send(
    JSON.stringify({
      type: "ping" satisfies PingMessage["type"],
      time: Date.now(),
    }),
  );
  setPingTimeout(ws);
}

export function startPingLoop(ws: ManagedServerWebSocket) {
  stopPingLoop(ws);
  sendPing(ws);

  const interval = setInterval(() => {
    sendPing(ws);
  }, 30_000);

  pingIntervals.set(ws, interval);
}

export function stopPingLoop(ws: ManagedServerWebSocket) {
  const interval = pingIntervals.get(ws);
  if (interval) {
    clearInterval(interval);
    pingIntervals.delete(ws);
  }

  const timeout = pingTimeouts.get(ws);
  if (timeout) {
    clearTimeout(timeout);
    pingTimeouts.delete(ws);
  }
}

export function handlePingMessage(
  ws: ManagedServerWebSocket,
  message: PingMessage,
) {
  const latency =
    typeof message.time === "number"
      ? Math.max(0, Date.now() - message.time)
      : 100;

  if (message.type === "ping") {
    if (ws.readyState === WebSocket.OPEN) {
      ws.send(
        JSON.stringify({
          type: "pong" satisfies PingMessage["type"],
          time: message.time ?? Date.now(),
        }),
      );
    }
    return;
  }

  if (message.type === "pong") {
    ws.data.latency = latency;
    setPingTimeout(ws);
  }
}
