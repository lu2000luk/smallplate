"use client";

import { useState } from "react";

import { Tabs, TabsList, TabsTab } from "@/components/ui/tabs";
import { cn } from "@/lib/utils";

export type CodeLanguage = "javascript" | "curl" | "go" | "rust";

const languageLabels: Record<CodeLanguage, string> = {
  javascript: "JS",
  curl: "cURL",
  go: "Go",
  rust: "Rust",
};

interface CodeBlockProps {
  code: Record<CodeLanguage, string>;
  className?: string;
}

export function CodeBlock({ code, className }: CodeBlockProps) {
  const [language, setLanguage] = useState<CodeLanguage>("javascript");

  return (
    <div className={cn("rounded-lg border bg-muted/50 overflow-hidden", className)}>
      <Tabs
        value={language}
        onValueChange={(v) => setLanguage(v as CodeLanguage)}
        className="w-full"
      >
        <TabsList variant="default" className="absolute right-3 top-3 z-10 h-8 bg-muted/80 backdrop-blur-sm">
          {(Object.keys(code) as CodeLanguage[]).map((lang) => (
            <TabsTab key={lang} value={lang} className="h-6 px-2 text-xs">
              {languageLabels[lang]}
            </TabsTab>
          ))}
        </TabsList>
      </Tabs>
      <pre className="overflow-x-auto p-4 pt-10 text-sm">
        <code>{code[language]}</code>
      </pre>
    </div>
  );
}
