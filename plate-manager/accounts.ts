import {
  createUser,
  createUserAuthKey,
  getClientIp,
  getUserByEmail,
  hashPassword,
  isValidEmail,
  isValidPassword,
  jsonResponse,
  normalizeEmail,
  parseJsonBody,
  sanitizeUser,
  verifyPassword,
} from "./utils";
import { log, warn } from "./log";

const REGISTER_WINDOW_MS =
  Number(process.env.REGISTER_WINDOW_MS) || 15 * 60 * 1000;
const REGISTER_MAX_ATTEMPTS_PER_IP =
  Number(process.env.REGISTER_MAX_ATTEMPTS_PER_IP) || 10;

const LOGIN_WINDOW_MS = Number(process.env.LOGIN_WINDOW_MS) || 15 * 60 * 1000;
const LOGIN_MAX_ATTEMPTS_PER_IP =
  Number(process.env.LOGIN_MAX_ATTEMPTS_PER_IP) || 15;
const LOGIN_MAX_ATTEMPTS_PER_EMAIL =
  Number(process.env.LOGIN_MAX_ATTEMPTS_PER_EMAIL) || 10;

type RateLimitEntry = {
  count: number;
  resetAt: number;
};

const registerRateLimitByIp = new Map<string, RateLimitEntry>();
const loginRateLimitByIp = new Map<string, RateLimitEntry>();
const loginRateLimitByEmail = new Map<string, RateLimitEntry>();

function cleanupRateLimitStore(
  store: Map<string, RateLimitEntry>,
  now: number,
) {
  for (const [key, entry] of store.entries()) {
    if (entry.resetAt <= now) {
      store.delete(key);
    }
  }
}

function checkRateLimit(
  store: Map<string, RateLimitEntry>,
  key: string,
  maxAttempts: number,
  windowMs: number,
): { allowed: boolean; retryAfterSeconds: number } {
  const now = Date.now();
  cleanupRateLimitStore(store, now);

  const existing = store.get(key);
  if (!existing || existing.resetAt <= now) {
    store.set(key, {
      count: 1,
      resetAt: now + windowMs,
    });
    return { allowed: true, retryAfterSeconds: 0 };
  }

  if (existing.count >= maxAttempts) {
    return {
      allowed: false,
      retryAfterSeconds: Math.max(
        1,
        Math.ceil((existing.resetAt - now) / 1000),
      ),
    };
  }

  existing.count += 1;
  store.set(key, existing);

  return { allowed: true, retryAfterSeconds: 0 };
}

function clearRateLimit(store: Map<string, RateLimitEntry>, key: string) {
  store.delete(key);
}

async function signup(req: Request): Promise<Response> {
  const ip = getClientIp(req);
  const rateLimit = checkRateLimit(
    registerRateLimitByIp,
    ip,
    REGISTER_MAX_ATTEMPTS_PER_IP,
    REGISTER_WINDOW_MS,
  );

  if (!rateLimit.allowed) {
    return jsonResponse(
      {
        success: false,
        message: "Too many registration attempts. Please try again later.",
      },
      429,
      {
        "Retry-After": String(rateLimit.retryAfterSeconds),
      },
    );
  }

  const body = await parseJsonBody(req);
  if (!body) {
    return jsonResponse(
      {
        success: false,
        message: "Request body must be valid JSON.",
      },
      400,
    );
  }

  const email = normalizeEmail(body.email);
  const password = body.password;

  if (!isValidEmail(email)) {
    return jsonResponse(
      {
        success: false,
        message: "A valid email address is required.",
      },
      400,
    );
  }

  if (!isValidPassword(password)) {
    return jsonResponse(
      {
        success: false,
        message: "Password must be a string between 8 and 72 characters long.",
      },
      400,
    );
  }

  const existingUser = getUserByEmail(email);
  if (existingUser) {
    return jsonResponse(
      {
        success: false,
        message: "An account with that email already exists.",
      },
      409,
    );
  }

  const passwordHash = await hashPassword(password);

  try {
    createUser(email, passwordHash);
  } catch (error) {
    warn("Failed to create user account.", error);

    const duplicateUser = getUserByEmail(email);
    if (duplicateUser) {
      return jsonResponse(
        {
          success: false,
          message: "An account with that email already exists.",
        },
        409,
      );
    }

    return jsonResponse(
      {
        success: false,
        message: "Failed to create account.",
      },
      500,
    );
  }

  const createdUser = getUserByEmail(email);
  if (!createdUser) {
    return jsonResponse(
      {
        success: false,
        message: "Account was created, but could not be loaded.",
      },
      500,
    );
  }

  log("Registered user", email);

  return jsonResponse(
    {
      success: true,
      message: "Registration successful.",
      user: sanitizeUser(createdUser),
    },
    201,
  );
}

async function login(req: Request): Promise<Response> {
  const ip = getClientIp(req);
  const ipRateLimit = checkRateLimit(
    loginRateLimitByIp,
    ip,
    LOGIN_MAX_ATTEMPTS_PER_IP,
    LOGIN_WINDOW_MS,
  );

  if (!ipRateLimit.allowed) {
    return jsonResponse(
      {
        success: false,
        message: "Too many login attempts. Please try again later.",
      },
      429,
      {
        "Retry-After": String(ipRateLimit.retryAfterSeconds),
      },
    );
  }

  const body = await parseJsonBody(req);
  if (!body) {
    return jsonResponse(
      {
        success: false,
        message: "Request body must be valid JSON.",
      },
      400,
    );
  }

  const email = normalizeEmail(body.email);
  const password = body.password;

  if (!isValidEmail(email) || typeof password !== "string") {
    return jsonResponse(
      {
        success: false,
        message: "Email and password are required.",
      },
      400,
    );
  }

  const emailRateLimit = checkRateLimit(
    loginRateLimitByEmail,
    email,
    LOGIN_MAX_ATTEMPTS_PER_EMAIL,
    LOGIN_WINDOW_MS,
  );

  if (!emailRateLimit.allowed) {
    return jsonResponse(
      {
        success: false,
        message:
          "Too many login attempts for this account. Please try again later.",
      },
      429,
      {
        "Retry-After": String(emailRateLimit.retryAfterSeconds),
      },
    );
  }

  const user = getUserByEmail(email);
  if (!user) {
    return jsonResponse(
      {
        success: false,
        message: "Invalid email or password.",
      },
      401,
    );
  }

  const passwordMatches = await verifyPassword(password, user.password);
  if (!passwordMatches) {
    return jsonResponse(
      {
        success: false,
        message: "Invalid email or password.",
      },
      401,
    );
  }

  clearRateLimit(loginRateLimitByEmail, email);
  clearRateLimit(loginRateLimitByIp, ip);

  let authKey: string;
  try {
    authKey = createUserAuthKey(user.id);
  } catch (error) {
    warn("Failed to create auth key for user.", error);
    return jsonResponse(
      {
        success: false,
        message: "Login succeeded but token creation failed.",
      },
      500,
    );
  }

  log("Logged in user", email);

  return jsonResponse(
    {
      success: true,
      message: "Login successful.",
      user: sanitizeUser(user),
      authKey,
    },
    200,
    {
      "Set-Token": authKey,
      Authorization: `Bearer ${authKey}`,
    },
  );
}

export function authRouter(
  req: Request,
  url: URL,
): Response | Promise<Response> {
  if (url.pathname === "/auth/login" && req.method === "POST") {
    return login(req);
  }

  if (url.pathname === "/auth/register" && req.method === "POST") {
    return signup(req);
  }

  return jsonResponse(
    {
      success: false,
      message: "[auth] Endpoint not found.",
    },
    404,
  );
}
