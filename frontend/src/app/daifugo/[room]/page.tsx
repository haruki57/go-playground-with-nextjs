"use client";

import { useParams } from "next/navigation";
import { ChangeEvent, useEffect, useState } from "react";

type AddPlayerData = { type: "ADD_PLAYER"; data: { playerName: string } };
type NumberData = { type: "number"; value: number };
type BooleanData = { type: "boolean"; value: boolean };

type Response = AddPlayerData | NumberData | BooleanData;

type OtherPlayer = {
  name: string;
  numHandCards: number;
};

export default function Page() {
  const { room } = useParams();
  //const [room, setRoom] = useState<string>("");
  const [messages, setMessages] = useState<string[]>([]);
  const [ws, setWs] = useState<WebSocket | undefined>();
  const [playerName, setPlayerName] = useState<string>("");
  const [otherPlayers, setOtherPlayers] = useState<OtherPlayer[]>([]);

  const handleData = (response: Response) => {
    if (response.type === "ADD_PLAYER") {
      setOtherPlayers([
        ...otherPlayers,
        { name: response.data.playerName, numHandCards: 0 },
      ]);
    }
  };

  useEffect(() => {
    const scheme = process.env.NODE_ENV === "development" ? "ws" : "wss";
    const ws = new WebSocket(
      `${scheme}://${process.env.NEXT_PUBLIC_BACKEND_DOMAIN}/daifugo/ws/rooms/${room}`
    );
    setWs(ws);
    ws.onopen = () => {
      console.log("Connected to room:" + room);
    };
    ws.onclose = () => {
      ws.close();
      console.log("Disconnected from WebSocket");
    };
    ws.onmessage = (event) => {
      handleData(event.data);
      setMessages((prev) => [...prev, event.data]);
    };
  }, []);
  const applyPlayerNameChange = (ws: WebSocket) => {
    ws.send(JSON.stringify({ type: "ADD_PLAYER", data: { playerName } }));
  };
  return (
    <div className="grid grid-rows-[20px_1fr_20px] items-center justify-items-center min-h-screen p-8 pb-20 gap-16 sm:p-20 font-[family-name:var(--font-geist-sans)]">
      <input
        type="text"
        value={playerName}
        onChange={(e) => setPlayerName(e.target.value)}
      />
      <input
        type="button"
        value="入室"
        onClick={() => applyPlayerNameChange(ws!)}
      ></input>
      <div>{`RoomId: ${room}`}</div>
      {JSON.stringify(messages)}
    </div>
  );
}
