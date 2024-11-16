"use client";

import { ChangeEvent, useState } from "react";

export default function Home() {
  const [room, setRoom] = useState<string>("");
  const [messages, setMessages] = useState<string[]>([]);
  const [ws, setWs] = useState<WebSocket | undefined>();
  const onChangeRoom = (e: ChangeEvent<HTMLInputElement>) => {
    setRoom(e.target.value);
  };
  const applyRoomChange = () => {
    if (ws) {
      ws.close();
    }
    if (room.length == 0) {
      return;
    }
    const newWs = new WebSocket(
      `${process.env.NEXT_PUBLIC_BACKEND_URL}/ws/${room}`
    );
    newWs.onopen = () => {
      console.log("Connected to room:" + room);
    };
    newWs.onclose = () => {
      console.log("Disconnected from WebSocket");
    };
    newWs.onmessage = (event) => {
      setMessages((prev) => [...prev, event.data]);
    };
    setInterval(() => {
      newWs.send(new Date().toISOString());
    }, 1000);
    setWs(newWs);
  };
  return (
    <div className="grid grid-rows-[20px_1fr_20px] items-center justify-items-center min-h-screen p-8 pb-20 gap-16 sm:p-20 font-[family-name:var(--font-geist-sans)]">
      <input type="text" value={room} onChange={onChangeRoom}></input>
      <input type="button" value="変更" onClick={applyRoomChange}></input>
      <div>{`RoomId: ${room}`}</div>
      {JSON.stringify(messages)}
    </div>
  );
}
