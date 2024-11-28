"use client";

import clsx from "clsx";
import { resolveSoa } from "dns";
import { useParams } from "next/navigation";
import { ChangeEvent, useCallback, useEffect, useState } from "react";

type CardType = "Club" | "Spade" | "Heart" | "Diamond" | "Joker";

type AddPlayerData = { type: "ADD_PLAYER"; data: { playerNames: string[] } };
type GameStartData = {
  type: "GAME_START";
  data: {
    handCards: Card[];
    players: Player[];
  };
};
type BooleanData = { type: "boolean"; value: boolean };

type Response = AddPlayerData | GameStartData | BooleanData;

type Card = {
  number: number;
  value: number;
  cardType: CardType;
};

type Player = {
  name: string;
  numHandCards: number;
};

const map = {
  Diamond: "♦️",
  Heart: "♥️",
  Club: "♣️",
  Spade: "♠️",
  Joker: "🃏",
} as const;

export default function CardComponent({
  number,
  cardType,
  isSelected = false,
  handleClick,
}: {
  number: number;
  cardType: CardType;
  isSelected?: boolean;
  handleClick?: (number: number, cardType: CardType) => void;
}) {
  const cardColor =
    cardType === "Diamond" || cardType === "Heart" ? "text-red-500" : "";
  return (
    <div
      className={clsx(
        "flex",
        "text-2xl",
        "gap-2",
        cardColor,
        isSelected && "font-bold"
      )}
      onClick={() => {
        if (handleClick) {
          handleClick(number, cardType);
        }
      }}
    >
      <div>{number}</div>
      <div>{map[cardType]}</div>
      {isSelected && <div>selected</div>}
    </div>
  );
}
