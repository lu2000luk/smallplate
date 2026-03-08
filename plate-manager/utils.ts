import { db } from ".";
import bcrypt from "bcrypt";
import { connectedServers } from "./index";

export type JsonObject = Record<string, unknown>;

const JSON_HEADERS = {
  "Content-Type": "application/json",
};

const EMAIL_MAX_LENGTH = 254;
const PASSWORD_MIN_LENGTH = 8;
const PASSWORD_MAX_LENGTH = 72;
const BCRYPT_COST = 12;

export function jsonResponse(
  body: JsonObject,
  status = 200,
  headers: HeadersInit = {},
): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: {
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

export function getUserByEmail(email: string): {
  id: number;
  email: string;
  password: string;
  data: string | null;
} | null {
  const statement = db.query(
    "SELECT id, email, password, data FROM users WHERE email = ? LIMIT 1",
  ) as {
    get(email: string): {
      id: number;
      email: string;
      password: string;
      data: string | null;
    } | null;
  };

  return statement.get(email) ?? null;
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
