import React from 'react';
import { CardDisplay } from './Card';
import { getCardInfo } from '../cardInfo';

function CardInfoPanel({ title, card, scoringMode }) {
  if (!card) return <div className="info-panel empty"><p className="info-empty">{title}：无</p></div>;

  const info = getCardInfo(card.rank);
  if (!info) return <div className="info-panel empty"><p className="info-empty">{title}：未知卡牌</p></div>;

  const scoring = info.scoring[scoringMode] || info.scoring[0];

  return (
    <div className="info-panel">
      <div className="info-panel-header">{title}</div>
      <div className="info-card-display">
        <CardDisplay card={card} faceUp={true} />
        <div className="info-card-meta">
          <span className="info-card-name">{info.name}</span>
          <span className="info-card-base">基础分值: {info.baseScore}</span>
        </div>
      </div>
      <div className="info-section">
        <div className="info-section-title">揭示效果</div>
        <p className="info-text">{info.effect}</p>
      </div>
      <div className="info-section">
        <div className="info-section-title">
          计分规则 — {scoring.name}
          {scoringMode === 1 && <span className="info-mode-tag">仅基础分</span>}
        </div>
        <p className="info-text">{scoring.desc}</p>
      </div>
    </div>
  );
}

export default function CardSidebar({ revealedCard, hoveredCard, scoringMode }) {
  return (
    <div className="card-sidebar">
      <CardInfoPanel title="揭示卡牌" card={revealedCard} scoringMode={scoringMode} />
      <CardInfoPanel title="卡牌详情" card={hoveredCard} scoringMode={scoringMode} />
    </div>
  );
}
