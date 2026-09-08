"use client";

const SUGGESTIONS = [
  "A portfolio for a photographer",
  "A pricing page with three tiers",
  "A landing page for a SaaS app",
];

type SuggestionChipsProps = {
  onSelect: () => void;
};

export function SuggestionChips({ onSelect }: SuggestionChipsProps) {
  return (
    <div
      className="vx-rise mt-4 flex flex-wrap items-center justify-center gap-2"
      style={{ animationDelay: "500ms" }}
    >
      {SUGGESTIONS.map((s) => (
        <button
          key={s}
          type="button"
          onClick={onSelect}
          className="border-hairline bg-surface text-foreground-dim hover:border-hairline-strong hover:bg-surface-2 hover:text-foreground focus-visible:ring-ring rounded-full border px-3.5 py-2 text-xs whitespace-nowrap transition-colors focus-visible:ring-2 focus-visible:outline-none"
        >
          {s}
        </button>
      ))}
    </div>
  );
}
