"use client";

import { useEffect, useState } from "react";
import {
  ServerIcon,
  CopyIcon,
  LinkIcon,
  Link,
  ExternalLink,
} from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import { assertManagerUrl } from "@/lib/utils";
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
  const [serverUrl, setServerUrl] = useState<string | null>(null);
  const [isLoadingUrl, setIsLoadingUrl] = useState(false);
  const [urlFetchError, setUrlFetchError] = useState<string | null>(null);
  const [copiedUrl, setCopiedUrl] = useState(false);

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

  const assignedServerId = getAssignedServerId();

  useEffect(() => {
    if (isEnabled && assignedServerId) {
      const fetchUrl = async () => {
        setIsLoadingUrl(true);
        setUrlFetchError(null);
        try {
          const managerUrl = assertManagerUrl();
          const response = await fetch(
            `${managerUrl}/services/url?id=${assignedServerId}`,
          );
          const data = await response.json();
          if (data.success && typeof data.url === "string") {
            setServerUrl(data.url);
          } else {
            setUrlFetchError(data.message || "Failed to fetch server URL.");
          }
        } catch (err) {
          setUrlFetchError("Network error while fetching server URL.");
        } finally {
          setIsLoadingUrl(false);
        }
      };

      fetchUrl();
    } else {
      setServerUrl(null);
    }
  }, [isEnabled, assignedServerId]);

  const handleCopyUrl = () => {
    if (serverUrl) {
      navigator.clipboard.writeText(serverUrl);
      setCopiedUrl(true);
      setTimeout(() => setCopiedUrl(false), 2000);
    }
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
                <span>{assignedServerId ?? "No server assigned"}</span>
              </div>
            </div>

            {assignedServerId && (
              <div className="space-y-2">
                <div className="text-sm font-medium">Server URL</div>
                <div className="text-sm text-muted-foreground">
                  The public connection URL for this server.
                </div>
                {isLoadingUrl ? (
                  <div className="flex items-center gap-2">
                    <Spinner />
                    <span className="text-sm text-muted-foreground">
                      Loading URL...
                    </span>
                  </div>
                ) : urlFetchError ? (
                  <div className="text-sm text-destructive-foreground">
                    {urlFetchError}
                  </div>
                ) : serverUrl ? (
                  <div className="flex items-center gap-2">
                    <div className="inline-flex items-center gap-2 rounded-lg border bg-muted/30 px-3 py-2 font-mono text-sm break-all">
                      <LinkIcon className="size-4" />
                      <span>{serverUrl}</span>
                    </div>
                    <Button
                      variant="outline"
                      size="icon"
                      onClick={handleCopyUrl}
                      title="Copy URL"
                    >
                      <CopyIcon className="size-4" />
                      <span className="sr-only">Copy URL</span>
                    </Button>
                    {copiedUrl && (
                      <span className="text-xs text-success font-medium">
                        Copied!
                      </span>
                    )}
                  </div>
                ) : (
                  <div className="text-sm text-muted-foreground">
                    No URL available
                  </div>
                )}
              </div>
            )}

            <div className="pt-1 gap-2 flex">
              {process.env.NEXT_PUBLIC_DOCS_URL && (
                <Button
                  variant="outline"
                  onClick={() => {
                    const docs_url =
                      process.env.NEXT_PUBLIC_DOCS_URL + "/" + serviceType;
                    window.open(docs_url, "_blank");
                  }}
                >
                  <ExternalLink /> Open docs
                </Button>
              )}
              <Button
                variant="destructive-outline"
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
