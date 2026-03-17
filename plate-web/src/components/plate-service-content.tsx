"use client";

import { ServerIcon } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Field, FieldDescription, FieldLabel } from "@/components/ui/field";
import { Spinner } from "@/components/ui/spinner";
import type { Plate, ServiceDefinition, ServiceType } from "./plate-dashboard";

type ServiceContentProps = {
  plate: Plate;
  serviceType: ServiceType;
  serviceDefinition: ServiceDefinition;
  isPending: boolean;
  error: string | null;
  onEnable: () => void;
  onDisable: () => void;
};

export function PlateServiceContent({
  plate,
  serviceType,
  serviceDefinition,
  isPending,
  error,
  onEnable,
  onDisable,
}: ServiceContentProps) {
  const isEnabled = (() => {
    const data =
      typeof plate.data === "object" && plate.data !== null ? plate.data : {};
    const services = data.enabled_services;

    if (!Array.isArray(services)) {
      return false;
    }

    return services.includes(serviceType);
  })();

  const getAssignedServerId = (): string | null => {
    const servers =
      typeof plate.servers === "object" && plate.servers !== null
        ? plate.servers
        : {};
    const serverId = servers[serviceType];
    return typeof serverId === "string" && serverId.length > 0
      ? serverId
      : null;
  };

  return (
    <div className="space-y-4">
      <div className="flex items-start justify-between gap-4">
        <div>
          <p className="text-sm text-muted-foreground">Service</p>
          <h2 className="text-2xl font-semibold">{serviceDefinition.label}</h2>
        </div>
        <Badge variant={isEnabled ? "success" : "outline"}>
          {isEnabled ? "Enabled" : "Disabled"}
        </Badge>
      </div>

      {isEnabled ? (
        <div className="space-y-6 max-w-2xl">
          <div className="space-y-4">
            <div className="space-y-2">
              <div className="text-sm font-medium">Assigned server</div>
              <div className="text-sm text-muted-foreground">
                The manager routes this service to the following server.
              </div>
              <div className="inline-flex items-center gap-2 rounded-lg border bg-muted/30 px-3 py-2 font-mono text-sm">
                <ServerIcon className="size-4" />
                <span>{getAssignedServerId() ?? "No server assigned"}</span>
              </div>
            </div>

            <div className="pt-4">
              <Button
                variant="outline"
                onClick={onDisable}
                disabled={isPending}
              >
                {isPending ? <Spinner /> : "Disable"}
              </Button>
            </div>
          </div>
        </div>
      ) : (
        <div className="space-y-4 max-w-2xl">
          <div className="space-y-1">
            <h3 className="text-lg font-medium">Service disabled</h3>
            <p className="text-sm text-muted-foreground">
              Enable {serviceDefinition.label.toLowerCase()} to see its config
              and assigned server.
            </p>
          </div>
          <Button onClick={onEnable} disabled={isPending}>
            {isPending ? <Spinner /> : `Enable ${serviceDefinition.label}`}
          </Button>
        </div>
      )}

      {error ? (
        <p className="text-sm text-destructive-foreground">{error}</p>
      ) : null}
    </div>
  );
}
