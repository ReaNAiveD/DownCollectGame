import React from 'react';
import { CardDisplay } from './Card';

export default function Results({ gameState }) {
  const gs = gameState;
  if (!gs || !gs.scores) return null;

  const winners = gs.winners || [];

  return (
    <div className="results">
      <h2>游戏结束</h2>
      <div className="winner">
        🏆 获胜者: {winners.map(id => {
          const p = gs.players?.find(pl => pl.playerId === id);
          return p?.nickname || id;
        }).join(', ')}
      </div>

      {gs.scores
        .sort((a, b) => a.totalScore - b.totalScore)
        .map((ps) => {
          const player = gs.players?.find(p => p.playerId === ps.playerId);
          const isWinner = winners.includes(ps.playerId);
          return (
            <div key={ps.playerId} className="player-result" style={isWinner ? { borderLeft: '4px solid #4ecdc4' } : {}}>
              <h3>
                <span>{player?.nickname || `座位 ${ps.seatIndex + 1}`} {isWinner && '🏆'}</span>
                <span style={{ color: '#e94560' }}>总分: {ps.totalScore}</span>
              </h3>
              <div style={{ display: 'flex', gap: 4, margin: '8px 0', flexWrap: 'wrap' }}>
                {ps.hand?.map((card, i) => (
                  <CardDisplay key={i} card={card} faceUp={true} />
                ))}
              </div>
              <div className="score-details">
                {ps.details?.map((d, i) => (
                  <div key={i} className="score-detail">
                    <span>
                      {d.card.rank}{d.card.suit !== 'None' ? ` (${d.card.suit})` : ''}
                      {' '} — {d.ruleName}
                      {d.skipped && ' [跳过]'}
                    </span>
                    <span className="score-value">
                      {d.skipped ? '-' : `${d.score >= 0 ? '+' : ''}${d.score}`}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          );
        })}
    </div>
  );
}
