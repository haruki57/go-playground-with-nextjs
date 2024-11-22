"use client";

import { ChangeEvent, useState } from "react";

export default function Home() {
  const [room, setRoom] = useState<string>("");
  const [rooms, setRooms] = useState<string>([]);
  //const [messages, setMessages] = useState<string[]>([]);
  const onChangeRoom = (e: ChangeEvent<HTMLInputElement>) => {
    setRoom(e.target.value);
  };
  const fetchRoom = () => {
    const _ = async () => {
      const scheme = process.env.NODE_ENV === "development" ? "http" : "https";
      const ret = await fetch(
        `${scheme}://${process.env.NEXT_PUBLIC_BACKEND_DOMAIN}/daifugo/rooms`
      );
      const json = await ret.json();
      setRooms(json);
    };
    _();
  };

  const createRoom = () => {
    const _ = async () => {
      const scheme = process.env.NODE_ENV === "development" ? "http" : "https";
      await fetch(
        `${scheme}://${process.env.NEXT_PUBLIC_BACKEND_DOMAIN}/daifugo/rooms/${room}`,
        { method: "POST" }
      );
    };
    _();
  };

  return (
    <div className="grid grid-rows-[20px_1fr_20px] items-center justify-items-center min-h-screen p-8 pb-20 gap-16 sm:p-20 font-[family-name:var(--font-geist-sans)]">
      <input type="button" value="部屋一覧取得" onClick={fetchRoom}></input>
      {JSON.stringify(rooms)}
      <input type="text" value={room} onChange={onChangeRoom}></input>
      <input type="button" value="部屋作成" onClick={createRoom}></input>
    </div>
  );
}
