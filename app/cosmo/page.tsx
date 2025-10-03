"use client";

import { useEffect, useState } from "react";

type TokenInfo = {
  mint: string;
  name: string;
  symbol: string;
  logo: string;
};

export default function CosmoPage() {
  const [tokens, setTokens] = useState<TokenInfo[]>([]);
  const [connected, setConnected] = useState(false);

  useEffect(() => {
    const ws = new WebSocket("ws://127.0.0.1:8080/connect");

    ws.onopen = () => {
      console.log("Connected to backend WebSocket");
      setConnected(true);
    };

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        console.log("New token:", data);

        setTokens((prev) => [data, ...prev].slice(0, 50)); // keep latest 50
      } catch (err) {
        console.error("Error parsing message:", err);
      }
    };

    ws.onclose = () => {
      console.log("WebSocket closed");
      setConnected(false);
    };

    return () => {
      ws.close();
    };
  }, []);

  return (
    <div style={{ padding: "20px", fontFamily: "Arial, sans-serif" }}>
      <h1>🚀 Cosmo - Live New Tokens Feed</h1>
      <p>Status: {connected ? "🟢 Connected" : "🔴 Disconnected"}</p>

      <div style={{ marginTop: "20px" }}>
        {tokens.length === 0 ? (
          <p>No tokens received yet...</p>
        ) : (
          <ul style={{ listStyle: "none", padding: 0 }}>
            {tokens.map((t, idx) => (
              <li
                key={idx}
                style={{
                  border: "1px solid #ccc",
                  borderRadius: "6px",
                  padding: "10px",
                  marginBottom: "10px",
                  display: "flex",
                  alignItems: "center",
                  gap: "10px",
                }}
              >
                {t.logo && (
                  <img
                    src={t.logo}
                    alt={t.symbol}
                    style={{ width: "32px", height: "32px", borderRadius: "50%" }}
                  />
                )}
                <div>
                  <strong>{t.name}</strong> ({t.symbol})
                  <br />
                  <small>{t.mint}</small>
                </div>
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  );
}
