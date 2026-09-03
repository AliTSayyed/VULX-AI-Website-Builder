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
    <div className="mt-6 flex flex-wrap items-center justify-center gap-2">
      {SUGGESTIONS.map((s) => (
        <button
          key={s}
          type="button"
          onClick={onSelect}
          className="border-border text-muted-foreground hover:border-ring/50 hover:text-foreground focus-visible:ring-ring rounded-full border px-4 py-2 text-sm transition-colors focus-visible:ring-2 focus-visible:outline-none"
        >
          {s}
        </button>
      ))}
    </div>
  );
}