"use client";

import clsx from "clsx";
import { useParams } from "next/navigation";
import { useCallback, useEffect, useState } from "react";
import CardComponent from "./CardComponent";

type CardType = "Club" | "Spade" | "Heart" | "Diamond" | "Joker";

type AddPlayerResponse = {
  type: "ADD_PLAYER";
  data: { playerNames: string[] };
};
type GameStartResponse = {
  type: "GAME_START";
  data: {
    handCards: Card[];
    players: Player[];
  };
};
type MessageResponse = { type: "MESSAGE"; data: { message: string } };
type Response = AddPlayerResponse | GameStartResponse | MessageResponse;

type Card = {
  number: number;
  value: number;
  cardType: CardType;
};

type Player = {
  name: string;
  numHandCards: number;
};

export default function Page() {
  const { room, player: playerName } = useParams();
  //const [room, setRoom] = useState<string>("");
  const [messages, setMessages] = useState<string[]>([]);
  const [debugMessages, setDebugMessages] = useState<string[]>([]);
  const [ws, setWs] = useState<WebSocket | undefined>();
  const [gameState, setGameState] = useState<
    "WaitingForPlayers" | "PlayingCards" | "GameEnded"
  >("WaitingForPlayers");

  const [handCards, setHandCards] = useState<Card[]>([]);
  const [selectedCards, setSelectedCards] = useState<Set<Card>>(new Set());
  const [turn, setTurn] = useState<number>(0);
  const [isEnteredRoom, setIsEnteredRoom] = useState<boolean>(false);
  const [players, setPlayers] = useState<Player[]>([]);
  const currentPlayer = players.length == 0 ? undefined : players[turn].name;

  const handleData = useCallback(
    (responseStr: string) => {
      const response = JSON.parse(responseStr) as Response;
      if (response.type === "ADD_PLAYER") {
        const playerNamesToAdd = response.data.playerNames;
        setPlayers(
          playerNamesToAdd.map((name) => {
            return { name, numHandCards: 0 };
          })
        );
      } else if (response.type === "GAME_START") {
        const gameStartData = response.data;
        console.log(gameStartData);
        setGameState("PlayingCards");
        setTurn(0);
        setHandCards(
          gameStartData.handCards.sort((a, b) => {
            if (a.value === b.value) {
              return a.cardType.localeCompare(b.cardType);
            }
            return a.value - b.value;
          })
        );
        setPlayers(gameStartData.players);
      } else if (response.type === "MESSAGE") {
        setMessages((prev) => [...prev, response.data.message]);
      } else {
        console.log("unknown response type");
      }
    },
    [setPlayers, setGameState, setTurn, setHandCards]
  );

  useEffect(() => {
    const scheme = process.env.NODE_ENV === "development" ? "ws" : "wss";
    const ws = new WebSocket(
      `${scheme}://${process.env.NEXT_PUBLIC_BACKEND_DOMAIN}/daifugo/ws/rooms/${room}/${playerName}`
    );
    setWs(ws);
    ws.onopen = () => {
      console.log("Connected to room:" + room);
    };
  }, []);

  useEffect(() => {
    if (!ws) {
      return;
    }
    ws.onmessage = (event) => {
      handleData(event.data);
      setDebugMessages((prev) => [...prev, event.data]);
    };
    ws.onclose = () => {
      ws.send(JSON.stringify({ type: "REMOVE_PLAYER", data: { playerName } }));
      ws.close();
      console.log("Disconnected from WebSocket");
    };
  }, [ws, playerName, handleData]);
  const applyPlayerNameChange = (ws: WebSocket) => {
    setIsEnteredRoom(true);
    ws.send(JSON.stringify({ type: "ADD_PLAYER", data: { playerName } }));
  };
  if (gameState === "PlayingCards") {
    return (
      <div>
        {players.map((player) => {
          return (
            <div key={player.name} className="m-4">
              <div
                className={clsx(currentPlayer === player.name && "font-bold")}
              >
                {player.name}
              </div>
              <div>カード枚数: {player.numHandCards}</div>
            </div>
          );
        })}
        {handCards.map((card) => {
          return (
            <CardComponent
              key={card.number + card.cardType}
              number={card.number}
              cardType={card.cardType}
              isSelected={selectedCards.has(card)}
              handleClick={(number, cardType) => {
                const newSelectedCards = new Set(selectedCards);
                const foundCard = handCards.find(
                  (card) => card.number === number && card.cardType === cardType
                );
                if (foundCard == undefined) {
                  return;
                }
                if (newSelectedCards.has(foundCard)) {
                  newSelectedCards.delete(foundCard);
                } else {
                  newSelectedCards.add(foundCard);
                }
                setSelectedCards(newSelectedCards);
              }}
            />
          );
        })}
        <button
          value="カードを出す"
          onClick={() => {
            ws?.send(
              JSON.stringify({
                type: "SUBMIT_CARDS",
                data: { playerName, cards: Array.from(selectedCards) },
              })
            );
          }}
        >
          カードを出す
        </button>

        {messages.reverse().map((message, idx) => (
          <div key={idx}>{message}</div>
        ))}
      </div>
    );
  }
  return (
    <div className="grid grid-rows-[20px_1fr_20px] items-center justify-items-center min-h-screen p-8 pb-20 gap-16 sm:p-20 font-[family-name:var(--font-geist-sans)]">
      {/* <input
        type="text"
        value={playerName}
        disabled={isEnteredRoom}
        onChange={(e) => setPlayerName(e.target.value)}
      /> */}
      <input
        type="button"
        value="入室"
        disabled={playerName === "" || isEnteredRoom}
        onClick={() => applyPlayerNameChange(ws!)}
      ></input>
      <ul>
        {players.map((op) => {
          return <li key={op.name}>{op.name}</li>;
        })}
      </ul>
      <div>{`RoomId: ${room}`}</div>
      <input
        type="button"
        disabled={players.length <= 1}
        onClick={() => {
          ws?.send(JSON.stringify({ type: "GAME_START" }));
        }}
        value="GAME START"
      />

      {JSON.stringify(debugMessages)}
    </div>
  );
}
