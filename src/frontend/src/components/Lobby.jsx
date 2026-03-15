import React, { useState } from 'react';

export default function Lobby({ nickname, setNickname, onSetNickname, onCreateRoom, onJoinRoom, connected }) {
  const [roomCode, setRoomCode] = useState('');
  const [showConfig, setShowConfig] = useState(false);
  const [initialHandSize, setInitialHandSize] = useState(1);
  const [revealRounds, setRevealRounds] = useState(6);
  const [pickRounds, setPickRounds] = useState(4);
  const [scoringMode, setScoringMode] = useState(0);

  const handleCreate = () => {
    if (nickname.trim()) onSetNickname(nickname.trim());
    onCreateRoom({
      initialHandSize,
      revealRounds,
      pickRounds,
      scoringMode,
    });
  };

  const handleJoin = () => {
    if (!roomCode.trim()) return;
    if (nickname.trim()) onSetNickname(nickname.trim());
    onJoinRoom(roomCode.trim());
  };

  return (
    <div className="lobby">
      <h1>DownCollect</h1>
      <p style={{ color: connected ? '#4ecdc4' : '#e94560' }}>
        {connected ? '已连接' : '连接中...'}
      </p>
      <input
        placeholder="输入昵称"
        value={nickname}
        onChange={(e) => setNickname(e.target.value)}
        maxLength={16}
      />

      <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
        <button onClick={handleCreate} disabled={!connected}>创建房间</button>
        <button className="btn-config-toggle" onClick={() => setShowConfig(!showConfig)}>
          {showConfig ? '收起设置' : '⚙ 设置'}
        </button>
      </div>

      {showConfig && (
        <div className="create-config">
          <div className="config-row">
            <label>起始手牌</label>
            <select value={initialHandSize} onChange={e => setInitialHandSize(Number(e.target.value))}>
              {[0, 1, 2, 3, 4, 5].map(n => <option key={n} value={n}>{n} 张</option>)}
            </select>
          </div>
          <div className="config-row">
            <label>揭示轮数</label>
            <select value={revealRounds} onChange={e => setRevealRounds(Number(e.target.value))}>
              {[1, 2, 3, 4, 5, 6, 7, 8, 9, 10].map(n => <option key={n} value={n}>{n} 轮</option>)}
            </select>
          </div>
          <div className="config-row">
            <label>选牌轮数</label>
            <select value={pickRounds} onChange={e => setPickRounds(Number(e.target.value))}>
              {[1, 2, 3, 4, 5, 6, 7, 8].map(n => <option key={n} value={n}>{n} 轮</option>)}
            </select>
          </div>
          <div className="config-row">
            <label>计分模式</label>
            <select value={scoringMode} onChange={e => setScoringMode(Number(e.target.value))}>
              <option value={0}>特殊计分规则</option>
              <option value={1}>仅基础分值</option>
            </select>
          </div>
        </div>
      )}

      <div className="join-section">
        <input
          placeholder="房间号"
          value={roomCode}
          onChange={(e) => setRoomCode(e.target.value.toUpperCase())}
          maxLength={6}
          style={{ width: '120px' }}
        />
        <button onClick={handleJoin} disabled={!connected || !roomCode.trim()}>
          加入房间
        </button>
      </div>
    </div>
  );
}
