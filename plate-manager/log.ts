function getTimestampAndFunctionName(): string {
  const now = new Date();
  const timestamp = `${String(now.getHours()).padStart(2, "0")}:${String(now.getMinutes()).padStart(2, "0")}:${String(now.getSeconds()).padStart(2, "0")}.${String(now.getMilliseconds()).padStart(3, "0")}`;
  const stack = new Error().stack;
  if (stack) {
    const lines = stack.split("\n");
    if (true) {
      // @ts-ignore
      const callerLine = lines[3].trim().split("\\").at(-1) || undefined;
      if (callerLine) {
        return `\x1b[90m${timestamp} <\x1b[0;38;2;104;173;0;49m${callerLine
          .replaceAll(")", "")
          .replaceAll("(", "")}\x1b[90m>\x1b[0m`;
      }
    }
  }
  return `[${timestamp}]`;
}

export function log(...args: any[]) {
  console.log(
    "\x1b[34m[plate::manager]\x1b[0m",
    getTimestampAndFunctionName(),
    ...args,
  );
}

export function error(...args: any[]) {
  console.error(
    "\x1b[31m[plate::manager]\x1b[0m " + getTimestampAndFunctionName(),
    ...args,
  );
}

export function warn(...args: any[]) {
  console.warn(
    "\x1b[33m[plate::manager]\x1b[0m " + getTimestampAndFunctionName(),
    ...args,
  );
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
