import { db } from ".";
import bcrypt from "bcrypt";
import { connectedServers } from "./index";

export type JsonObject = Record<string, unknown>;

export type UserRecord = {
  id: number;
  email: string;
  password: string;
  data: string | null;
};

export type UserAuthKeyRecord = {
  id: number;
  user_id: number;
  auth_key: string;
  created_at: string;
};

const JSON_HEADERS = {
  "Content-Type": "application/json",
};

const EMAIL_MAX_LENGTH = parseInt(process.env.EMAIL_MAX_LENGTH ?? "254", 10);
const PASSWORD_MIN_LENGTH = parseInt(
  process.env.PASSWORD_MIN_LENGTH ?? "8",
  10,
);
const PASSWORD_MAX_LENGTH = parseInt(
  process.env.PASSWORD_MAX_LENGTH ?? "72",
  10,
);
const BCRYPT_COST = parseInt(process.env.BCRYPT_COST ?? "12", 10);
const AUTH_KEY_LENGTH = parseInt(process.env.AUTH_KEY_LENGTH ?? "64", 10);
const AUTH_HEADER_NAMES = ["authorization", "x-auth"];

export function jsonResponse(
  body: JsonObject,
  status = 200,
  headers = {},
): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: {
      "Access-Control-Allow-Origin": "*",
      "Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
      "Access-Control-Allow-Headers": "Content-Type, Authorization",
      ...JSON_HEADERS,
      ...headers,
    },
  });
}

export function getClientIp(req: Request): string {
  const forwardedFor = req.headers.get("x-forwarded-for");
  if (forwardedFor) {
    const firstIp = forwardedFor.split(",")[0]?.trim();
    if (firstIp) return firstIp;
  }

  const realIp = req.headers.get("x-real-ip")?.trim();
  if (realIp) return realIp;

  return "unknown";
}

export function normalizeEmail(email: unknown): string {
  return typeof email === "string" ? email.trim().toLowerCase() : "";
}

export function isValidEmail(email: string): boolean {
  if (!email || email.length > EMAIL_MAX_LENGTH) return false;
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
}

export function isValidPassword(password: unknown): password is string {
  if (typeof password !== "string") return false;
  if (password.length < PASSWORD_MIN_LENGTH) return false;
  if (password.length > PASSWORD_MAX_LENGTH) return false;
  return true;
}

export async function parseJsonBody(req: Request): Promise<JsonObject | null> {
  const contentType = req.headers.get("content-type") ?? "";
  if (!contentType.toLowerCase().includes("application/json")) {
    return null;
  }

  try {
    const body = await req.json();
    if (!body || typeof body !== "object" || Array.isArray(body)) {
      return null;
    }

    return body as JsonObject;
  } catch {
    return null;
  }
}

export async function hashPassword(password: string): Promise<string> {
  const salt = await bcrypt.genSalt(BCRYPT_COST);
  return bcrypt.hash(password, salt);
}

export async function verifyPassword(
  password: string,
  hash: string,
): Promise<boolean> {
  return bcrypt.compare(password, hash);
}

export function getUserByEmail(email: string): UserRecord | null {
  const statement = db.query(
    "SELECT id, email, password, data FROM users WHERE email = ? LIMIT 1",
  ) as {
    get(email: string): UserRecord | null;
  };

  return statement.get(email) ?? null;
}

export function getUserById(userId: number): UserRecord | null {
  const statement = db.query(
    "SELECT id, email, password, data FROM users WHERE id = ? LIMIT 1",
  ) as {
    get(userId: number): UserRecord | null;
  };

  return statement.get(userId) ?? null;
}

export function createUser(email: string, passwordHash: string) {
  const statement = db.query(
    "INSERT INTO users (email, password, data) VALUES (?, ?, ?)",
  ) as {
    run(email: string, passwordHash: string, data: string): unknown;
  };

  statement.run(email, passwordHash, JSON.stringify({ createdAt: Date.now() }));
}

export function sanitizeUser(user: {
  id: number;
  email: string;
  data: string | null;
}) {
  let parsedData: unknown = null;

  if (user.data) {
    try {
      parsedData = JSON.parse(user.data);
    } catch {
      parsedData = null;
    }
  }

  return {
    id: user.id,
    email: user.email,
    data: parsedData,
  };
}

function randomString(length: number): string {
  const bytes = new Uint8Array(length);
  crypto.getRandomValues(bytes);

  let result = "";
  for (const byte of bytes) {
    result += (byte % 36).toString(36);
  }

  return result;
}

export function generateAuthKey(length = AUTH_KEY_LENGTH): string {
  return randomString(length);
}

export function createUserAuthKey(userId: number): string {
  const authKey = generateAuthKey();
  const statement = db.query(
    "INSERT INTO user_auth_keys (user_id, auth_key) VALUES (?, ?)",
  ) as {
    run(userId: number, authKey: string): unknown;
  };

  statement.run(userId, authKey);
  return authKey;
}

export function getUserAuthKey(authKey: string): UserAuthKeyRecord | null {
  const statement = db.query(
    "SELECT id, user_id, auth_key, created_at FROM user_auth_keys WHERE auth_key = ? LIMIT 1",
  ) as {
    get(authKey: string): UserAuthKeyRecord | null;
  };

  return statement.get(authKey) ?? null;
}

export function getUserByAuthKey(authKey: string): UserRecord | null {
  const statement = db.query(`
    SELECT users.id, users.email, users.password, users.data
    FROM user_auth_keys
    INNER JOIN users ON users.id = user_auth_keys.user_id
    WHERE user_auth_keys.auth_key = ?
    LIMIT 1
  `) as {
    get(authKey: string): UserRecord | null;
  };

  return statement.get(authKey) ?? null;
}

export function revokeUserAuthKey(authKey: string) {
  const statement = db.query(
    "DELETE FROM user_auth_keys WHERE auth_key = ?",
  ) as {
    run(authKey: string): unknown;
  };

  statement.run(authKey);
}

export function revokeUserAuthKeysByUserId(userId: number) {
  const statement = db.query(
    "DELETE FROM user_auth_keys WHERE user_id = ?",
  ) as {
    run(userId: number): unknown;
  };

  statement.run(userId);
}

export function extractAuthKey(req: Request): string | null {
  for (const headerName of AUTH_HEADER_NAMES) {
    const value = req.headers.get(headerName)?.trim();
    if (!value) continue;

    if (headerName === "authorization") {
      const bearerPrefix = "Bearer ";
      if (value.startsWith(bearerPrefix)) {
        const token = value.slice(bearerPrefix.length).trim();
        if (token) return token;
      }

      if (value) return value;
      continue;
    }

    return value;
  }

  return null;
}

export function getAuthenticatedUser(req: Request): UserRecord | null {
  const authKey = extractAuthKey(req);
  if (!authKey) return null;
  return getUserByAuthKey(authKey);
}

export function requireLogin(
  req: Request,
):
  | { ok: true; user: UserRecord; authKey: string }
  | { ok: false; response: Response } {
  const authKey = extractAuthKey(req);

  if (!authKey) {
    return {
      ok: false,
      response: jsonResponse(
        {
          success: false,
          message: "Authentication required.",
        },
        401,
      ),
    };
  }

  const user = getUserByAuthKey(authKey);
  if (!user) {
    return {
      ok: false,
      response: jsonResponse(
        {
          success: false,
          message: "Invalid auth key.",
        },
        401,
      ),
    };
  }

  return {
    ok: true,
    user,
    authKey,
  };
}

export function getRandomServerId(length = 16): string {
  const chars =
    "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";
  let result = "";
  do {
    result = "";
    for (let i = 0; i < length; i++) {
      result += chars.charAt(Math.floor(Math.random() * chars.length));
    }
  } while (result in connectedServers);
  return result;
}
