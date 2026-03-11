import { db } from ".";
import { connectedServers } from "./index";

export type PlateRecord = {
  id: number;
  user_id: number;
  name: string;
  servers: string | null;
  data: string | null;
};

export type PlateApiKeyRecord = {
  id: number;
  plate_id: number;
  api_key: string;
  created_at: number;
};

export type PlateDataObject = Record<string, unknown>;
export type PlateServersObject = Partial<Record<"db" | "kv", string>>;

function safeParseJson<T>(value: string | null, fallback: T): T {
  if (!value) return fallback;

  try {
    return JSON.parse(value) as T;
  } catch {
    return fallback;
  }
}

function createRandomToken(length = 48): string {
  const chars =
    "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";
  let result = "";

  for (let i = 0; i < length; i++) {
    result += chars.charAt(Math.floor(Math.random() * chars.length));
  }

  return result;
}

export function createApiKey(plateId: number, key?: string): PlateApiKeyRecord {
  const apiKey = key ?? createRandomToken(64);
  const createdAt = Date.now();

  const statement = db.query(
    "INSERT INTO api_keys (plate_id, api_key, created_at) VALUES (?, ?, ?)",
  ) as {
    run(
      plateId: number,
      apiKey: string,
      createdAt: number,
    ): {
      lastInsertRowid: number | bigint;
    };
  };

  const result = statement.run(plateId, apiKey, createdAt);

  return {
    id: Number(result.lastInsertRowid),
    plate_id: plateId,
    api_key: apiKey,
    created_at: createdAt,
  };
}

export function listApiKeysByPlate(plateId: number): PlateApiKeyRecord[] {
  const statement = db.query(
    "SELECT id, plate_id, api_key, created_at FROM api_keys WHERE plate_id = ? ORDER BY created_at DESC, id DESC",
  ) as {
    all(plateId: number): PlateApiKeyRecord[];
  };

  return statement.all(plateId);
}

export function deleteApiKeyById(id: number): boolean {
  const statement = db.query("DELETE FROM api_keys WHERE id = ?") as {
    run(id: number): { changes: number };
  };

  const result = statement.run(id);
  return result.changes > 0;
}

export function createPlate(userId: number, name: string): PlateRecord {
  const statement = db.query(
    "INSERT INTO plates (user_id, name, servers, data) VALUES (?, ?, ?, ?)",
  ) as {
    run(
      userId: number,
      name: string,
      servers: string,
      data: string,
    ): { lastInsertRowid: number | bigint };
  };

  const servers = JSON.stringify({});
  const data = JSON.stringify({ enabled_services: [] });
  const result = statement.run(userId, name, servers, data);

  return {
    id: Number(result.lastInsertRowid),
    user_id: userId,
    name,
    servers,
    data,
  };
}

export function listUserPlates(userId: number): PlateRecord[] {
  const statement = db.query(
    "SELECT id, user_id, name, servers, data FROM plates WHERE user_id = ? ORDER BY id DESC",
  ) as {
    all(userId: number): PlateRecord[];
  };

  return statement.all(userId);
}

export function deletePlate(id: number): boolean {
  const statement = db.query("DELETE FROM plates WHERE id = ?") as {
    run(id: number): { changes: number };
  };

  const result = statement.run(id);
  return result.changes > 0;
}

export function setPlateDataObject(
  plateId: number,
  data: PlateDataObject,
): boolean {
  const statement = db.query("UPDATE plates SET data = ? WHERE id = ?") as {
    run(data: string, plateId: number): { changes: number };
  };

  const result = statement.run(JSON.stringify(data ?? {}), plateId);
  return result.changes > 0;
}

export function getPlateDataObject(plateId: number): PlateDataObject | null {
  const statement = db.query(
    "SELECT data FROM plates WHERE id = ? LIMIT 1",
  ) as {
    get(plateId: number): { data: string | null } | null;
  };

  const row = statement.get(plateId);
  if (!row) return null;

  return safeParseJson<PlateDataObject>(row.data, {});
}

export function getPlateServers(plateId: number): PlateServersObject | null {
  const statement = db.query(
    "SELECT servers FROM plates WHERE id = ? LIMIT 1",
  ) as {
    get(plateId: number): { servers: string | null } | null;
  };

  const row = statement.get(plateId);
  if (!row) return null;

  const parsed = safeParseJson<unknown>(row.servers, {});
  if (typeof parsed !== "object" || parsed === null || Array.isArray(parsed)) {
    return {};
  }

  const servers: PlateServersObject = {};

  for (const [key, value] of Object.entries(parsed)) {
    if ((key === "db" || key === "kv") && typeof value === "string") {
      servers[key] = value;
    }
  }

  return servers;
}

export function setPlateServers(
  plateId: number,
  servers: PlateServersObject,
): boolean {
  const statement = db.query("UPDATE plates SET servers = ? WHERE id = ?") as {
    run(servers: string, plateId: number): { changes: number };
  };

  const result = statement.run(JSON.stringify(servers ?? {}), plateId);
  return result.changes > 0;
}

export function getPlateById(id: number): PlateRecord | null {
  const statement = db.query(
    "SELECT id, user_id, name, servers, data FROM plates WHERE id = ? LIMIT 1",
  ) as {
    get(id: number): PlateRecord | null;
  };

  return statement.get(id) ?? null;
}

export function plateBelongsToUser(plateId: number, userId: number): boolean {
  const statement = db.query(
    "SELECT id FROM plates WHERE id = ? AND user_id = ? LIMIT 1",
  ) as {
    get(plateId: number, userId: number): { id: number } | null;
  };

  return statement.get(plateId, userId) !== null;
}

export function listConnectedServersForPlate(plateId: number): Array<{
  id: string;
  type: "db" | "kv";
  latency: number;
}> {
  const serverMap = getPlateServers(plateId) ?? {};
  const connectedPlateServers: Array<{
    id: string;
    type: "db" | "kv";
    latency: number;
  }> = [];

  for (const [type, serverId] of Object.entries(serverMap)) {
    if ((type !== "db" && type !== "kv") || typeof serverId !== "string") {
      continue;
    }

    const server = connectedServers[serverId];
    if (!server) continue;

    connectedPlateServers.push({
      id: serverId,
      type,
      latency: server.latency,
    });
  }

  return connectedPlateServers;
}

export function listAllPlates(): PlateRecord[] {
  const statement = db.query(
    "SELECT id, user_id, name, servers, data FROM plates ORDER BY id DESC",
  ) as {
    all(): PlateRecord[];
  };

  return statement.all();
}
