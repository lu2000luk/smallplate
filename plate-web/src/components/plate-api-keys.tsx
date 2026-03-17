"use client";

import { CopyIcon, PlusIcon, Trash2Icon } from "lucide-react";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import type { PlateApiKey } from "./plate-dashboard";

type PlateApiKeysProps = {
  apiKeys: PlateApiKey[];
  isLoading: boolean;
  error: string | null;
  isCreating: boolean;
  deletingKeyId: number | null;
  createdKey: string | null;
  copiedKey: boolean;
  onCreate: () => void;
  onDelete: (apiKeyId: number) => void;
  onCopy: () => void;
};

function formatCreatedAt(timestamp: number): string {
  return new Date(timestamp).toLocaleString();
}

export function PlateApiKeys({
  apiKeys,
  isLoading,
  error,
  isCreating,
  deletingKeyId,
  createdKey,
  copiedKey,
  onCreate,
  onDelete,
  onCopy,
}: PlateApiKeysProps) {
  return (
    <div className="space-y-4">
      <div className="flex items-start justify-between gap-4">
        <div>
          <p className="text-sm text-muted-foreground">Manage</p>
          <h2 className="text-2xl font-semibold">API keys</h2>
        </div>
        <Button onClick={onCreate} disabled={isCreating}>
          {isCreating ? <Spinner /> : <PlusIcon />}
          Create API key
        </Button>
      </div>

      {createdKey ? (
        <Alert variant="success">
          <AlertTitle>Copy this key now</AlertTitle>
          <AlertDescription>
            <div className="space-y-3">
              <p>You will not be able to view this full API key again.</p>
              <div className="rounded-lg border bg-background px-3 py-2 font-mono text-xs break-all">
                {createdKey}
              </div>
              <Button variant="outline" onClick={onCopy}>
                <CopyIcon />
                {copiedKey ? "Copied" : "Copy key"}
              </Button>
            </div>
          </AlertDescription>
        </Alert>
      ) : null}

      {error ? (
        <p className="text-sm text-destructive-foreground">{error}</p>
      ) : null}

      {isLoading ? (
        <div className="flex py-12 items-center justify-center">
          <Spinner />
        </div>
      ) : apiKeys.length === 0 ? (
        <div className="py-12 text-center text-sm text-muted-foreground border border-dashed rounded-lg">
          No API keys yet. Create a key to connect clients to this plate.
        </div>
      ) : (
        <div className="divide-y rounded-lg border max-w-3xl">
          {apiKeys.map((apiKey) => (
            <div
              key={apiKey.id}
              className="flex flex-col gap-3 p-4 sm:flex-row sm:items-center sm:justify-between bg-card hover:bg-muted/30 transition-colors first:rounded-t-lg last:rounded-b-lg"
            >
              <div>
                <p className="font-mono text-sm font-medium">
                  {apiKey.created_at}
                </p>
                <p className="mt-1 text-xs text-muted-foreground">
                  Created {formatCreatedAt(apiKey.created_at)}
                </p>
              </div>
              <Button
                variant="destructive-outline"
                size="sm"
                onClick={() => onDelete(apiKey.id)}
                disabled={deletingKeyId === apiKey.id}
              >
                {deletingKeyId === apiKey.id ? (
                  <Spinner className="size-3" />
                ) : (
                  <Trash2Icon className="size-3" />
                )}
                Delete
              </Button>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
