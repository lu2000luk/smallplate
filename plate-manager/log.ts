export function log(...args: any[]) {
  console.log("[plate::manager]", ...args);
}

export function error(...args: any[]) {
  console.error("[plate::manager] Error: ", ...args);
}

export function warn(...args: any[]) {
  console.warn("[plate::manager] Warning: ", ...args);
}

export function tryCatch<T>(fn: () => T, errorMessage: string): T | null {
  try {
    return fn();
  } catch (e) {
    error(errorMessage, e);
    return null;
  }
}

export async function tryCatchAsync<T>(
  fn: () => Promise<T>,
  errorMessage: string,
): Promise<T | null> {
  try {
    return await fn();
  } catch (e) {
    error(errorMessage, e);
    return null;
  }
}
