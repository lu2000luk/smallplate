import { connectedServers, type ServerTypes } from ".";
import {
  getApiKeyRecord,
  getPlateById,
  getPlateDataObject,
  getPlateServers,
  setPlateDataObject,
  setPlateServers,
} from "./db";
import { db } from ".";
import { log, warn } from "./log";

type PlateServerMap = Partial<Record<ServerTypes, string>>;

type PlateDataShape = {
  enabled_services?: unknown;
} & Record<string, unknown>;

type CreateServiceResult =
  | {
      success: true;
      plateId: number;
      service: ServerTypes;
      serverId: string;
    }
  | {
      success: false;
      message: string;
    };

type DisableServiceResult =
  | {
      success: true;
      plateId: number;
      service: ServerTypes;
      serverId: string | null;
    }
  | {
      success: false;
      message: string;
    };

type PendingCreateRequest = {
  serverId: string;
  plateId: number;
  service: ServerTypes;
  resolve: (plateId: number) => void;
  reject: (error: Error) => void;
  timeout: ReturnType<typeof setTimeout>;
};

type ServerCreatedMessage = {
  type: "created";
  id: number;
};

type ServerErrorMessage = {
  type: "error";
  id?: number;
  message?: string;
  short?: string;
};

type PlateLifecycleEvent = {
  type: "delete";
  id: number;
};

type ApiKeyInvalidateEvent = {
  type: "invalidate";
  key: string;
};

type ServerCheckMessage = {
  type: "check";
  request_id: string;
  plate_id: string | number;
  key: string;
  service: string;
};

type ServerCheckResponseMessage = {
  type: "check_response";
  request_id: string;
  plate_id: string;
  key: string;
  valid: boolean;
};

const CREATE_SERVICE_TIMEOUT_MS = 30_000;
const pendingCreateRequests = new Map<string, PendingCreateRequest>();

function getPendingKey(
  serverId: string,
  plateId: number,
  service: ServerTypes,
) {
  return `${serverId}:${plateId}:${service}`;
}

function isServiceType(value: unknown): value is ServerTypes {
  return value === "db" || value === "kv";
}

function parseEnabledServices(data: unknown): ServerTypes[] {
  if (!data || typeof data !== "object" || Array.isArray(data)) {
    return [];
  }

  const rawEnabledServices = (data as PlateDataShape).enabled_services;
  if (!Array.isArray(rawEnabledServices)) {
    return [];
  }

  return rawEnabledServices.filter(isServiceType);
}

function parsePlateServerMap(value: unknown): PlateServerMap {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    return {};
  }

  const result: PlateServerMap = {};
  for (const [key, serverId] of Object.entries(value)) {
    if (
      isServiceType(key) &&
      typeof serverId === "string" &&
      serverId.length > 0
    ) {
      result[key] = serverId;
    }
  }

  return result;
}

function getStoredPlateServerMap(plateId: number): PlateServerMap | null {
  const servers = getPlateServers(plateId);
  if (servers === null) return null;
  return parsePlateServerMap(servers);
}

function listAllPlateServerMaps(): Array<{
  plateId: number;
  servers: PlateServerMap;
}> {
  const statement = db.query("SELECT id, servers FROM plates") as {
    all(): Array<{ id: number; servers: string | null }>;
  };

  return statement.all().map((row) => {
    let parsed: unknown = {};
    if (row.servers) {
      try {
        parsed = JSON.parse(row.servers);
      } catch {
        parsed = {};
      }
    }

    return {
      plateId: row.id,
      servers: parsePlateServerMap(parsed),
    };
  });
}

function countServicesOnServer(serverId: string, service: ServerTypes): number {
  let count = 0;

  for (const plate of listAllPlateServerMaps()) {
    if (plate.servers[service] === serverId) {
      count += 1;
    }
  }

  return count;
}

function chooseLeastLoadedServer(service: ServerTypes): string | null {
  const compatibleServerIds = Object.entries(connectedServers)
    .filter(([, server]) => server.type === service)
    .map(([serverId]) => serverId);

  if (compatibleServerIds.length === 0) {
    return null;
  }

  let lowestCount = Number.POSITIVE_INFINITY;
  let candidates: string[] = [];

  for (const serverId of compatibleServerIds) {
    const count = countServicesOnServer(serverId, service);

    if (count < lowestCount) {
      lowestCount = count;
      candidates = [serverId];
      continue;
    }

    if (count === lowestCount) {
      candidates.push(serverId);
    }
  }

  if (candidates.length === 0) {
    return null;
  }

  return candidates[Math.floor(Math.random() * candidates.length)] ?? null;
}

function requiresCreateConfirmation(service: ServerTypes): boolean {
  return service === "db";
}

function waitForCreated(
  serverId: string,
  plateId: number,
  service: ServerTypes,
): Promise<number> {
  return new Promise<number>((resolve, reject) => {
    const key = getPendingKey(serverId, plateId, service);

    const timeout = setTimeout(() => {
      pendingCreateRequests.delete(key);
      reject(
        new Error("Timed out waiting for server to confirm service creation."),
      );
    }, CREATE_SERVICE_TIMEOUT_MS);

    pendingCreateRequests.set(key, {
      serverId,
      plateId,
      service,
      resolve,
      reject,
      timeout,
    });
  });
}

function clearPendingRequest(key: string, pending: PendingCreateRequest) {
  clearTimeout(pending.timeout);
  pendingCreateRequests.delete(key);
}

function rollbackPlateState(
  plateId: number,
  previousData: Record<string, unknown>,
  previousServers: PlateServerMap,
) {
  setPlateDataObject(plateId, previousData);
  setPlateServers(plateId, previousServers);
}

function parsePlateId(value: unknown): number | null {
  if (typeof value === "number" && Number.isInteger(value) && value > 0) {
    return value;
  }

  if (typeof value === "string" && /^\d+$/.test(value)) {
    const parsed = Number.parseInt(value, 10);
    return parsed > 0 ? parsed : null;
  }

  return null;
}

function sendCheckResponse(serverId: string, message: ServerCheckResponseMessage) {
  const server = connectedServers[serverId];
  if (!server || server.socket.readyState !== WebSocket.OPEN) {
    return;
  }

  server.socket.send(JSON.stringify(message));
}

function handleCheckMessage(serverId: string, message: Partial<ServerCheckMessage>) {
  if (
    typeof message.request_id !== "string" ||
    message.request_id.length === 0 ||
    typeof message.key !== "string" ||
    message.key.length === 0 ||
    typeof message.service !== "string"
  ) {
    return true;
  }

  const plateId = parsePlateId(message.plate_id);
  if (plateId === null) {
    sendCheckResponse(serverId, {
      type: "check_response",
      request_id: message.request_id,
      plate_id: String(message.plate_id ?? ""),
      key: message.key,
      valid: false,
    });
    return true;
  }

  const connectedServer = connectedServers[serverId];
  const assignedServers = getPlateServers(plateId);
  const apiKey = getApiKeyRecord(message.key);
  const valid =
    !!connectedServer &&
    isServiceType(message.service) &&
    connectedServer.type === message.service &&
    assignedServers !== null &&
    assignedServers[message.service] === serverId &&
    apiKey?.plate_id === plateId;

  sendCheckResponse(serverId, {
    type: "check_response",
    request_id: message.request_id,
    plate_id: String(plateId),
    key: message.key,
    valid,
  });

  return true;
}

export function handleServerMessage(
  serverId: string,
  message: unknown,
): boolean {
  if (!message || typeof message !== "object" || Array.isArray(message)) {
    return false;
  }

  const type = (message as { type?: unknown }).type;

  if (type === "created") {
    const createdMessage = message as Partial<ServerCreatedMessage>;

    if (typeof createdMessage.id !== "number") {
      return true;
    }

    for (const [key, pending] of pendingCreateRequests.entries()) {
      if (pending.serverId !== serverId) continue;
      if (pending.plateId !== createdMessage.id) continue;

      clearPendingRequest(key, pending);
      pending.resolve(createdMessage.id);
      return true;
    }

    return true;
  }

  if (type === "create_error") {
    const errorMessage = message as ServerErrorMessage;

    for (const [key, pending] of pendingCreateRequests.entries()) {
      if (pending.serverId !== serverId) continue;
      if (
        typeof errorMessage.id === "number" &&
        pending.plateId !== errorMessage.id
      ) {
        continue;
      }

      clearPendingRequest(key, pending);
      pending.reject(
        new Error(
          errorMessage.message ||
            errorMessage.short ||
            "Server rejected service creation.",
        ),
      );
      return true;
    }

    return true;
  }

  if (type === "check") {
    return handleCheckMessage(serverId, message as Partial<ServerCheckMessage>);
  }

  return false;
}

export function handleServerDisconnect(serverId: string) {
  for (const [key, pending] of pendingCreateRequests.entries()) {
    if (pending.serverId !== serverId) continue;

    clearPendingRequest(key, pending);
    pending.reject(
      new Error(`Server ${serverId} disconnected before confirming creation.`),
    );
  }
}

export async function createService(
  plateId: number,
  service: ServerTypes,
): Promise<CreateServiceResult> {
  const plate = getPlateById(plateId);
  if (!plate) {
    return {
      success: false,
      message: "Plate does not exist.",
    };
  }

  const currentDataRaw = getPlateDataObject(plateId);
  const currentServersRaw = getStoredPlateServerMap(plateId);

  if (currentDataRaw === null || currentServersRaw === null) {
    return {
      success: false,
      message: "Failed to load plate state.",
    };
  }

  const currentData: Record<string, unknown> = {
    ...currentDataRaw,
  };
  const currentServers: PlateServerMap = {
    ...currentServersRaw,
  };

  const enabledServices = new Set(parseEnabledServices(currentData));
  if (
    enabledServices.has(service) ||
    typeof currentServers[service] === "string"
  ) {
    return {
      success: false,
      message: "Service is already enabled for this plate.",
    };
  }

  const serverId = chooseLeastLoadedServer(service);
  if (!serverId) {
    return {
      success: false,
      message: `No connected ${service} server is available.`,
    };
  }

  const selectedServer = connectedServers[serverId];
  if (!selectedServer || selectedServer.type !== service) {
    return {
      success: false,
      message: "Selected server is no longer available.",
    };
  }

  const nextData: Record<string, unknown> = {
    ...currentData,
    enabled_services: [...enabledServices, service],
  };

  const nextServers: PlateServerMap = {
    ...currentServers,
    [service]: serverId,
  };

  const dataUpdated = setPlateDataObject(plateId, nextData);
  const serversUpdated = setPlateServers(plateId, nextServers);

  if (!dataUpdated || !serversUpdated) {
    rollbackPlateState(plateId, currentData, currentServers);
    return {
      success: false,
      message: "Failed to update plate state before provisioning service.",
    };
  }

  const updatedPlate = getPlateById(plateId);
  if (!updatedPlate) {
    rollbackPlateState(plateId, currentData, currentServers);
    return {
      success: false,
      message: "Plate disappeared after update.",
    };
  }

  const payload = {
    ...updatedPlate,
    servers: nextServers,
    data: nextData,
  };

  try {
    let createdId = plateId;

    if (requiresCreateConfirmation(service)) {
      const waitPromise = waitForCreated(serverId, plateId, service);

      selectedServer.socket.send(
        JSON.stringify({
          type: "create",
          data: payload,
        }),
      );

      createdId = await waitPromise;
    }

    log(
      "Created service for plate",
      plateId,
      "service",
      service,
      "on server",
      serverId,
    );

    return {
      success: true,
      plateId: createdId,
      service,
      serverId,
    };
  } catch (error) {
    rollbackPlateState(plateId, currentData, currentServers);
    warn("Failed to create service on server.", error);

    return {
      success: false,
      message:
        error instanceof Error
          ? error.message
          : "Failed while waiting for server confirmation.",
    };
  }
}

export function disableService(
  plateId: number,
  service: ServerTypes,
): DisableServiceResult {
  const plate = getPlateById(plateId);
  if (!plate) {
    return {
      success: false,
      message: "Plate does not exist.",
    };
  }

  const currentDataRaw = getPlateDataObject(plateId);
  const currentServersRaw = getStoredPlateServerMap(plateId);

  if (currentDataRaw === null || currentServersRaw === null) {
    return {
      success: false,
      message: "Failed to load plate state.",
    };
  }

  const currentData: Record<string, unknown> = {
    ...currentDataRaw,
  };
  const currentServers: PlateServerMap = {
    ...currentServersRaw,
  };

  const enabledServices = new Set(parseEnabledServices(currentData));
  const assignedServerId = currentServers[service] ?? null;

  if (!enabledServices.has(service) && assignedServerId === null) {
    return {
      success: false,
      message: "Service is not enabled for this plate.",
    };
  }

  enabledServices.delete(service);

  const nextData: Record<string, unknown> = {
    ...currentData,
    enabled_services: [...enabledServices],
  };

  const nextServers: PlateServerMap = {
    ...currentServers,
  };
  delete nextServers[service];

  const dataUpdated = setPlateDataObject(plateId, nextData);
  const serversUpdated = setPlateServers(plateId, nextServers);

  if (!dataUpdated || !serversUpdated) {
    rollbackPlateState(plateId, currentData, currentServers);
    return {
      success: false,
      message: "Failed to update plate state while disabling service.",
    };
  }

  return {
    success: true,
    plateId,
    service,
    serverId: assignedServerId,
  };
}

export function invalidateApiKeyEverywhere(apiKey: string) {
  const payload: ApiKeyInvalidateEvent = {
    type: "invalidate",
    key: apiKey,
  };
  const encodedPayload = JSON.stringify(payload);

  for (const server of Object.values(connectedServers)) {
    if (server.socket.readyState !== WebSocket.OPEN) {
      continue;
    }

    server.socket.send(encodedPayload);
  }
}

export function deletePlateOnServer(serverId: string, plateId: number) {
  const server = connectedServers[serverId];
  if (!server || server.socket.readyState !== WebSocket.OPEN) {
    return false;
  }

  server.socket.send(
    JSON.stringify({
      type: "delete",
      id: plateId,
    } satisfies PlateLifecycleEvent),
  );

  return true;
}

export function deletePlateEverywhere(plateId: number) {
  for (const serverId of Object.keys(connectedServers)) {
    deletePlateOnServer(serverId, plateId);
  }
}
