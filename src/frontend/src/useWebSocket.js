import { useRef, useState, useCallback, useEffect } from 'react';

const COpSetNickname = 1;
const COpCreateRoom = 2;
const COpJoinRoom = 3;
const COpStartGame = 4;
const COpPlayerAction = 5;

const SOpGameState = 101;
const SOpError = 107;
const SOpRoomUpdate = 108;
const SOpRoomCreated = 109;

function getDeviceId() {
  let id = localStorage.getItem('dc_device_id');
  if (!id) {
    id = crypto.randomUUID();
    localStorage.setItem('dc_device_id', id);
  }
  return id;
}

export function useWebSocket() {
  const wsRef = useRef(null);
  const [connected, setConnected] = useState(false);
  const [gameState, setGameState] = useState(null);
  const [roomInfo, setRoomInfo] = useState(null);
  const [error, setError] = useState(null);
  const [playerId, setPlayerId] = useState(null);

  const connect = useCallback(() => {
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) return;

    const deviceId = getDeviceId();
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.host;
    const ws = new WebSocket(`${protocol}//${host}/ws?deviceId=${deviceId}`);

    ws.onopen = () => {
      setConnected(true);
      setError(null);
    };

    ws.onclose = () => {
      setConnected(false);
      setTimeout(() => connect(), 2000);
    };

    ws.onerror = () => {
      setError('Connection failed');
    };

    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data);
        switch (msg.op) {
          case SOpGameState:
            setGameState(msg.data);
            break;
          case SOpRoomUpdate:
            setRoomInfo(msg.data);
            break;
          case SOpRoomCreated:
            setPlayerId(msg.data.playerId);
            setRoomInfo({ roomCode: msg.data.roomCode, phase: 'WaitingPlayers', players: [] });
            break;
          case SOpError:
            setError(msg.data.error);
            setTimeout(() => setError(null), 3000);
            break;
          default:
            break;
        }
      } catch (e) {
        console.error('Failed to parse message', e);
      }
    };

    wsRef.current = ws;
  }, []);

  useEffect(() => {
    connect();
    return () => {
      if (wsRef.current) wsRef.current.close();
    };
  }, [connect]);

  const send = useCallback((op, data) => {
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({ op, data }));
    }
  }, []);

  const setNickname = useCallback((nickname) => send(COpSetNickname, { nickname }), [send]);
  const createRoom = useCallback((config) => send(COpCreateRoom, { config }), [send]);
  const joinRoom = useCallback((roomCode) => send(COpJoinRoom, { roomCode: roomCode.toUpperCase() }), [send]);
  const startGame = useCallback(() => send(COpStartGame, {}), [send]);
  const sendAction = useCallback((action) => send(COpPlayerAction, action), [send]);

  return {
    connected,
    gameState,
    roomInfo,
    error,
    playerId,
    setNickname,
    createRoom,
    joinRoom,
    startGame,
    sendAction,
  };
}
