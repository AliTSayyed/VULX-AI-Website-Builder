"use client";

import { useEffect, useRef } from "react";
import { cn } from "@/lib/utils";

/*
 * Animated silk weave. Ported from .planning/refs/silk-chat-hero-preview.jsx.
 *
 * The reference walks every second pixel of a full-size ImageData on every
 * frame, which is a few million JS iterations per frame at desktop sizes. We
 * evaluate the same field into a quarter-scale buffer and let the compositor
 * upscale it: identical look (the field has no high-frequency detail), ~16x
 * less work. Frames are capped at 30fps, paused while the tab is hidden, and
 * reduced to a single static frame under prefers-reduced-motion.
 */

const SPEED = 0.02;
const SCALE = 2;
const NOISE_INTENSITY = 0.8;
const RESOLUTION = 0.25;
const FRAME_MS = 1000 / 30;

// Silk thread colour at full intensity. Steel blue, replacing the reference's
// violet rgb(123, 116, 129). Luminance-matched to it (Y 0.183, L* 49.8) so the
// weave sits at the same brightness and text contrast is unchanged — raising
// these toward rgb(112, 132, 158) costs ~5 L* and eats into the small type.
const THREAD_R = 102;
const THREAD_G = 120;
const THREAD_B = 144;

function noise(x: number, y: number) {
  const G = 2.71828;
  const rx = G * Math.sin(G * x);
  const ry = G * Math.sin(G * y);
  return (rx * ry * (1 + x)) % 1;
}

export function SilkBackground({ className }: { className?: string }) {
  const canvasRef = useRef<HTMLCanvasElement>(null);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const ctx = canvas.getContext("2d");
    if (!ctx) return;

    const buffer = document.createElement("canvas");
    const bufferCtx = buffer.getContext("2d");
    if (!bufferCtx) return;

    const reducedMotion = window.matchMedia(
      "(prefers-reduced-motion: reduce)"
    ).matches;

    let image: ImageData | null = null;
    let frame = 0;
    let raf = 0;
    let last = 0;

    const resize = () => {
      const parent = canvas.parentElement;
      const width = Math.max(1, parent?.clientWidth ?? window.innerWidth);
      const height = Math.max(1, parent?.clientHeight ?? window.innerHeight);

      canvas.width = width;
      canvas.height = height;
      buffer.width = Math.max(1, Math.round(width * RESOLUTION));
      buffer.height = Math.max(1, Math.round(height * RESOLUTION));
      image = bufferCtx.createImageData(buffer.width, buffer.height);
    };

    const draw = () => {
      if (!image) return;

      const { width: bw, height: bh } = buffer;
      const data = image.data;
      const tOffset = SPEED * frame;

      for (let y = 0; y < bh; y++) {
        const v = (y / bh) * SCALE;

        for (let x = 0; x < bw; x++) {
          const u = (x / bw) * SCALE;

          const texX = u;
          const texY = v + 0.03 * Math.sin(8.0 * texX - tOffset);

          const pattern =
            0.6 +
            0.4 *
              Math.sin(
                5.0 *
                  (texX +
                    texY +
                    Math.cos(3.0 * texX + 5.0 * texY) +
                    0.02 * tOffset) +
                  Math.sin(20.0 * (texX + texY - 0.1 * tOffset))
              );

          const grain = noise(x, y);
          const intensity = Math.max(
            0,
            pattern - (grain / 15.0) * NOISE_INTENSITY
          );

          const i = (y * bw + x) * 4;
          data[i] = THREAD_R * intensity;
          data[i + 1] = THREAD_G * intensity;
          data[i + 2] = THREAD_B * intensity;
          data[i + 3] = 255;
        }
      }

      bufferCtx.putImageData(image, 0, 0);

      const { width, height } = canvas;
      ctx.imageSmoothingEnabled = true;
      ctx.imageSmoothingQuality = "high";
      ctx.clearRect(0, 0, width, height);
      ctx.drawImage(buffer, 0, 0, bw, bh, 0, 0, width, height);

      const vignette = ctx.createRadialGradient(
        width / 2,
        height / 2,
        0,
        width / 2,
        height / 2,
        Math.max(width, height) / 2
      );
      vignette.addColorStop(0, "rgba(0, 0, 0, 0.1)");
      vignette.addColorStop(1, "rgba(0, 0, 0, 0.4)");

      ctx.fillStyle = vignette;
      ctx.fillRect(0, 0, width, height);
    };

    const animate = (now: number) => {
      raf = requestAnimationFrame(animate);
      if (now - last < FRAME_MS) return;
      last = now;
      frame += 1;
      draw();
    };

    const start = () => {
      if (raf || reducedMotion) return;
      last = 0;
      raf = requestAnimationFrame(animate);
    };

    const stop = () => {
      if (!raf) return;
      cancelAnimationFrame(raf);
      raf = 0;
    };

    const onResize = () => {
      resize();
      draw();
    };

    const onVisibility = () => {
      if (document.hidden) stop();
      else start();
    };

    resize();
    draw();
    start();

    window.addEventListener("resize", onResize);
    document.addEventListener("visibilitychange", onVisibility);

    return () => {
      stop();
      window.removeEventListener("resize", onResize);
      document.removeEventListener("visibilitychange", onVisibility);
    };
  }, []);

  return (
    <div className={cn("pointer-events-none absolute inset-0", className)} aria-hidden>
      <canvas ref={canvasRef} className="absolute inset-0 size-full" />
      {/*
       * Legibility scrim. The mid stop carries the load: the top and bottom of the
       * weave are already crushed by the canvas vignette, but the centre band — where
       * the hero text sits directly on silk with no panel under it — is the brightest
       * part of the page. At via-black/10 the subhead measured 2.38:1 against a bright
       * ribbon; at /55 it holds 5.19:1. Do not lighten this stop.
       */}
      <div className="absolute inset-0 bg-gradient-to-b from-black/40 via-black/55 to-black/60" />
    </div>
  );
}
