"use client";

import { useEffect, useState } from "react";
import { Home as HomeView } from "@/components/home";
import { Login } from "@/components/login";

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

  return <div>{!loggedIn ? <Login /> : <HomeView />}</div>;
}
