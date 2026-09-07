export function Hero() {
  return (
    <div className="flex flex-col items-center text-center">
      <h1 className="vx-rise text-balance">
        <span className="text-foreground-dim block text-4xl font-normal tracking-tight sm:text-5xl">
          VULX is your personal
        </span>
        <strong className="text-foreground mt-1 block text-5xl font-medium tracking-tight sm:text-6xl">
          AI Website Creator
        </strong>
      </h1>

      <p
        className="text-muted-foreground vx-rise mt-5 max-w-md text-sm text-balance"
        style={{ animationDelay: "200ms" }}
      >
        Describe what you want. Watch it build itself.
      </p>
    </div>
  );
}
