"use client";

import { useEffect, useState } from "react";

const ws = new WebSocket(`ws://localhost:8080/ws/123`);

ws.onopen = () => {
  console.log("Connected to room:123");
};

ws.onclose = () => {
  console.log("Disconnected from WebSocket");
};
export default function Home() {
  const [messages, setMessages] = useState<string[]>([]);
  ws.onmessage = (event) => {
    setMessages((prev) => [...prev, event.data]);
  };

  useEffect(() => {
    setInterval(() => {
      ws.send(new Date().toISOString());
    }, 1000);
  }, []);
  return (
    <div className="grid grid-rows-[20px_1fr_20px] items-center justify-items-center min-h-screen p-8 pb-20 gap-16 sm:p-20 font-[family-name:var(--font-geist-sans)]">
      {JSON.stringify(messages)}
    </div>
  );
}
