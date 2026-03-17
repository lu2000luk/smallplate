"use client";

import type { LucideIcon } from "lucide-react";
import {
  ArrowLeftIcon,
  DatabaseIcon,
  HardDriveIcon,
  KeyRoundIcon,
  SettingsIcon,
} from "lucide-react";
import Link from "next/link";
import { useCallback, useEffect, useMemo, useState } from "react";

import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarInset,
  SidebarMenu,
  SidebarMenuBadge,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarProvider,
} from "@/components/ui/sidebar";
import { Spinner } from "@/components/ui/spinner";
import { assertManagerUrl, authenticatedFetch } from "@/lib/utils";
import { PlateApiKeys } from "./plate-api-keys";
import { PlateServiceContent } from "./plate-service-content";
import { PlateSettings } from "./plate-settings";
import { Button } from "./ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "./ui/card";

export type ServiceType = "db" | "kv";
export type DashboardSection = ServiceType | "api-keys" | "settings";

export type Plate = {
  id: number;
  user_id: number;
  name: string;
  servers: Record<string, unknown>;
  data: Record<string, unknown>;
};

export type PlateApiKey = {
  id: number;
  plate_id: number;
  api_key: string;
  created_at: number;
};

type PlateResponse = {
  success?: boolean;
  message?: string;
  plate?: Plate;
};

type ApiKeysResponse = {
  success?: boolean;
  message?: string;
  api_keys?: PlateApiKey[];
};

type CreateApiKeyResponse = {
  success?: boolean;
  message?: string;
  api_key?: PlateApiKey;
};

type ServiceMutationResponse = {
  success?: boolean;
  message?: string;
  plateId?: number;
  service?: ServiceType;
  serverId?: string | null;
};

type DeleteResponse = {
  success?: boolean;
  message?: string;
};

export type ServiceDefinition = {
  type: ServiceType;
  label: string;
  description: string;
  icon: LucideIcon;
};

export const serviceDefinitions: ServiceDefinition[] = [
  {
    type: "db",
    label: "Database",
    description: "Structured storage for your plate data.",
    icon: DatabaseIcon,
  },
  {
    type: "kv",
    label: "Key-Value",
    description: "Fast key-value access for lightweight state.",
    icon: HardDriveIcon,
  },
];

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function getEnabledServices(plate: Plate): ServiceType[] {
  const data = isRecord(plate.data) ? plate.data : {};
  const services = data.enabled_services;

  if (!Array.isArray(services)) {
    return [];
  }

  return services.filter(
    (service): service is ServiceType => service === "db" || service === "kv",
  );
}

function isServiceEnabled(plate: Plate, service: ServiceType): boolean {
  return getEnabledServices(plate).includes(service);
}

function _getAssignedServerId(
  plate: Plate,
  service: ServiceType,
): string | null {
  const servers = isRecord(plate.servers) ? plate.servers : {};
  const serverId = servers[service];
  return typeof serverId === "string" && serverId.length > 0 ? serverId : null;
}

function getErrorMessage(body: unknown, fallback: string): string {
  if (isRecord(body) && typeof body.message === "string" && body.message) {
    return body.message;
  }

  return fallback;
}

function _formatCreatedAt(timestamp: number): string {
  return new Date(timestamp).toLocaleString();
}

export function PlateDashboard({ plateId }: { plateId: number }) {
  const [plate, setPlate] = useState<Plate | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [activeSection, setActiveSection] = useState<DashboardSection>("db");

  const [servicePending, setServicePending] = useState<ServiceType | null>(
    null,
  );
  const [serviceErrors, setServiceErrors] = useState<
    Partial<Record<ServiceType, string>>
  >({});

  const [apiKeys, setApiKeys] = useState<PlateApiKey[]>([]);
  const [isApiKeysLoading, setIsApiKeysLoading] = useState(false);
  const [apiKeysError, setApiKeysError] = useState<string | null>(null);
  const [isCreatingApiKey, setIsCreatingApiKey] = useState(false);
  const [deletingApiKeyId, setDeletingApiKeyId] = useState<number | null>(null);
  const [createdApiKey, setCreatedApiKey] = useState<string | null>(null);
  const [copiedCreatedKey, setCopiedCreatedKey] = useState(false);

  const [settingsError, setSettingsError] = useState<string | null>(null);
  const [isDeletingPlate, setIsDeletingPlate] = useState(false);

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

  const loadPlate = useCallback(async () => {
    setIsLoading(true);
    setError(null);

    try {
      const body = await requestJson<PlateResponse>(
        `/plates/get?plateId=${plateId}`,
        undefined,
        "Failed to load plate.",
      );

      setPlate(body.plate ?? null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load plate.");
    } finally {
      setIsLoading(false);
    }
  }, [plateId, requestJson]);

  const loadApiKeys = useCallback(async () => {
    setIsApiKeysLoading(true);
    setApiKeysError(null);

    try {
      const body = await requestJson<ApiKeysResponse>(
        `/api-keys/list?plateId=${plateId}`,
        undefined,
        "Failed to load API keys.",
      );

      setApiKeys(Array.isArray(body.api_keys) ? body.api_keys : []);
    } catch (err) {
      setApiKeysError(
        err instanceof Error ? err.message : "Failed to load API keys.",
      );
    } finally {
      setIsApiKeysLoading(false);
    }
  }, [plateId, requestJson]);

  useEffect(() => {
    void loadPlate();
  }, [loadPlate]);

  useEffect(() => {
    if (activeSection === "api-keys") {
      void loadApiKeys();
    }
  }, [activeSection, loadApiKeys]);

  const handleEnableService = async (service: ServiceType) => {
    setServicePending(service);
    setServiceErrors((current) => {
      const next = { ...current };
      delete next[service];
      return next;
    });

    try {
      await requestJson<ServiceMutationResponse>(
        "/services/enable",
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({ plateId, service }),
        },
        "Failed to enable service.",
      );

      await loadPlate();
    } catch (err) {
      setServiceErrors((current) => ({
        ...current,
        [service]:
          err instanceof Error ? err.message : "Failed to enable service.",
      }));
    } finally {
      setServicePending(null);
    }
  };

  const handleDisableService = async (service: ServiceType) => {
    setServicePending(service);
    setServiceErrors((current) => {
      const next = { ...current };
      delete next[service];
      return next;
    });

    try {
      await requestJson<ServiceMutationResponse>(
        "/services/disable",
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({ plateId, service }),
        },
        "Failed to disable service.",
      );

      await loadPlate();
    } catch (err) {
      setServiceErrors((current) => ({
        ...current,
        [service]:
          err instanceof Error ? err.message : "Failed to disable service.",
      }));
    } finally {
      setServicePending(null);
    }
  };

  const handleCreateApiKey = async () => {
    setIsCreatingApiKey(true);
    setApiKeysError(null);
    setCopiedCreatedKey(false);

    try {
      const body = await requestJson<CreateApiKeyResponse>(
        "/api-keys/create",
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({ plateId }),
        },
        "Failed to create API key.",
      );

      setCreatedApiKey(body.api_key?.api_key ?? null);
      await loadApiKeys();
    } catch (err) {
      setApiKeysError(
        err instanceof Error ? err.message : "Failed to create API key.",
      );
    } finally {
      setIsCreatingApiKey(false);
    }
  };

  const handleDeleteApiKey = async (apiKeyId: number) => {
    setDeletingApiKeyId(apiKeyId);
    setApiKeysError(null);

    try {
      await requestJson<DeleteResponse>(
        "/api-keys/delete",
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({ apiKeyId }),
        },
        "Failed to delete API key.",
      );

      await loadApiKeys();
    } catch (err) {
      setApiKeysError(
        err instanceof Error ? err.message : "Failed to delete API key.",
      );
    } finally {
      setDeletingApiKeyId(null);
    }
  };

  const handleCopyCreatedKey = async () => {
    if (!createdApiKey) {
      return;
    }

    try {
      await navigator.clipboard.writeText(createdApiKey);
      setCopiedCreatedKey(true);
    } catch {
      setCopiedCreatedKey(false);
    }
  };

  const handleDeletePlate = async () => {
    setIsDeletingPlate(true);
    setSettingsError(null);

    try {
      await requestJson<DeleteResponse>(
        "/plates/delete",
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({ plateId }),
        },
        "Failed to delete plate.",
      );

      window.location.href = "/";
    } catch (err) {
      setSettingsError(
        err instanceof Error ? err.message : "Failed to delete plate.",
      );
    } finally {
      setIsDeletingPlate(false);
    }
  };

  const activeService =
    activeSection === "db" || activeSection === "kv" ? activeSection : null;

  const serviceContent =
    plate && activeService
      ? (serviceDefinitions.find((service) => service.type === activeService) ??
        null)
      : null;

  return (
    <SidebarProvider>
      <Sidebar variant="inset">
        <SidebarHeader className="gap-3 p-4">
          <Link
            href="/"
            className="inline-flex items-center gap-2 text-sm text-sidebar-foreground/80 hover:text-sidebar-foreground"
          >
            <ArrowLeftIcon className="size-4" />
            Back to plates
          </Link>
          <div className="rounded-xl border border-sidebar-border bg-sidebar-accent/40 px-3 py-3">
            <p className="font-medium text-sidebar-foreground">
              {plate ? plate.name : `Plate #${plateId}`}
            </p>
            <p className="mt-1 text-xs text-sidebar-foreground/70">
              Plate #{plateId}
            </p>
          </div>
        </SidebarHeader>

        <SidebarContent>
          <SidebarGroup>
            <SidebarGroupLabel>Services</SidebarGroupLabel>
            <SidebarGroupContent>
              <SidebarMenu>
                {serviceDefinitions.map((service) => {
                  const Icon = service.icon;
                  const enabled = plate
                    ? isServiceEnabled(plate, service.type)
                    : false;

                  return (
                    <SidebarMenuItem key={service.type}>
                      <SidebarMenuButton
                        isActive={activeSection === service.type}
                        onClick={() => setActiveSection(service.type)}
                        tooltip={service.label}
                      >
                        <Icon />
                        <span>{service.label}</span>
                      </SidebarMenuButton>
                      <SidebarMenuBadge>
                        {enabled ? "on" : "off"}
                      </SidebarMenuBadge>
                    </SidebarMenuItem>
                  );
                })}
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>

          <SidebarGroup>
            <SidebarGroupLabel>Manage</SidebarGroupLabel>
            <SidebarGroupContent>
              <SidebarMenu>
                <SidebarMenuItem>
                  <SidebarMenuButton
                    isActive={activeSection === "api-keys"}
                    onClick={() => setActiveSection("api-keys")}
                    tooltip="API keys"
                  >
                    <KeyRoundIcon />
                    <span>API keys</span>
                  </SidebarMenuButton>
                </SidebarMenuItem>
                <SidebarMenuItem>
                  <SidebarMenuButton
                    isActive={activeSection === "settings"}
                    onClick={() => setActiveSection("settings")}
                    tooltip="Settings"
                  >
                    <SettingsIcon />
                    <span>Settings</span>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>
        </SidebarContent>
      </Sidebar>

      <SidebarInset>
        <div className="flex min-h-screen flex-col">
          <div className="flex-1 px-4 py-6 sm:px-6">
            {isLoading ? (
              <div className="flex min-h-60 items-center justify-center rounded-2xl border">
                <Spinner />
              </div>
            ) : error ? (
              <Card>
                <CardHeader>
                  <CardTitle>Could not load plate</CardTitle>
                </CardHeader>
                <CardContent className="flex items-center justify-between gap-4">
                  <p className="text-sm text-muted-foreground">{error}</p>
                  <Button variant="outline" onClick={() => void loadPlate()}>
                    Retry
                  </Button>
                </CardContent>
              </Card>
            ) : plate === null ? (
              <Card>
                <CardHeader>
                  <CardTitle>Plate not found</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-sm text-muted-foreground">
                    This plate could not be found or you no longer have access
                    to it.
                  </p>
                </CardContent>
              </Card>
            ) : activeSection === "api-keys" ? (
              <PlateApiKeys
                apiKeys={apiKeys}
                isLoading={isApiKeysLoading}
                error={apiKeysError}
                isCreating={isCreatingApiKey}
                deletingKeyId={deletingApiKeyId}
                createdKey={createdApiKey}
                copiedKey={copiedCreatedKey}
                onCreate={() => void handleCreateApiKey()}
                onDelete={(id) => void handleDeleteApiKey(id)}
                onCopy={() => void handleCopyCreatedKey()}
              />
            ) : activeSection === "settings" ? (
              <PlateSettings
                plate={plate}
                serviceDefinitions={serviceDefinitions}
                error={settingsError}
                isDeleting={isDeletingPlate}
                onDeletePlate={() => void handleDeletePlate()}
              />
            ) : serviceContent ? (
              <PlateServiceContent
                plate={plate}
                serviceType={serviceContent.type}
                serviceDefinition={serviceContent}
                isPending={servicePending === serviceContent.type}
                error={serviceErrors[serviceContent.type] ?? null}
                onEnable={() => void handleEnableService(serviceContent.type)}
                onDisable={() => void handleDisableService(serviceContent.type)}
              />
            ) : null}
          </div>
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}
