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
  headers.set("Authorization", `Bearer ${authKey}`);

  return fetch(input, {
    ...init,
    headers,
  });
}
