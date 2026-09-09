/*
 * Relative timestamps for the project list. Input is the RFC3339 string the API
 * sends (Project.created_at / updated_at), not a Date.
 */

const rtf = new Intl.RelativeTimeFormat("en", { numeric: "auto" });

const MINUTE = 60_000;
const HOUR = 60 * MINUTE;
const DAY = 24 * HOUR;
const WEEK = 7 * DAY;

export function formatRelative(rfc3339: string): string {
  const then = new Date(rfc3339).getTime();
  if (Number.isNaN(then)) return "";

  const diff = Date.now() - then;
  if (diff < MINUTE) return "Just now";
  if (diff < HOUR) return rtf.format(-Math.floor(diff / MINUTE), "minute");
  if (diff < DAY) return rtf.format(-Math.floor(diff / HOUR), "hour");
  if (diff < WEEK) return rtf.format(-Math.floor(diff / DAY), "day");

  return new Date(then).toLocaleDateString(undefined, {
    month: "short",
    day: "numeric",
  });
}
