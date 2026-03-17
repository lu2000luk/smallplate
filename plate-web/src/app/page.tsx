"use client";

import { Login } from "@/components/login";
import { useEffect, useState } from "react";

export default function Home() {
  const [loggedIn, setLoggedIn] = useState(false);

  useEffect(() => {
    if (
      localStorage.getItem("authKey") &&
      localStorage.getItem("authKey") !== ""
    ) {
      setLoggedIn(true);
    }
  }, []);

  return <div className="">{!loggedIn ? <Login /> : null}</div>;
}
