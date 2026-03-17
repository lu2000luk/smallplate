"use client";

import { useState } from "react";
import { assertManagerUrl } from "@/lib/utils";
import { Button } from "./ui/button";
import { Input } from "./ui/input";
import { Spinner } from "./ui/spinner";

type AuthResponse = {
  success?: boolean;
  message?: string;
  authKey?: string;
};

export function Login() {
  const [isLoading, setIsLoading] = useState(false);
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");

  const parseAuthResponse = async (res: Response): Promise<AuthResponse> => {
    const isJson = res.headers
      .get("content-type")
      ?.includes("application/json");
    if (!isJson) return {};
    try {
      return (await res.json()) as AuthResponse;
    } catch {
      return {};
    }
  };

  const resolveAuthKey = (res: Response, body: AuthResponse): string | null => {
    if (typeof body.authKey === "string" && body.authKey.length > 0) {
      return body.authKey;
    }

    const setToken = res.headers.get("Set-Token");
    if (setToken) return setToken;

    const authorization = res.headers.get("Authorization");
    if (authorization?.startsWith("Bearer ")) {
      return authorization.slice("Bearer ".length);
    }

    return null;
  };

  const doLogin = async (emailValue: string, passwordValue: string) => {
    const success = true;
    const baseUrl = assertManagerUrl();
    const res = await fetch(`${baseUrl}/auth/login`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ email: emailValue, password: passwordValue }),
    });

    const body = await parseAuthResponse(res);

    if (!res.ok || body.success === false) {
      throw new Error(body.message || "Login failed.");
    }

    const authKey = resolveAuthKey(res, body);
    if (!authKey) {
      throw new Error("Login succeeded but no auth key was returned.");
    }

    localStorage.setItem("authKey", authKey);
    return success;
  };

  const handleLogin = async () => {
    setIsLoading(true);
    try {
      await doLogin(email, password);
    } catch (e) {
      console.error("Login failed", e);
      alert("Login failed. Check console.");
    } finally {
      setIsLoading(false);
      location.reload();
    }
  };

  const handleRegister = async () => {
    setIsLoading(true);
    let success = true;
    try {
      const baseUrl = assertManagerUrl();

      const registerRes = await fetch(`${baseUrl}/auth/register`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ email, password }),
      });

      const registerBody = await parseAuthResponse(registerRes);
      if (!registerRes.ok || registerBody.success === false) {
        throw new Error(registerBody.message || "Registration failed.");
      }

      success = await doLogin(email, password);
    } catch (e) {
      console.error("Registration failed", e);
      alert("Registration failed. Check console.");
    } finally {
      setIsLoading(false);
      if (success) {
        location.reload();
      }
    }
  };

  return (
    <div className="absolute top-1/2 left-1/2 -translate-1/2">
      <div className="grid grid-cols-2 gap-2 w-xs">
        <p className="col-span-2 text-5xl mb-4 font-semibold text-center select-none">
          smallplate.
        </p>
        <Input
          type="email"
          className="col-span-2"
          placeholder="Email..."
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          disabled={isLoading}
        />
        <Input
          type="password"
          className="col-span-2"
          placeholder="Password..."
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          disabled={isLoading}
        />
        {isLoading ? (
          <Button variant="outline" disabled className="col-span-2">
            <Spinner />
          </Button>
        ) : (
          <>
            <Button variant="outline" onClick={handleLogin}>
              Login
            </Button>
            <Button variant="outline" onClick={handleRegister}>
              Register
            </Button>
          </>
        )}
      </div>
    </div>
  );
}
