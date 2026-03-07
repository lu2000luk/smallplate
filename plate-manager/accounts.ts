export function authRouter(req: Request, url: URL): Response {
  if (url.pathname === "/auth/login" && req.method === "POST") {
    return new Response(
      JSON.stringify({
        success: false,
        message: "Login is not implemented yet.",
      }),
      {
        status: 200,
        headers: {
          "Content-Type": "application/json",
        },
      },
    );
  }

  if (url.pathname === "/auth/register" && req.method === "POST") {
    return new Response(
      JSON.stringify({
        success: false,
        message: "Registration is not implemented yet.",
      }),
      {
        status: 200,
        headers: {
          "Content-Type": "application/json",
        },
      },
    );
  }

  return new Response(
    JSON.stringify({
      success: false,
      message: "[auth] Endpoint not found.",
    }),
    {
      status: 404,
      headers: {
        "Content-Type": "application/json",
      },
    },
  );
}
