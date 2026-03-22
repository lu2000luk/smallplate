#!/usr/bin/env bun
import chalk from "chalk";
import inquirer from "inquirer";
import ora from "ora";
import { writeFileSync, existsSync, mkdirSync } from "node:fs";
import { spawn } from "node:child_process";

type ServiceType = "db" | "kv" | "vec" | "link";

interface Endpoint {
  method: string;
  path: string;
  body?: unknown;
  rawBody?: string;
  contentType?: string;
  requiresPlateId?: boolean;
  customId?: string;
  queryParams?: Record<string, string>;
  extraHeaders?: Record<string, string>;
}

interface ServiceEndpoints {
  service: ServiceType;
  endpoints: Endpoint[];
}

interface TestCategory {
  key: string;
  label: string;
  service: ServiceType;
  endpoints: Endpoint[];
}

interface TestResult {
  name: string;
  service: ServiceType;
  category: string;
  method: string;
  path: string;
  fullUrl: string;
  requestHeaders?: Record<string, string>;
  requestBody?: string;
  responseBody?: string;
  responseHeaders?: Record<string, string>;
  success: boolean;
  status: number;
  statusText?: string;
  time: number;
  error?: string;
  errorDetails?: string;
  curl?: string;
}

interface CliOptions {
  managerUrl: string;
  plateId: string;
  apiKey: string;
  serverDbUrl: string;
  serverKvUrl: string;
  serverVecUrl: string;
  serverLinkUrl: string;
  debug: boolean;
  noOpen: boolean;
  openReport: boolean;
  timeoutMs: number;
  reportDir: string;
  nonInteractive: boolean;
  selectedServices: ServiceType[];
  openrouterKey: string;
  openrouterModel: string;
}

const SERVICES: ServiceEndpoints[] = [
  {
    service: "db",
    endpoints: [
      { method: "POST", path: "/{id}" },
      { method: "GET", path: "/{id}/info" },
      { method: "POST", path: "/{id}/execute", body: { sql: "CREATE TABLE IF NOT EXISTS test (id INTEGER PRIMARY KEY, name TEXT)", params: [] } },
      { method: "POST", path: "/{id}/query", body: { sql: "SELECT name FROM sqlite_master WHERE type=?", params: ["table"] } },
      {
        method: "POST",
        path: "/{id}/batch",
        body: {
          transaction: true,
          statements: [
            { sql: "INSERT INTO test (name) VALUES (?)", params: ["tester"] },
            { sql: "SELECT COUNT(*) FROM test", params: [] },
          ],
        },
      },
      { method: "POST", path: "/{id}/transactions" },
      { method: "GET", path: "/{id}/tables" },
      { method: "GET", path: "/{id}/tables/test" },
      { method: "GET", path: "/{id}/tables/test/indexes" },
      { method: "GET", path: "/{id}/schema" },
      { method: "GET", path: "/{id}/export", queryParams: { format: "sql" } },
      {
        method: "POST",
        path: "/{id}/import/sql",
        rawBody: "CREATE TABLE IF NOT EXISTS import_test (id INTEGER);",
        contentType: "application/sql",
      },
      { method: "GET", path: "/healthz" },
    ],
  },
  {
    service: "kv",
    endpoints: [
      { method: "POST", path: "/{id}/strings/set", body: { key: "testkey", value: "testvalue" } },
      { method: "GET", path: "/{id}/strings/get/testkey" },
      { method: "POST", path: "/{id}/keys/ttl", body: { key: "testkey", ttl_ms: 60000 } },
      { method: "DELETE", path: "/{id}/keys/ttl", body: { key: "testkey" } },
      { method: "POST", path: "/{id}/hashes/set", body: { key: "testhash", value: { name: "tester", age: 1 } } },
      { method: "GET", path: "/{id}/hashes/get/testhash" },
      { method: "POST", path: "/{id}/lists/right/push", body: { key: "testlist", values: ["item1", "item2"] } },
      { method: "GET", path: "/{id}/lists/range/testlist", queryParams: { start: "0", stop: "10" } },
      { method: "POST", path: "/{id}/sets/add", body: { key: "testset", members: ["member1", "member2"] } },
      { method: "GET", path: "/{id}/sets/members/testset" },
      { method: "POST", path: "/{id}/zsets/add", body: { key: "testzset", members: [{ member: "member1", score: 1 }] } },
      { method: "GET", path: "/{id}/zsets/range/testzset", queryParams: { start: "0", stop: "-1" } },
      { method: "POST", path: "/{id}/streams/add", body: { key: "teststream", values: { event: "test" } } },
      { method: "GET", path: "/{id}/streams/length/teststream" },
      { method: "POST", path: "/{id}/bitmaps/set", body: { key: "testbit", bit: 1, value: 1 } },
      { method: "GET", path: "/{id}/bitmaps/get/testbit/1" },
      { method: "POST", path: "/{id}/geo/add", body: { key: "testgeo", locations: [{ member: "sf", longitude: -122.4194, latitude: 37.7749 }] } },
      { method: "POST", path: "/{id}/geo/positions", body: { key: "testgeo", members: ["sf"] } },
      { method: "POST", path: "/{id}/json/testjson", body: { value: { hello: "world" } } },
      { method: "GET", path: "/{id}/json/testjson" },
      { method: "POST", path: "/{id}/publish/testchannel", body: { message: "hello" } },
      { method: "POST", path: "/{id}/pipeline", body: [["SET", "pipekey", "pipevalue"], ["GET", "pipekey"]] },
      { method: "POST", path: "/{id}/transaction", body: [["SET", "txnkey", "txnvalue"], ["GET", "txnkey"]] },
      { method: "GET", path: "/{id}/info" },
      { method: "GET", path: "/healthz" },
    ],
  },
  {
    service: "vec",
    endpoints: [
      { method: "POST", path: "/{plateID}" },
      { method: "GET", path: "/{plateID}/info" },
      { method: "GET", path: "/{plateID}/limits" },
      { method: "GET", path: "/{plateID}/embedding-providers" },
      { method: "POST", path: "/{plateID}/databases", body: { name: "testdb" } },
      { method: "GET", path: "/{plateID}/databases/testdb" },
      { method: "POST", path: "/{plateID}/databases/testdb/collections", body: { name: "testcol", get_or_create: true } },
      {
        method: "POST",
        path: "/{plateID}/databases/testdb/collections/testcol/records/upsert",
        body: { records: [{ id: "rec1", embedding: [0.11, 0.22, 0.33, 0.44], metadata: { source: "tester" } }] },
      },
      { method: "POST", path: "/{plateID}/databases/testdb/collections/testcol/records/get", body: { ids: ["rec1"] } },
      {
        method: "POST",
        path: "/{plateID}/databases/testdb/collections/testcol/records/query",
        body: { query_embeddings: [[0.11, 0.22, 0.33, 0.44]], n_results: 1 },
      },
      { method: "GET", path: "/{plateID}/databases/testdb/collections/testcol/records/count" },
      { method: "GET", path: "/healthz" },
    ],
  },
  {
    service: "link",
    endpoints: [
      { method: "POST", path: "/{plateID}/create", body: { destination: "https://example.com", id_prefix: "test_" } },
      { method: "POST", path: "/{plateID}/create/dynamic", body: { template: "https://example.com/{}", id_prefix: "dyn_" } },
      { method: "GET", path: "/{plateID}/links" },
      { method: "GET", path: "/healthz" },
    ],
  },
];

const SERVICE_DOCS: Record<ServiceType, string> = {
  db: "/docs/db/",
  kv: "/docs/kv/",
  vec: "/docs/vec/",
  link: "/docs/link/",
};

function parseServiceList(value: string): ServiceType[] {
  const requested = value
    .split(",")
    .map((s) => s.trim().toLowerCase())
    .filter(Boolean);

  const services = requested.filter((s): s is ServiceType => ["db", "kv", "vec", "link"].includes(s));
  return [...new Set(services)];
}

function parseCliOptions(args: string[]): CliOptions {
  const options: CliOptions = {
    managerUrl: "",
    plateId: "",
    apiKey: "",
    serverDbUrl: "",
    serverKvUrl: "",
    serverVecUrl: "",
    serverLinkUrl: "",
    debug: false,
    noOpen: false,
    openReport: false,
    timeoutMs: 15000,
    reportDir: ".",
    nonInteractive: false,
    selectedServices: [],
    openrouterKey: "",
    openrouterModel: "",
  };

  for (let i = 0; i < args.length; i++) {
    const arg = args[i];
    const next = args[i + 1];

    if (arg === "--url" || arg === "-u") {
      options.managerUrl = next || "";
      i++;
    } else if (arg === "--plate" || arg === "-p") {
      options.plateId = next || "";
      i++;
    } else if (arg === "--key" || arg === "-k") {
      options.apiKey = next || "";
      i++;
    } else if (arg === "--db") {
      options.serverDbUrl = next || "";
      i++;
    } else if (arg === "--kv") {
      options.serverKvUrl = next || "";
      i++;
    } else if (arg === "--vec") {
      options.serverVecUrl = next || "";
      i++;
    } else if (arg === "--link") {
      options.serverLinkUrl = next || "";
      i++;
    } else if (arg === "--services" || arg === "-s") {
      options.selectedServices = parseServiceList(next || "");
      i++;
    } else if (arg === "--timeout-ms") {
      const parsed = Number.parseInt(next || "", 10);
      if (Number.isFinite(parsed) && parsed > 0) {
        options.timeoutMs = parsed;
      }
      i++;
    } else if (arg === "--report-dir") {
      options.reportDir = next || ".";
      i++;
    } else if (arg === "--openrouter-key") {
      options.openrouterKey = next || "";
      i++;
    } else if (arg === "--openrouter-model") {
      options.openrouterModel = next || "";
      i++;
    } else if (arg === "--debug" || arg === "-d") {
      options.debug = true;
    } else if (arg === "--non-interactive" || arg === "-y") {
      options.nonInteractive = true;
    } else if (arg === "--no-open") {
      options.noOpen = true;
    } else if (arg === "--open") {
      options.openReport = true;
    }
  }

  return options;
}

function buildFullUrl(baseUrl: string, endpointPath: string, id: string): URL {
  const normalizedBase = baseUrl.endsWith("/") ? baseUrl.slice(0, -1) : baseUrl;
  const normalizedPath = endpointPath.startsWith("/") ? endpointPath : `/${endpointPath}`;
  const rawUrl = `${normalizedBase}${normalizedPath}`
    .replaceAll("{id}", id)
    .replaceAll("{plateId}", id)
    .replaceAll("{plateID}", id);

  return new URL(rawUrl);
}

function shellEscapeSingleQuotes(input: string): string {
  return input.replace(/'/g, `'"'"'`);
}

function toCurlCommand(method: string, fullUrl: string, headers: Record<string, string>, requestBody?: string): string {
  const parts = [`curl -X ${method}`, `"${fullUrl}"`];

  for (const [key, value] of Object.entries(headers)) {
    parts.push(`-H "${key}: ${value}"`);
  }

  if (requestBody) {
    parts.push(`--data '${shellEscapeSingleQuotes(requestBody)}'`);
  }

  return parts.join(" ");
}

async function readResponseBody(response: Response): Promise<string> {
  try {
    return await response.text();
  } catch {
    return "";
  }
}

function buildRuntimeCategories(options: CliOptions): TestCategory[] {
  const baseCategories: TestCategory[] = SERVICES.map((serviceConfig) => ({
    key: serviceConfig.service,
    label: serviceConfig.service.toUpperCase(),
    service: serviceConfig.service,
    endpoints: serviceConfig.endpoints,
  }));

  const extras: TestCategory[] = [];

  if (options.openrouterKey && options.openrouterModel) {
    extras.push({
      key: "vec-openrouter-embeddings",
      label: "VEC OpenRouter Embeddings",
      service: "vec",
      endpoints: [
        {
          method: "POST",
          path: "/{plateID}/embeddings/run",
          body: {
            inputs: ["smallplate embedding test"],
            provider: "openrouter",
            model: options.openrouterModel,
          },
          extraHeaders: {
            "X-Embedding-Api-Key": options.openrouterKey,
          },
        },
      ],
    });
  }

  return [...baseCategories, ...extras];
}

async function fetchWithTimeout(url: string, options: RequestInit, timeout = 15000): Promise<Response> {
  const controller = new AbortController();
  const id = setTimeout(() => controller.abort(), timeout);
  try {
    const response = await fetch(url, { ...options, signal: controller.signal });
    return response;
  } finally {
    clearTimeout(id);
  }
}

async function testEndpoint(
  service: ServiceType,
  category: string,
  baseUrl: string,
  apiKey: string,
  endpoint: Endpoint,
  id: string,
  timeoutMs: number
): Promise<TestResult> {
  const startTime = Date.now();
  const fullUrl = buildFullUrl(baseUrl, endpoint.path, id);

  if (endpoint.queryParams) {
    for (const [key, value] of Object.entries(endpoint.queryParams)) {
      fullUrl.searchParams.set(key, value);
    }
  }

  const headers: Record<string, string> = {
    "Content-Type": endpoint.contentType || "application/json",
    "Authorization": apiKey,
    ...(endpoint.extraHeaders || {}),
  };

  const requestBody = endpoint.rawBody ?? (endpoint.body ? JSON.stringify(endpoint.body, null, 2) : undefined);
  const curl = toCurlCommand(endpoint.method, fullUrl.toString(), headers, requestBody);

  try {
    const response = await fetchWithTimeout(fullUrl.toString(), {
      method: endpoint.method,
      headers,
      body: endpoint.rawBody ?? (endpoint.body ? JSON.stringify(endpoint.body) : undefined),
    }, timeoutMs);

    const time = Date.now() - startTime;
    const responseText = await readResponseBody(response);
    const responseHeaders: Record<string, string> = {};
    response.headers.forEach((value, key) => {
      responseHeaders[key] = value;
    });
    let responseBody = "";
    let parsedOk = true;
    let errorCode = "";
    let errorMessage = "";

    try {
      const json = JSON.parse(responseText);
      responseBody = JSON.stringify(json, null, 2);
      parsedOk = json.ok !== false;
      errorCode = json.error?.code || "";
      errorMessage = json.error?.message || "";
    } catch {
      responseBody = responseText;
    }

    const success = response.ok && parsedOk;

    return {
      name: `${endpoint.method} ${endpoint.path}`,
      service,
      category,
      method: endpoint.method,
      path: endpoint.path,
      fullUrl: fullUrl.toString(),
      requestHeaders: headers,
      requestBody,
      responseBody,
      responseHeaders,
      success,
      status: response.status,
      statusText: response.statusText,
      time,
      error: !success ? (errorMessage || errorCode || `HTTP ${response.status}`) : undefined,
      errorDetails: !success ? responseBody.substring(0, 1000) : undefined,
      curl,
    };
  } catch (err) {
    const time = Date.now() - startTime;
    const errorMsg = err instanceof Error ? err.message : "Unknown error";
    return {
      name: `${endpoint.method} ${endpoint.path}`,
      service,
      category,
      method: endpoint.method,
      path: endpoint.path,
      fullUrl: fullUrl.toString(),
      requestHeaders: headers,
      requestBody,
      responseBody: "",
      success: false,
      status: 0,
      statusText: "Network Error",
      time,
      error: errorMsg,
      errorDetails: err instanceof Error ? err.stack : undefined,
      curl,
    };
  }
}

async function getServiceUrl(managerUrl: string, serverId: string): Promise<string | null> {
  try {
    const response = await fetch(`${managerUrl}/services/url?id=${serverId}`);
    const data = await response.json() as { success?: boolean; url?: string };
    return data.success && data.url ? data.url : null;
  } catch {
    return null;
  }
}

async function getPlateData(managerUrl: string, plateId: string): Promise<{ id: number; servers: Record<string, string> } | null> {
  try {
    const response = await fetch(`${managerUrl}/plates/get?plateId=${plateId}`);
    const data = await response.json() as {
      success?: boolean;
      plate?: { id?: number; servers?: Record<string, string> };
    };
    if (data.success && data.plate) {
      return {
        id: data.plate.id || Number.parseInt(plateId, 10),
        servers: data.plate.servers || {},
      };
    }
    return null;
  } catch {
    return null;
  }
}

function classifyLikelyIssue(result: TestResult): string {
  if (result.status === 405) {
    return "Method mismatch: docs and route likely changed (check HTTP verb).";
  }

  if (result.status === 404) {
    return "Route not found: path may be outdated or wrong service base URL.";
  }

  if (result.status === 401 || result.status === 403) {
    return "Auth issue: verify Authorization key and service auth rules.";
  }

  if (result.status === 0) {
    return "Network/timeout issue: verify service availability and URL.";
  }

  if (result.status >= 500) {
    return "Server error: check service logs for internal exceptions.";
  }

  return "Inspect request/response payload for validation or contract mismatch.";
}

function printDebugBlock(result: TestResult): void {
  console.log(chalk.yellow("       Debug:"));
  console.log(chalk.gray(`         Hint: ${classifyLikelyIssue(result)}`));
  console.log(chalk.gray(`         cURL: ${result.curl}`));

  if (result.responseHeaders && Object.keys(result.responseHeaders).length > 0) {
    const contentType = result.responseHeaders["content-type"] || result.responseHeaders["Content-Type"] || "";
    if (contentType) {
      console.log(chalk.gray(`         Response Content-Type: ${contentType}`));
    }
  }

  console.log(chalk.gray(`         Docs: ${SERVICE_DOCS[result.service]}`));
}

function printResultDetails(result: TestResult, debug: boolean): void {
  const statusIcon = result.success ? chalk.green("✓") : chalk.red("✗");
  const statusText = result.success ? chalk.green(`[${result.status}]`) : chalk.red(`[${result.status}]`);

  console.log(`  ${statusIcon} ${statusText} ${result.name} (${result.time}ms)`);
  console.log(chalk.gray(`       URL: ${result.fullUrl}`));

  if (result.requestHeaders) {
    console.log(chalk.gray("       Request Headers:"));
    for (const [key, value] of Object.entries(result.requestHeaders)) {
      console.log(chalk.gray(`         ${key}: ${value}`));
    }
  }

  if (result.requestBody) {
    console.log(chalk.gray("       Request Body:"));
    for (const line of result.requestBody.split("\n")) {
      console.log(chalk.gray(`         ${line}`));
    }
  }

  if (result.responseHeaders && Object.keys(result.responseHeaders).length > 0) {
    console.log(chalk.gray("       Response Headers:"));
    for (const [key, value] of Object.entries(result.responseHeaders)) {
      console.log(chalk.gray(`         ${key}: ${value}`));
    }
  }

  const responseColor = result.success ? chalk.gray : chalk.red;
  if (result.responseBody) {
    console.log(responseColor("       Response Body:"));
    for (const line of result.responseBody.split("\n")) {
      console.log(responseColor(`         ${line}`));
    }
  }

  if (!result.success) {
    console.log(chalk.red(`       Error: ${result.error}`));
    if (debug) {
      printDebugBlock(result);
    }
  }
}

function generateReport(
  results: { service: string; results: TestResult[] }[],
  plateId: string,
  apiKey: string,
  debug: boolean
): string {
  const lines: string[] = [];

  lines.push("=".repeat(70));
  lines.push("PLATE SERVICE TEST REPORT");
  lines.push("=".repeat(70));
  lines.push(`Generated: ${new Date().toLocaleString()}`);
  lines.push(`Plate ID: ${plateId}`);
  lines.push(`API Key: ${apiKey.substring(0, 8)}...`);
  lines.push("");

  let totalPassed = 0;
  let totalFailed = 0;

  for (const serviceResult of results) {
    const passed = serviceResult.results.filter(r => r.success).length;
    const failed = serviceResult.results.filter(r => !r.success).length;
    totalPassed += passed;
    totalFailed += failed;

    lines.push("=".repeat(70));
    lines.push(`SERVICE: ${serviceResult.service.toUpperCase()}`);
    lines.push("=".repeat(70));
    lines.push(`Passed: ${passed} | Failed: ${failed}`);
    lines.push("");

    for (const result of serviceResult.results) {
      const status = result.success ? "PASS" : "FAIL";
      lines.push(`[${status}] ${result.method} ${result.path}`);
      lines.push(`       Category: ${result.category}`);
      lines.push(`       Status: ${result.status} (${result.time}ms)`);
      
      if (!result.success) {
        lines.push(`       Error: ${result.error}`);
        if (result.errorDetails) {
          lines.push(`       Details: ${result.errorDetails.substring(0, 200)}`);
        }
        if (debug) {
          lines.push(`       Hint: ${classifyLikelyIssue(result)}`);
        }
      }
      
      if (result.requestBody) {
        lines.push("       Request Body:");
        lines.push(result.requestBody);
      }
      if (result.requestHeaders && Object.keys(result.requestHeaders).length > 0) {
        lines.push("       Request Headers:");
        for (const [key, value] of Object.entries(result.requestHeaders)) {
          lines.push(`         ${key}: ${value}`);
        }
      }
      if (result.responseBody) {
        lines.push("       Response Body:");
        lines.push(result.responseBody);
      }
      if (result.responseHeaders && Object.keys(result.responseHeaders).length > 0) {
        lines.push("       Response Headers:");
        for (const [key, value] of Object.entries(result.responseHeaders)) {
          lines.push(`         ${key}: ${value}`);
        }
      }
      if (debug && !result.success && result.curl) {
        lines.push(`       cURL: ${result.curl}`);
        lines.push(`       Docs: ${SERVICE_DOCS[result.service]}`);
      }
      lines.push("");
    }
  }

  lines.push("=".repeat(70));
  lines.push("SUMMARY");
  lines.push("=".repeat(70));
  const total = totalPassed + totalFailed;
  lines.push(`Total: ${totalPassed}/${total} passed`);
  if (totalFailed > 0) {
    lines.push(`Failed: ${totalFailed}`);
  } else {
    lines.push("All tests passed!");
  }
  lines.push("");

  return lines.join("\n");
}

async function main() {
  console.log(chalk.cyan(`
  ╔═══════════════════════════════════════════════════╗
  ║           Plate Service Test Runner               ║
  ╚═══════════════════════════════════════════════════╝
  `));

  const options = parseCliOptions(process.argv.slice(2));

  if (!options.plateId || !options.apiKey) {
    console.log(chalk.yellow("Usage: bun run index.ts -p <plate-id> -k <api-key> [options]"));
    console.log(chalk.gray("Options:"));
    console.log(chalk.gray("  -u, --url <url>      Manager URL (for auto-discovery)"));
    console.log(chalk.gray("  -p, --plate <id>     Plate ID (required)"));
    console.log(chalk.gray("  -k, --key <key>      API Key (required)"));
    console.log(chalk.gray("  --db <url>           DB service URL"));
    console.log(chalk.gray("  --kv <url>           KV service URL"));
    console.log(chalk.gray("  --vec <url>          Vec service URL"));
    console.log(chalk.gray("  --link <url>         Link service URL"));
    console.log(chalk.gray("  -s, --services <csv> Services to test (db,kv,vec,link)"));
    console.log(chalk.gray("  --timeout-ms <ms>    Request timeout in milliseconds"));
    console.log(chalk.gray("  --report-dir <dir>   Directory to save reports"));
    console.log(chalk.gray("  --openrouter-key <key> OpenRouter API key for embedding tests"));
    console.log(chalk.gray("  --openrouter-model <id> OpenRouter embedding model id"));
    console.log(chalk.gray("  -d, --debug          Show failure debugging details"));
    console.log(chalk.gray("  -y, --non-interactive Skip prompts and run selected services"));
    console.log(chalk.gray("  --no-open            Do not open report after saving"));
    console.log(chalk.gray("  --open               Force opening report after saving"));
    console.log(chalk.gray("\nExample:"));
    console.log(chalk.gray("  bun run index.ts -p 1 -k key --db http://localhost:3001 --kv http://localhost:3002 --debug -y"));
    process.exit(1);
  }

  const serviceUrls: Record<ServiceType, string> = {} as Record<ServiceType, string>;
  let availableServices: ServiceType[] = [];

  if (options.managerUrl) {
    const spinner = ora("Fetching plate data...").start();
    const plateData = await getPlateData(options.managerUrl, options.plateId);
    
    if (!plateData) {
      spinner.fail("Failed to fetch plate data (requires login auth). Use --db, --kv, etc. to specify URLs directly.");
      console.log(chalk.yellow("\nTip: You can specify service URLs directly:"));
      console.log(chalk.gray("  bun run index.ts -p 1 -k key --db http://... --kv http://..."));
      process.exit(1);
    }

    spinner.succeed(`Plate #${plateData.id} fetched successfully`);

    for (const [service, serverId] of Object.entries(plateData.servers)) {
      if (serverId && ["db", "kv", "vec", "link"].includes(service)) {
        const url = await getServiceUrl(options.managerUrl, serverId);
        if (url) {
          availableServices.push(service as ServiceType);
          serviceUrls[service as ServiceType] = url;
        }
      }
    }
  } else if (options.serverDbUrl || options.serverKvUrl || options.serverVecUrl || options.serverLinkUrl) {
    if (options.serverDbUrl) { availableServices.push("db"); serviceUrls.db = options.serverDbUrl; }
    if (options.serverKvUrl) { availableServices.push("kv"); serviceUrls.kv = options.serverKvUrl; }
    if (options.serverVecUrl) { availableServices.push("vec"); serviceUrls.vec = options.serverVecUrl; }
    if (options.serverLinkUrl) { availableServices.push("link"); serviceUrls.link = options.serverLinkUrl; }
    console.log(chalk.green("Using manually specified service URLs"));
  } else {
    console.log(chalk.red("No service URLs provided. Use --url for auto-discovery or specify URLs with --db, --kv, --vec, --link"));
    process.exit(1);
  }

  if (availableServices.length === 0) {
    console.log(chalk.red("No services available"));
    process.exit(1);
  }

  console.log(chalk.bold("\nAvailable services:"));
  for (const s of availableServices) {
    console.log(chalk.cyan(`  - ${s}`) + chalk.gray(` (${serviceUrls[s]})`));
  }

  console.log(chalk.bold("\n" + "=".repeat(60)));
  console.log(chalk.bold("AVAILABLE ENDPOINTS BY SERVICE"));
  console.log(chalk.bold("=".repeat(60)));

  const categories = buildRuntimeCategories(options).filter((category) => availableServices.includes(category.service));

  for (const category of categories) {
    console.log(chalk.cyan(`\n${category.label} (${category.endpoints.length} endpoints):`));
    for (const ep of category.endpoints) {
      console.log(`  ${ep.method.padEnd(7)} ${ep.path}`);
    }
  }
  console.log(chalk.bold("=".repeat(60)));

  let selectedServices: ServiceType[] = [];

  if (options.selectedServices.length > 0) {
    selectedServices = options.selectedServices.filter((s) => availableServices.includes(s));
  }

  if (selectedServices.length === 0) {
    if (options.nonInteractive) {
      selectedServices = [...availableServices];
    } else {
      const prompt = await inquirer.prompt([
        {
          type: "checkbox",
          name: "selectedServices",
          message: "Select services to test:",
          choices: availableServices.map(s => ({ name: s, value: s, checked: true })),
        },
      ]);
      selectedServices = prompt.selectedServices;
    }
  }

  if (selectedServices.length === 0) {
    console.log(chalk.yellow("No services selected. Exiting."));
    process.exit(0);
  }

  console.log(chalk.bold("\nSelected services: ") + chalk.cyan(selectedServices.join(", ")));

  const allResults: { service: string; results: TestResult[] }[] = [];

  const categoriesToRun = categories.filter((category) => selectedServices.includes(category.service));

  for (const serviceType of selectedServices) {
    const serviceCategories = categoriesToRun.filter((category) => category.service === serviceType);
    if (serviceCategories.length === 0) continue;

    console.log(chalk.bold(`\n${"=".repeat(50)}`));
    console.log(chalk.bold(`Testing ${serviceType.toUpperCase()} endpoints`));
    console.log(chalk.bold(`${"=".repeat(50)}`));

    const serviceSpinner = ora(`Testing ${serviceType}...`).start();
    const serviceResults: TestResult[] = [];

    const baseUrl = serviceUrls[serviceType];
    let completed = 0;
    const total = serviceCategories.reduce((sum, category) => sum + category.endpoints.length, 0);

    for (const category of serviceCategories) {
      for (const endpoint of category.endpoints) {
        const result = await testEndpoint(serviceType, category.label, baseUrl, options.apiKey, endpoint, options.plateId, options.timeoutMs);
        serviceResults.push(result);
        completed++;
        serviceSpinner.text = `Testing ${serviceType}... ${completed}/${total}`;
      }
    }

    const passed = serviceResults.filter(r => r.success).length;
    serviceSpinner.succeed(`Completed ${serviceType} tests (${passed}/${total} passed)`);

    allResults.push({
      service: serviceType,
      results: serviceResults,
    });
  }

  console.log("\n" + chalk.bold("=".repeat(60)));
  console.log(chalk.bold("TEST RESULTS SUMMARY"));
  console.log(chalk.bold("=".repeat(60)) + "\n");

  let totalPassed = 0;
  let totalFailed = 0;

  for (const serviceResult of allResults) {
    const passed = serviceResult.results.filter(r => r.success).length;
    const failed = serviceResult.results.filter(r => !r.success).length;
    totalPassed += passed;
    totalFailed += failed;

    const color = failed === 0 ? chalk.green : chalk.red;
    console.log(chalk.bold(`${serviceResult.service.toUpperCase()}:`));
    console.log(color(`  Passed: ${passed}`) + chalk.red(`  Failed: ${failed}`));

    for (const result of serviceResult.results) {
      printResultDetails(result, options.debug);
    }
    console.log("");
  }

  console.log(chalk.bold("=".repeat(60)));
  const total = totalPassed + totalFailed;
  console.log(chalk.bold(`Total: ${totalPassed}/${total} passed`));
  if (totalFailed > 0) {
    console.log(chalk.red(`Failed: ${totalFailed}`));
  } else {
    console.log(chalk.green("All tests passed!"));
  }
  console.log(chalk.bold("=".repeat(60)));

  console.log(chalk.bold("\nGenerating report..."));

  const report = generateReport(allResults, options.plateId, options.apiKey, options.debug);
  const timestamp = new Date().toISOString().replace(/[:.]/g, "-").slice(0, 19);
  const reportDir = options.reportDir || ".";

  if (!existsSync(reportDir)) {
    mkdirSync(reportDir, { recursive: true });
  }

  const filename = `${reportDir}/test-report-${options.plateId}-${timestamp}.txt`;

  try {
    writeFileSync(filename, report);
    console.log(chalk.green(`Report saved to: ${filename}`));

    const shouldOpen = options.openReport || (!options.noOpen && !options.nonInteractive);
    if (shouldOpen) {
      const editor = process.env.EDITOR || process.env.VISUAL || "nano";
      console.log(chalk.gray(`Opening with ${editor}...`));
      spawn(editor, [filename], { stdio: "inherit" });
    }
  } catch (err) {
    console.log(chalk.red(`Failed to save report: ${err}`));
  }
}

main().catch(console.error);
