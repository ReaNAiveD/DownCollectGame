import React, { useState } from 'react';
import { useWebSocket } from './useWebSocket';
import Lobby from './components/Lobby';
import WaitingRoom from './components/WaitingRoom';
import GameBoard from './components/GameBoard';
import Results from './components/Results';

export default function App() {
  const ws = useWebSocket();
  const [nickname, setNickname] = useState('');

  // Determine which screen to show
  const gamePhase = ws.gameState?.phase;
  const inGame = gamePhase && gamePhase !== 'WaitingPlayers';
  const inRoom = ws.roomInfo != null;
  const isFinished = gamePhase === 'Finished';

  return (
    <div className="app">
      {ws.error && <div className="error-toast">{ws.error}</div>}

      {!inRoom && !inGame && (
        <Lobby
          nickname={nickname}
          setNickname={setNickname}
          onSetNickname={ws.setNickname}
          onCreateRoom={ws.createRoom}
          onJoinRoom={ws.joinRoom}
          connected={ws.connected}
        />
      )}

      {inRoom && !inGame && (
        <WaitingRoom
          roomInfo={ws.roomInfo}
          playerId={ws.playerId}
          onStartGame={ws.startGame}
        />
      )}

      {inGame && !isFinished && (
        <GameBoard
          gameState={ws.gameState}
          sendAction={ws.sendAction}
        />
      )}

      {isFinished && (
        <Results gameState={ws.gameState} />
      )}
    </div>
  );
}
