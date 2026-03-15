import React from 'react';

const SUIT_SYMBOLS = { Hearts: '♥', Diamonds: '♦', Spades: '♠', Clubs: '♣', None: '' };
const SUIT_COLOR = { Hearts: 'red', Diamonds: 'red', Spades: 'black', Clubs: 'black', None: 'black' };

export function cardLabel(c) {
  if (!c) return '?';
  return `${c.rank}${SUIT_SYMBOLS[c.suit] || ''}`;
}

export function CardDisplay({ card, faceUp, selectable, selected, peeked, onClick, small, onHover }) {
  const sizeClass = small ? 'card-sm' : '';
  const displayCard = (faceUp || peeked) ? (peeked || card) : null;
  const hoverHandlers = displayCard && onHover ? {
    onMouseEnter: () => onHover(displayCard),
    onMouseLeave: () => onHover(null),
  } : {};

  if (!faceUp && !peeked) {
    return (
      <div
        className={`card face-down ${sizeClass} ${selectable ? 'selectable' : ''} ${selected ? 'selected' : ''}`}
        onClick={selectable ? onClick : undefined}
      >
        <span className="rank">?</span>
      </div>
    );
  }

  if (!displayCard) {
    return <div className={`card face-down ${sizeClass}`}><span className="rank">?</span></div>;
  }

  const colorClass = SUIT_COLOR[displayCard.suit] || 'black';

  return (
    <div
      className={`card face-up ${colorClass} ${sizeClass} ${selectable ? 'selectable' : ''} ${selected ? 'selected' : ''} ${peeked && !faceUp ? 'peeked' : ''}`}
      onClick={selectable ? onClick : undefined}
      {...hoverHandlers}
    >
      <span className="rank">{displayCard.rank}</span>
      <span className="suit">{SUIT_SYMBOLS[displayCard.suit] || ''}</span>
    </div>
  );
}

export function InsertMarker({ position, onClick }) {
  return (
    <div className="insert-marker" onClick={() => onClick(position)} title={`插入位置 ${position}`}>
      ▼
    </div>
  );
}
