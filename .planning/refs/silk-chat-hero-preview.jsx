import React, { useEffect, useRef, useState } from 'react';

export default function SilkChatHero() {
  const canvasRef = useRef(null);
  const animationRef = useRef();
  const [isLoaded, setIsLoaded] = useState(false);

  useEffect(() => {
    const timer = setTimeout(() => setIsLoaded(true), 300);
    return () => clearTimeout(timer);
  }, []);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    let time = 0;
    const speed = 0.02;
    const scale = 2;
    const noiseIntensity = 0.8;

    const resizeCanvas = () => {
      const parent = canvas.parentElement;
      canvas.width = parent.clientWidth;
      canvas.height = parent.clientHeight;
    };

    resizeCanvas();
    window.addEventListener('resize', resizeCanvas);

    const noise = (x, y) => {
      const G = 2.71828;
      const rx = G * Math.sin(G * x);
      const ry = G * Math.sin(G * y);
      return (rx * ry * (1 + x)) % 1;
    };

    const animate = () => {
      const { width, height } = canvas;

      const gradient = ctx.createLinearGradient(0, 0, width, height);
      gradient.addColorStop(0, '#1a1a1a');
      gradient.addColorStop(0.5, '#2a2a2a');
      gradient.addColorStop(1, '#1a1a1a');

      ctx.fillStyle = gradient;
      ctx.fillRect(0, 0, width, height);

      const imageData = ctx.createImageData(width, height);
      const data = imageData.data;

      for (let x = 0; x < width; x += 2) {
        for (let y = 0; y < height; y += 2) {
          const u = (x / width) * scale;
          const v = (y / height) * scale;

          const tOffset = speed * time;
          let tex_x = u;
          let tex_y = v + 0.03 * Math.sin(8.0 * tex_x - tOffset);

          const pattern = 0.6 + 0.4 * Math.sin(
            5.0 * (tex_x + tex_y +
              Math.cos(3.0 * tex_x + 5.0 * tex_y) +
              0.02 * tOffset) +
            Math.sin(20.0 * (tex_x + tex_y - 0.1 * tOffset))
          );

          const rnd = noise(x, y);
          const intensity = Math.max(0, pattern - (rnd / 15.0) * noiseIntensity);

          const r = Math.floor(123 * intensity);
          const g = Math.floor(116 * intensity);
          const b = Math.floor(129 * intensity);

          const index = (y * width + x) * 4;
          if (index < data.length) {
            data[index] = r;
            data[index + 1] = g;
            data[index + 2] = b;
            data[index + 3] = 255;
          }
        }
      }

      ctx.putImageData(imageData, 0, 0);

      const overlayGradient = ctx.createRadialGradient(
        width / 2, height / 2, 0,
        width / 2, height / 2, Math.max(width, height) / 2
      );
      overlayGradient.addColorStop(0, 'rgba(0, 0, 0, 0.1)');
      overlayGradient.addColorStop(1, 'rgba(0, 0, 0, 0.4)');

      ctx.fillStyle = overlayGradient;
      ctx.fillRect(0, 0, width, height);

      time += 1;
      animationRef.current = requestAnimationFrame(animate);
    };

    animate();

    return () => {
      window.removeEventListener('resize', resizeCanvas);
      if (animationRef.current) cancelAnimationFrame(animationRef.current);
    };
  }, []);

  const fadeStyle = (delayMs) => ({
    opacity: isLoaded ? 1 : 0,
    transform: isLoaded ? 'translateY(0)' : 'translateY(12px)',
    transition: `opacity 0.9s ease-out ${delayMs}ms, transform 0.9s ease-out ${delayMs}ms`,
  });

  return (
    <div style={{ position: 'relative', height: '600px', width: '100%', overflow: 'hidden', background: '#000', fontFamily: '-apple-system, BlinkMacSystemFont, Segoe UI, Inter, sans-serif' }}>
      <canvas ref={canvasRef} style={{ position: 'absolute', inset: 0, width: '100%', height: '100%', zIndex: 0 }} />

      <div style={{ position: 'absolute', inset: 0, zIndex: 1, background: 'linear-gradient(to bottom, rgba(0,0,0,0.4), rgba(0,0,0,0.1), rgba(0,0,0,0.6))' }} />

      <div style={{ position: 'relative', zIndex: 2, display: 'flex', height: '100%', alignItems: 'center', justifyContent: 'center' }}>
        <div style={{ width: '100%', maxWidth: '460px', padding: '0 24px', textAlign: 'center' }}>
          <div
            style={{
              ...fadeStyle(0),
              width: '56px',
              height: '56px',
              margin: '0 auto 22px',
              borderRadius: '14px',
              background: 'rgba(255,255,255,0.06)',
              border: '1px solid rgba(255,255,255,0.14)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              backdropFilter: 'blur(4px)',
            }}
          >
            <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="#f1f0f4" strokeWidth="1.6">
              <path d="M4 14a8 8 0 0 1 14.5-4.6M20 10a8 8 0 0 1-14.5 4.6" strokeLinecap="round" />
            </svg>
          </div>

          <h1 style={{ ...fadeStyle(0), margin: '0 0 4px', fontSize: '26px', fontWeight: 400, letterSpacing: '-0.01em', color: 'rgba(230,228,235,0.7)' }}>
            Good to see you!
            <strong style={{ display: 'block', fontWeight: 500, color: '#f5f4f8', fontSize: '28px', marginTop: '2px' }}>
              How can I be an assistance?
            </strong>
          </h1>

          <p style={{ ...fadeStyle(200), margin: '14px 0 32px', fontSize: '13.5px', color: 'rgba(230,228,235,0.45)' }}>
            I'm available 24/7 for you, ask me anything.
          </p>

          <div
            style={{
              ...fadeStyle(400),
              background: 'rgba(255,255,255,0.05)',
              border: '1px solid rgba(255,255,255,0.14)',
              borderRadius: '16px',
              padding: '14px',
              backdropFilter: 'blur(10px)',
            }}
          >
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', fontSize: '11.5px', color: 'rgba(230,228,235,0.55)', marginBottom: '12px', padding: '0 2px' }}>
              <span>Unlock more features with the Pro plan.</span>
              <span>
                <span style={{ display: 'inline-block', width: '6px', height: '6px', borderRadius: '50%', background: '#b7f5c8', marginRight: '6px' }} />
                Active extensions
              </span>
            </div>
            <div style={{ display: 'flex', alignItems: 'center', gap: '10px', background: 'rgba(255,255,255,0.05)', border: '1px solid rgba(255,255,255,0.14)', borderRadius: '10px', padding: '10px 12px' }}>
              <span style={{ color: 'rgba(230,228,235,0.5)' }}>+</span>
              <input
                type="text"
                placeholder="Ask anything ..."
                style={{ flex: 1, background: 'none', border: 'none', outline: 'none', color: '#f1f0f4', fontSize: '13.5px', fontFamily: 'inherit' }}
              />
              <span style={{ color: 'rgba(230,228,235,0.5)', fontSize: '12px' }}>|||</span>
            </div>
          </div>

          <div style={{ ...fadeStyle(500), display: 'flex', flexWrap: 'wrap', gap: '8px', justifyContent: 'center', marginTop: '16px' }}>
            {['Any advice for me?', 'Some youtube video idea', 'Life lessons from kratos', '···'].map((label) => (
              <div
                key={label}
                style={{
                  padding: '8px 13px',
                  borderRadius: '18px',
                  border: '1px solid rgba(255,255,255,0.14)',
                  background: 'rgba(255,255,255,0.04)',
                  color: 'rgba(230,228,235,0.75)',
                  fontSize: '12px',
                  whiteSpace: 'nowrap',
                }}
              >
                {label}
              </div>
            ))}
          </div>
        </div>
      </div>

      <div style={{ position: 'absolute', bottom: '18px', left: 0, right: 0, textAlign: 'center', fontSize: '11.5px', color: 'rgba(230,228,235,0.4)', zIndex: 2 }}>
        Unlock new era with AetherAI.{' '}
        <span style={{ color: 'rgba(230,228,235,0.7)', textDecoration: 'underline' }}>share us</span>
      </div>
    </div>
  );
}
