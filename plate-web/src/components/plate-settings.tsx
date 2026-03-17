"use client";

import { Trash2Icon } from "lucide-react";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Spinner } from "@/components/ui/spinner";
import type { Plate, ServiceDefinition } from "./plate-dashboard";

type PlateSettingsProps = {
  plate: Plate;
  serviceDefinitions: ServiceDefinition[];
  error: string | null;
  isDeleting: boolean;
  onDeletePlate: () => void;
};

function getEnabledServices(plate: Plate): string[] {
  const data =
    typeof plate.data === "object" && plate.data !== null ? plate.data : {};
  const services = data.enabled_services;

  if (!Array.isArray(services)) {
    return [];
  }

  return services.filter(
    (service): service is string => service === "db" || service === "kv",
  );
}

function getAssignedServerId(plate: Plate, service: string): string | null {
  const servers =
    typeof plate.servers === "object" && plate.servers !== null
      ? plate.servers
      : {};
  const serverId = servers[service];
  return typeof serverId === "string" && serverId.length > 0 ? serverId : null;
}

export function PlateSettings({
  plate,
  serviceDefinitions,
  error,
  isDeleting,
  onDeletePlate,
}: PlateSettingsProps) {
  return (
    <div className="space-y-4">
      <div>
        <p className="text-sm text-muted-foreground">Manage</p>
        <h2 className="text-2xl font-semibold">Settings</h2>
      </div>

      <div className="space-y-4 max-w-2xl">
        <h3 className="text-lg font-medium">Plate details</h3>
        <div className="space-y-3 text-sm text-muted-foreground">
          <div>
            <span className="font-medium text-foreground">Plate ID:</span>{" "}
            <span className="font-mono">{plate.id}</span>
          </div>
          <div>
            <span className="font-medium text-foreground">
              Enabled services:
            </span>{" "}
            {getEnabledServices(plate).length > 0
              ? getEnabledServices(plate).join(", ")
              : "none"}
          </div>
          <div className="space-y-2 pt-2">
            {serviceDefinitions.map((service) => (
              <div
                key={service.type}
                className="flex items-center justify-between gap-3 rounded-lg border px-3 py-2 bg-muted/10"
              >
                <span>{service.label}</span>
                <span className="font-mono text-xs">
                  {getAssignedServerId(plate, service.type) ?? "unassigned"}
                </span>
              </div>
            ))}
          </div>
        </div>
      </div>

      <div className="space-y-4 max-w-2xl pt-8 border-t">
        <h3 className="text-lg font-medium text-destructive">Danger zone</h3>
        <p className="text-sm text-muted-foreground">
          Deleting a plate removes it and its API keys permanently. This action
          cannot be undone.
        </p>

        {error ? (
          <p className="text-sm text-destructive-foreground">{error}</p>
        ) : null}

        <Button
          variant="destructive"
          onClick={onDeletePlate}
          disabled={isDeleting}
        >
          {isDeleting ? <Spinner /> : <Trash2Icon className="size-4" />}
          Delete plate
        </Button>
      </div>
    </div>
  );
}
