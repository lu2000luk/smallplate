"use client";

import { useEffect, useState } from "react";

export default function Home() {
  const [loggedIn, setLoggedIn] = useState(false);

  useEffect(() => {
    if (localStorage.getItem("authKey")) {
      setLoggedIn(true);
    }
  }, []);

  return <div className=""></div>;
}
