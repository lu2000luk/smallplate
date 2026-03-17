"use client";

import { ArrowRight } from "lucide-react";
import Link from "next/link";
import { useCallback, useEffect, useMemo, useState } from "react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Dialog,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogPanel,
  DialogPopup,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Spinner } from "@/components/ui/spinner";
import { assertManagerUrl, authenticatedFetch } from "@/lib/utils";

type Plate = {
  id: number;
  user_id: number;
  name: string;
  servers: Record<string, unknown>;
  data: Record<string, unknown>;
};

type PlatesResponse = {
  success?: boolean;
  message?: string;
  plates?: Plate[];
};

type CreatePlateResponse = {
  success?: boolean;
  message?: string;
  plate?: Plate;
};

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function getEnabledServicesCount(plate: Plate): number {
  const data = isRecord(plate.data) ? plate.data : {};
  const services = data.enabled_services;

  if (!Array.isArray(services)) {
    return 0;
  }

  return services.filter((service) => service === "db" || service === "kv")
    .length;
}

function getErrorMessage(body: unknown, fallback: string): string {
  if (isRecord(body) && typeof body.message === "string" && body.message) {
    return body.message;
  }

  return fallback;
}

export function Home() {
  const [plates, setPlates] = useState<Plate[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [isCreating, setIsCreating] = useState(false);
  const [createError, setCreateError] = useState<string | null>(null);
  const [plateName, setPlateName] = useState("");

  const baseUrl = useMemo(() => assertManagerUrl(), []);

  const requestJson = useCallback(
    async <T,>(path: string, init?: RequestInit, fallbackMessage?: string) => {
      const response = await authenticatedFetch(`${baseUrl}${path}`, init);

      let body: unknown = null;
      try {
        body = await response.json();
      } catch {
        body = null;
      }

      if (!response.ok || (isRecord(body) && body.success === false)) {
        throw new Error(
          getErrorMessage(body, fallbackMessage || "Request failed."),
        );
      }

      return body as T;
    },
    [baseUrl],
  );

  const loadPlates = useCallback(async () => {
    setIsLoading(true);
    setError(null);

    try {
      const body = await requestJson<PlatesResponse>(
        "/plates/list",
        undefined,
        "Failed to load plates.",
      );

      setPlates(Array.isArray(body.plates) ? body.plates : []);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load plates.");
    } finally {
      setIsLoading(false);
    }
  }, [requestJson]);

  useEffect(() => {
    void loadPlates();
  }, [loadPlates]);

  const resetCreateDialog = () => {
    setIsCreateOpen(false);
    setCreateError(null);
    setPlateName("");
  };

  const handleCreatePlate = async () => {
    const name = plateName.trim();
    if (!name) {
      setCreateError("Plate name is required.");
      return;
    }

    setIsCreating(true);
    setCreateError(null);

    try {
      await requestJson<CreatePlateResponse>(
        "/plates/create",
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({ name }),
        },
        "Failed to create plate.",
      );

      resetCreateDialog();
      await loadPlates();
    } catch (err) {
      setCreateError(
        err instanceof Error ? err.message : "Failed to create plate.",
      );
    } finally {
      setIsCreating(false);
    }
  };

  return (
    <div className="mx-auto flex min-h-screen w-full max-w-4xl flex-col gap-6 px-4 py-10 sm:px-6">
      <div className="flex items-end justify-between gap-4 border-b pb-4">
        <div>
          <p className="text-sm text-muted-foreground">Dashboard</p>
          <h1 className="text-3xl font-semibold tracking-tight">Your plates</h1>
        </div>

        <Dialog
          onOpenChange={(open) => {
            if (!open) {
              resetCreateDialog();
              return;
            }

            setIsCreateOpen(true);
          }}
          open={isCreateOpen}
        >
          <DialogTrigger render={<Button variant="outline" />}>
            Create plate
          </DialogTrigger>
          <DialogPopup>
            <DialogHeader>
              <DialogTitle>Create plate</DialogTitle>
              <DialogDescription>
                Add a new plate and open its dashboard when you are ready.
              </DialogDescription>
            </DialogHeader>
            <DialogPanel>
              <div className="space-y-3">
                <Input
                  placeholder="Plate name"
                  value={plateName}
                  onChange={(event) => setPlateName(event.target.value)}
                  disabled={isCreating}
                  onKeyDown={(event) => {
                    if (event.key === "Enter") {
                      event.preventDefault();
                      void handleCreatePlate();
                    }
                  }}
                />
                {createError ? (
                  <p className="text-sm text-destructive-foreground">
                    {createError}
                  </p>
                ) : null}
              </div>
            </DialogPanel>
            <DialogFooter>
              <Button
                variant="outline"
                onClick={resetCreateDialog}
                disabled={isCreating}
              >
                Cancel
              </Button>
              <Button onClick={handleCreatePlate} disabled={isCreating}>
                {isCreating ? <Spinner /> : "Create plate"}
              </Button>
            </DialogFooter>
          </DialogPopup>
        </Dialog>
      </div>

      {isLoading ? (
        <div className="flex min-h-40 items-center justify-center rounded-2xl border">
          <Spinner />
        </div>
      ) : error ? (
        <Card>
          <CardHeader>
            <CardTitle>Could not load plates</CardTitle>
          </CardHeader>
          <CardContent className="flex items-center justify-between gap-4">
            <p className="text-sm text-muted-foreground">{error}</p>
            <Button variant="outline" onClick={() => void loadPlates()}>
              Retry
            </Button>
          </CardContent>
        </Card>
      ) : plates.length === 0 ? (
        <Card>
          <CardHeader>
            <CardTitle>No plates yet</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">
              Create your first plate to open its dedicated dashboard.
            </p>
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {plates.map((plate) => {
            const enabledCount = getEnabledServicesCount(plate);

            return (
              <Link href={`/${plate.id}`} key={plate.id}>
                <Card className="transition-colors hover:bg-muted/20 hover:border-white/20">
                  <CardHeader className="flex flex-row items-start justify-between gap-4 space-y-0">
                    <div>
                      <CardTitle>{plate.name}</CardTitle>
                      <div className="mt-2 flex gap-1">
                        <Badge variant={"info"}>#{plate.id}</Badge>
                        <Badge variant="secondary">
                          {enabledCount} service{enabledCount !== 1 ? "s" : ""}
                        </Badge>
                      </div>
                    </div>
                    <ArrowRight className="h-4 w-4 text-muted-foreground" />
                  </CardHeader>
                </Card>
              </Link>
            );
          })}
        </div>
      )}
    </div>
  );
}
