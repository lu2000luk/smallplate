import { type ClassValue, clsx } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function authenticatedFetch(input: RequestInfo, init?: RequestInit) {
  const authKey = localStorage.getItem("authKey");
  if (!authKey) {
    throw new Error("No auth key found in localStorage");
  }

  const headers = new Headers(init?.headers || {});
  headers.set("Authorization", `${authKey}`);

  return fetch(input, {
    ...init,
    headers,
  });
}

export const assertManagerUrl = () => {
  const raw =
    process.env.NEXT_PUBLIC_MANAGER_URL?.trim().replace(/\/+$/, "") ?? "";

  if (!raw) {
    throw new Error(
      "Missing NEXT_PUBLIC_MANAGER_URL. Set it in your plate-web env file.",
    );
  }
  return raw;
};
