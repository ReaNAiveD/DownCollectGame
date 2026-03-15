import React from 'react';

const SCORING_LABELS = { 0: '特殊计分规则', 1: '仅基础分值' };

export default function WaitingRoom({ roomInfo, playerId, onStartGame }) {
  if (!roomInfo) return null;

  const players = roomInfo.players || [];
  const isHost = players.length > 0 && players[0]?.playerId === playerId;
  const canStart = players.length >= 2;
  const config = roomInfo.config;

  return (
    <div className="waiting-room">
      <h2>等待玩家加入</h2>
      <div className="room-code">{roomInfo.roomCode}</div>
      <p style={{ color: '#888' }}>将房间号分享给朋友</p>

      <ul className="players-list">
        {players.map((p, i) => (
          <li key={p.playerId || i}>
            {p.nickname || `玩家 ${i + 1}`}
            {p.isHost && ' 👑'}
            {!p.connected && ' (离线)'}
          </li>
        ))}
      </ul>

      {config && (
        <div className="room-config">
          <h3 style={{ color: '#888', marginBottom: 6 }}>游戏设置</h3>
          <div className="config-grid">
            <span className="config-label">起始手牌</span>
            <span className="config-value">{config.initialHandSize} 张</span>
            <span className="config-label">揭示轮数</span>
            <span className="config-value">{config.revealRounds} 轮</span>
            <span className="config-label">选牌轮数</span>
            <span className="config-value">{config.pickRounds} 轮</span>
            <span className="config-label">计分模式</span>
            <span className="config-value">{SCORING_LABELS[config.scoringMode] || '特殊计分规则'}</span>
          </div>
        </div>
      )}

      {isHost && (
        <button onClick={onStartGame} disabled={!canStart}>
          {canStart ? '开始游戏' : `等待更多玩家 (${players.length}/2)`}
        </button>
      )}
      {!isHost && <p style={{ color: '#888', marginTop: 16 }}>等待房主开始游戏...</p>}
    </div>
  );
}
