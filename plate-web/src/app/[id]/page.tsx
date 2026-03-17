"use client";

import { useParams } from "next/navigation";
import { useEffect, useMemo, useState } from "react";

import { Login } from "@/components/login";
import { PlateDashboard } from "@/components/plate-dashboard";

export default function PlatePage() {
  const params = useParams<{ id: string }>();
  const [loggedIn, setLoggedIn] = useState(false);

  useEffect(() => {
    const authKey = localStorage.getItem("authKey");
    if (authKey && authKey !== "") {
      setLoggedIn(true);
    }
  }, []);

  const plateId = useMemo(() => Number(params.id), [params.id]);

  if (!loggedIn) {
    return <Login />;
  }

  if (!Number.isInteger(plateId) || plateId <= 0) {
    return (
      <div className="flex min-h-screen items-center justify-center px-4">
        <p className="text-sm text-muted-foreground">Invalid plate id.</p>
      </div>
    );
  }

  return <PlateDashboard plateId={plateId} />;
}
