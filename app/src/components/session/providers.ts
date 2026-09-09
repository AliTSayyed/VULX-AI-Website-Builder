import { AiProvider } from "@/gen/api/v1/enums_pb";

/*
 * Labels diverge from wire values on purpose. The AI service's routes are the
 * path segments openai / google / anthropic — it rejects "gemini" and "claude" —
 * but nobody calls the model "google". The Go API maps the enum to the segment;
 * the browser only ever sends the enum.
 */
export const PROVIDERS: { value: AiProvider; label: string }[] = [
  { value: AiProvider.ANTHROPIC, label: "Claude" },
  { value: AiProvider.OPENAI, label: "OpenAI" },
  { value: AiProvider.GOOGLE, label: "Gemini" },
];

export const DEFAULT_PROVIDER = AiProvider.OPENAI;

/** A project created before a provider was recorded, or an unknown value, must
 *  still render something selectable rather than an empty trigger. */
export function providerOrDefault(p: AiProvider | undefined): AiProvider {
  return p && p !== AiProvider.UNSPECIFIED ? p : DEFAULT_PROVIDER;
}
