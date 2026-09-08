import { cn } from "@/lib/utils";

/*
 * VULX monogram: a "VX" drawn as strokes rather than set as type, so it stays
 * crisp at any tile size and does not shift while the webfont loads. The tile
 * uses opaque tokens (--surface-2 on a --hairline border).
 *
 * Geometry is tuned for small sizes. The V is 9 wide and the X 8, so the two
 * letters read as siblings — the first draft paired a 9-wide V with a 6-wide X,
 * which is what made it look lopsided. Cap height is 12 of 24 (up from 10) so
 * the mark fills the tile, and the stroke is 1.7 rather than 1.8 to keep the X's
 * crossing open instead of blobbing into a solid diamond at 20px.
 */

export function VxMark({ className }: { className?: string }) {
  return (
    <span
      className={cn(
        "border-hairline bg-surface-2 flex size-7 shrink-0 items-center justify-center rounded-lg border",
        className
      )}
      aria-hidden
    >
      <svg
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth={1.7}
        strokeLinecap="round"
        strokeLinejoin="round"
        className="text-foreground size-5"
        focusable="false"
      >
        <path d="M3 6 7.5 18 12 6" />
        <path d="M13 6 21 18M21 6 13 18" />
      </svg>
    </span>
  );
}
