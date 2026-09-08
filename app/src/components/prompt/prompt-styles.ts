/*
 * The prompt box is the one element shared between the logged-out hero and the
 * logged-in home, and the two must stay visually identical. Keeping the class
 * strings here rather than duplicating them is what stops that drifting.
 * See design-system.md §3 for the tokens.
 */

export const PROMPT_BOX =
  "border-hairline bg-surface w-full rounded-3xl p-3 shadow-none";

export const PROMPT_TEXTAREA =
  "text-foreground min-h-21 bg-transparent text-[15px] leading-6 dark:bg-transparent";
