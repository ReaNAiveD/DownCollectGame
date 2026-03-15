import React, { useState } from 'react';
import { CardDisplay, InsertMarker, cardLabel } from './Card';

const PHASE_LABELS = {
  Reveal: '揭示阶段', Pick: '选牌阶段', Scoring: '计分中...',
  CharacterSelect: '角色选择', Setup: '准备中...',
};

// ─── Visual Event Panel: renders cards + slots, not just text ───
function EventPanel({ events, players, board }) {
  if (!events || events.length === 0) return null;
  const getName = (seat) => players?.find(p => p.seatIndex === seat)?.nickname || `P${seat + 1}`;

  return (
    <div className="event-panel">
      {events.map((evt, i) => (
        <div key={i} className="event-block">
          <div className="event-header">
            <span className="event-player">{getName(evt.seatIndex)}</span>
            <span className="event-desc">{evt.description}</span>
          </div>

          {/* Show revealed card for cardRevealed */}
          {evt.type === 'cardRevealed' && evt.card && (
            <div className="event-cards-row">
              <span className="event-label">揭示:</span>
              <CardDisplay card={evt.card} faceUp={true} small />
            </div>
          )}

          {/* Show flipped card for effect 9 */}
          {evt.type === 'effectFlip' && evt.card && (
            <div className="event-cards-row">
              <span className="event-label">翻明:</span>
              <div className="event-slot-group">
                {evt.slots?.map((s, j) => (
                  <span key={j} className="slot-tag">第{s.row + 1}行 位置{s.position + 1}</span>
                ))}
              </div>
              <CardDisplay card={evt.card} faceUp={true} small />
            </div>
          )}

          {/* Show affected positions for swap (3) */}
          {evt.type === 'effectSwap' && (
            <div className="event-cards-row">
              <span className="event-label">交换位置:</span>
              <div className="event-slot-group">
                {evt.slots?.map((s, j) => (
                  <span key={j} className="slot-tag">第{s.row + 1}行#{s.position + 1}</span>
                ))}
              </div>
            </div>
          )}

          {/* Show replaced position for 5/6 */}
          {evt.type === 'effectReplace' && (
            <div className="event-cards-row">
              <span className="event-label">替换位置:</span>
              {evt.slots?.map((s, j) => (
                <span key={j} className="slot-tag">第{s.row + 1}行 位置{s.position + 1}</span>
              ))}
            </div>
          )}

          {/* Show row for J flip */}
          {evt.type === 'effectFlipRow' && evt.row != null && (
            <div className="event-cards-row">
              <span className="event-label">翻转行:</span>
              <span className="slot-tag">第{evt.row + 1}行 全部明暗翻转</span>
            </div>
          )}

          {/* Show ALL revealed cards for Q effect */}
          {evt.type === 'effectQueenShuffle' && evt.row != null && (
            <div className="event-visual">
              <div className="event-cards-row">
                <span className="event-label">第{evt.row + 1}行展示:</span>
              </div>
              <div className="event-cards-display">
                {evt.cards?.map((c, j) => (
                  <div key={j} className="event-card-slot">
                    <CardDisplay card={c} faceUp={true} small />
                    {evt.slots?.[j] && (
                      <span className="slot-pos">#{evt.slots[j].position + 1}</span>
                    )}
                  </div>
                ))}
              </div>
              <div className="event-note">基础分值≤6的卡牌已洗混暗置放回</div>
            </div>
          )}

          {/* Show ALL revealed cards for K mix */}
          {evt.type === 'effectKingMix' && (
            <div className="event-visual">
              <div className="event-cards-row">
                <span className="event-label">选中并展示的卡牌 (含K):</span>
              </div>
              <div className="event-cards-display">
                {evt.cards?.map((c, j) => (
                  <div key={j} className="event-card-slot">
                    <CardDisplay card={c} faceUp={true} small />
                    {evt.slots?.[j] && (
                      <span className="slot-pos">第{evt.slots[j].row + 1}行#{evt.slots[j].position + 1}</span>
                    )}
                  </div>
                ))}
              </div>
              <div className="event-note">已洗混暗置放回原位</div>
            </div>
          )}

          {/* Swap with hand */}
          {evt.type === 'swap' && (
            <div className="event-cards-row">
              <span className="event-label">交换位置:</span>
              {evt.slots?.map((s, j) => (
                <span key={j} className="slot-tag">第{s.row + 1}行 位置{s.position + 1}</span>
              ))}
              <span className="event-note">手牌↔展示区</span>
            </div>
          )}

          {/* Insert position for 7/8 */}
          {evt.type === 'effectInsert' && evt.row != null && evt.position != null && (
            <div className="event-cards-row">
              <span className="event-label">插入位置:</span>
              <span className="slot-tag">第{evt.row + 1}行 位置{evt.position + 1}</span>
              <span className="event-note">后方卡牌右移</span>
            </div>
          )}

          {/* Small Joker: all face-up flipped down */}
          {evt.type === 'effectSmallJoker' && (
            <div className="event-cards-row">
              <span className="event-label">影响:</span>
              <span className="slot-tag">{evt.slots?.length || 0} 张明置卡牌被翻为暗置</span>
            </div>
          )}

          {/* Big Joker: all face-down shuffled */}
          {evt.type === 'effectBigJoker' && (
            <div className="event-cards-row">
              <span className="event-label">影响:</span>
              <span className="slot-tag">{evt.slots?.length || 0} 张暗置卡牌位置被洗混</span>
            </div>
          )}

          {/* Ace: card to deck, new card placed */}
          {evt.type === 'effectAce' && (
            <div className="event-cards-row">
              <span className="event-label">结果:</span>
              {evt.slots?.map((s, j) => (
                <span key={j} className="slot-tag">新卡牌暗置于 第{s.row + 1}行 位置{s.position + 1}</span>
              ))}
              {(!evt.slots || evt.slots.length === 0) && <span className="event-note">牌堆为空</span>}
            </div>
          )}

          {/* 2: shuffle current row */}
          {evt.type === 'effectShuffle' && evt.row != null && (
            <div className="event-cards-row">
              <span className="event-label">影响:</span>
              <span className="slot-tag">第{evt.row + 1}行 {evt.slots?.length || 0} 张暗置卡牌位置被洗混</span>
            </div>
          )}

          {/* 4: enters hand, recursive reveal */}
          {evt.type === 'effectFour' && (
            <div className="event-cards-row">
              <span className="event-note">将继续揭示下一张卡牌...</span>
            </div>
          )}

          {/* 10: peek */}
          {evt.type === 'effectPeek' && (
            <div className="event-cards-row">
              <span className="event-label">偷看位置:</span>
              {evt.slots?.map((s, j) => (
                <span key={j} className="slot-tag">第{s.row + 1}行 位置{s.position + 1}</span>
              ))}
              <span className="event-note">(仅操作者可见)</span>
            </div>
          )}

          {/* Effect skipped */}
          {evt.type === 'effectSkipped' && (
            <div className="event-cards-row">
              <span className="event-note" style={{ color: '#888' }}>效果被跳过</span>
            </div>
          )}

          {/* Swap skipped */}
          {evt.type === 'swapSkipped' && (
            <div className="event-cards-row">
              <span className="event-note">不交换</span>
            </div>
          )}
        </div>
      ))}
    </div>
  );
}

// ─── Other Player Rows (excludes self, ordered from mySeat+1 wrapping) ───
function PlayerRows({ players, mySeat, activeSeat }) {
  if (!players) return null;
  // Order: mySeat+1, mySeat+2, ..., mySeat-1 (skip self)
  const n = players.length;
  const ordered = [];
  for (let i = 1; i < n; i++) {
    const seat = (mySeat + i) % n;
    const p = players.find(pl => pl.seatIndex === seat);
    if (p) ordered.push(p);
  }
  return (
    <div className="player-rows">
      {ordered.map(p => (
        <div key={p.seatIndex} className={`player-row ${p.seatIndex === activeSeat ? 'active' : ''}`}>
          <div className="player-info-cell">
            <span className="player-name">{p.nickname || `P${p.seatIndex + 1}`}</span>
            {p.seatIndex === activeSeat && <span className="player-turn-indicator">◆</span>}
          </div>
          <div className="player-hand-indicator">
            {Array.from({ length: p.handSize }).map((_, i) => (
              <div key={i} className="mini-card">?</div>
            ))}
            {p.handSize === 0 && <span className="no-cards">无牌</span>}
          </div>
        </div>
      ))}
    </div>
  );
}

// ─── Compute highlighted slots from events ───
function getHighlightedSlots(events) {
  const set = new Set();
  if (!events) return set;
  for (const evt of events) {
    if (evt.slots) {
      for (const s of evt.slots) {
        set.add(`${s.row},${s.position}`);
      }
    }
  }
  return set;
}

import CardSidebar from './CardSidebar';

// ─── Main GameBoard ───
export default function GameBoard({ gameState, sendAction }) {
  const [selectedSlots, setSelectedSlots] = useState([]);
  const [selectedHandCard, setSelectedHandCard] = useState(-1);
  const [hoveredCard, setHoveredCard] = useState(null);

  const gs = gameState;
  if (!gs) return null;

  const pending = gs.pendingAction;
  const isMyTurn = pending != null;
  const actionType = pending?.type;
  const selectCount = pending?.selectCount || 1;

  const highlightedSlots = getHighlightedSlots(gs.events);

  const handleSlotClick = (row, pos) => {
    if (!isMyTurn) return;
    const coord = { row, position: pos };
    if (actionType === 'selectSlots' || actionType === 'selectSlotsPerRow') {
      setSelectedSlots(prev => {
        const exists = prev.find(s => s.row === row && s.position === pos);
        if (exists) return prev.filter(s => !(s.row === row && s.position === pos));
        if (actionType === 'selectSlotsPerRow') return [...prev.filter(s => s.row !== row), coord];
        if (prev.length >= selectCount) return prev;
        return [...prev, coord];
      });
    } else if (actionType === 'pickCard') {
      sendAction({ selectedSlots: [coord] });
      setSelectedSlots([]);
    }
  };

  const handleInsertPos = (pos) => { if (actionType === 'selectPosition') sendAction({ selectedPos: pos }); };
  const handleRowClick = (ri) => { if (actionType === 'selectRow') sendAction({ selectedRow: ri }); };
  const confirmSelection = () => { sendAction({ selectedSlots }); setSelectedSlots([]); };
  const handleConfirm = () => sendAction({});
  const handleReveal = () => sendAction({});
  const handleSwap = (doSwap) => {
    sendAction({ doSwap, handCardIndex: selectedHandCard >= 0 ? selectedHandCard : 0 });
    setSelectedHandCard(-1);
  };

  const isSlotSelectable = (row, pos) => {
    if (!isMyTurn) return false;
    if (actionType === 'pickCard' || actionType === 'selectSlots')
      return pending.validSlots?.some(s => s.row === row && s.position === pos);
    if (actionType === 'selectSlotsPerRow')
      return pending.validRows?.includes(row);
    return false;
  };
  const isSlotSelected = (row, pos) => selectedSlots.some(s => s.row === row && s.position === pos);
  const isSlotHighlighted = (row, pos) => highlightedSlots.has(`${row},${pos}`);
  const isRowClickable = actionType === 'selectRow';
  const showInsertMarkers = actionType === 'selectPosition';
  const insertRow = gs.round;

  const canConfirm = (() => {
    if (actionType === 'selectSlots') return selectedSlots.length === selectCount;
    if (actionType === 'selectSlotsPerRow') return selectedSlots.length === (pending?.validRows?.length || 0);
    return false;
  })();

  const boardRows = gs.board || [];
  const displayRowCount = showInsertMarkers ? Math.max(boardRows.length, insertRow + 1) : boardRows.length;

  return (
    <div className="game-layout">
    <div className="game">
      {/* Header */}
      <div className="game-header">
        <div className="phase-info">
          {PHASE_LABELS[gs.phase] || gs.phase} · 第 {gs.round + 1} 轮 · 回合 {gs.turn + 1}
        </div>
        <div className="deck-info">牌堆: {gs.deckRemaining} 张</div>
      </div>

      {/* Players */}
      <PlayerRows players={gs.players} mySeat={gs.mySeat} activeSeat={gs.activeSeat} />

      {/* Visual Event Panel */}
      <EventPanel events={gs.events} players={gs.players} board={gs.board} />

      {/* Revealed card overlay */}
      {gs.revealedCard && (
        <div className="revealed-overlay">
          <div className="revealed-inner">
            {actionType === 'swapDecision' ? (
              <>
                <p className="revealed-label">你从牌堆顶抽到了这张卡牌（仅你可见）</p>
                <CardDisplay card={gs.revealedCard} faceUp={true} />
                <p style={{ color: '#aaa', marginTop: 10, fontSize: '0.9rem' }}>
                  你可以选择一张手牌与抽到的牌互换：抽到的牌收入手牌，选中的手牌作为本回合需要处理的卡牌。也可以不交换，直接处理抽到的牌。
                </p>
                <div className="swap-hand-row">
                  {gs.myHand?.map((card, i) => (
                    <CardDisplay key={i} card={card} faceUp={true} small
                      selectable selected={selectedHandCard === i}
                      onClick={() => setSelectedHandCard(i === selectedHandCard ? -1 : i)} />
                  ))}
                </div>
                <div style={{ display: 'flex', gap: 8, marginTop: 12, justifyContent: 'center' }}>
                  <button className="btn-action" onClick={() => handleSwap(true)} disabled={selectedHandCard < 0}>
                    换入手牌{selectedHandCard >= 0 ? `（第 ${selectedHandCard + 1} 张）` : ''}
                  </button>
                  <button className="btn-secondary" onClick={() => handleSwap(false)}>不交换，直接处理此牌</button>
                </div>
              </>
            ) : (
              <>
                <p className="revealed-label">本回合需要处理的卡牌</p>
                <CardDisplay card={gs.revealedCard} faceUp={true} />
                {actionType === 'confirmReveal' && (
                  <button className="btn-confirm" onClick={handleConfirm}>所有人已看到，继续处理效果</button>
                )}
              </>
            )}
          </div>
        </div>
      )}

      {/* Board */}
      <div className="board-area">
        <h3 className="section-title">展示区</h3>
        {Array.from({ length: displayRowCount }).map((_, ri) => {
          const row = boardRows[ri] || [];
          const isInsertRow = showInsertMarkers && ri === insertRow;
          const rowClickable = isRowClickable && pending?.validRows?.includes(ri);
          return (
            <div key={ri} className={`board-row ${rowClickable ? 'row-clickable' : ''}`}
              onClick={rowClickable ? () => handleRowClick(ri) : undefined}>
              <div className="board-row-label">第 {ri + 1} 行</div>
              <div className="card-slots">
                {isInsertRow && <InsertMarker position={0} onClick={handleInsertPos} />}
                {row.map((slot, pos) => (
                  <React.Fragment key={pos}>
                    {slot.hasCard ? (
                      <div className={`card-wrapper ${isSlotHighlighted(ri, pos) ? 'highlighted' : ''}`}>
                        <CardDisplay card={slot.card} faceUp={slot.faceUp} peeked={slot.peekedCard}
                          selectable={isSlotSelectable(ri, pos)} selected={isSlotSelected(ri, pos)}
                          onClick={() => handleSlotClick(ri, pos)} onHover={setHoveredCard} />
                      </div>
                    ) : (
                      <div className="card face-down empty-slot"><span className="rank">·</span></div>
                    )}
                    {isInsertRow && <InsertMarker position={pos + 1} onClick={handleInsertPos} />}
                  </React.Fragment>
                ))}
                {isInsertRow && row.length === 0 && (
                  <span className="insert-hint">← 点击 ▼ 插入</span>
                )}
              </div>
            </div>
          );
        })}
        {displayRowCount === 0 && <p className="empty-text">展示区为空</p>}
      </div>

      {/* My Area */}
      <div className={`my-area ${gs.activeSeat === gs.mySeat ? 'my-turn' : ''}`} style={{ paddingBottom: isMyTurn ? 70 : 12 }}>
        <div className="my-header">
          <span className="my-name">
            {gs.players?.find(p => p.seatIndex === gs.mySeat)?.nickname || '你'}
            {gs.activeSeat === gs.mySeat && <span className="player-turn-indicator"> ◆ 你的回合</span>}
          </span>
          <span className="my-hand-count">{gs.myHand?.length || 0} 张手牌</span>
        </div>
        <div className="hand-cards">
          {gs.myHand?.map((card, i) => (
            <CardDisplay key={i} card={card} faceUp={true} onHover={setHoveredCard} />
          ))}
          {(!gs.myHand || gs.myHand.length === 0) && <p className="empty-text">无手牌</p>}
        </div>
      </div>

      {/* Action Panel */}
      {isMyTurn && actionType !== 'confirmReveal' && actionType !== 'swapDecision' && (
        <div className="action-panel">
          {actionType === 'revealCard' && (
            <><span className="action-label">轮到你从牌堆顶抽牌</span><button className="btn-action" onClick={handleReveal}>抽牌</button></>
          )}
          {(actionType === 'selectSlots' || actionType === 'selectSlotsPerRow') && (
            <>
              <span className="action-label">
                {pending?.context?.effect === 'three_swap' && '选择展示区中 2 张暗置卡牌，互换它们的位置'}
                {pending?.context?.effect === 'five_replace' && '选择展示区中 1 张暗置卡牌，将其移至牌堆底并用牌堆顶替补'}
                {pending?.context?.effect === 'six_replace' && '选择展示区中 1 张暗置卡牌，将其移至牌堆底并用牌堆顶替补'}
                {pending?.context?.effect === 'nine_flip' && '选择展示区中 1 张暗置卡牌，将其翻为明置'}
                {pending?.context?.effect === 'ten_peek' && '选择展示区中 1 张暗置卡牌进行偷看（仅你可见）'}
                {pending?.context?.effect === 'king_mix' && '从展示区每行选 1 张卡牌，将与 K 一起展示后洗混放回'}
                {!pending?.context?.effect && `选择 ${selectCount} 张`}
              </span>
              <span className="select-count">({selectedSlots.length}/{actionType === 'selectSlotsPerRow' ? pending?.validRows?.length || 0 : selectCount})</span>
              <button className="btn-action" onClick={confirmSelection} disabled={!canConfirm}>确认</button>
            </>
          )}
          {actionType === 'selectPosition' && <span className="action-label">点击 ▼ 标记选择要将卡牌插入当前行的哪个位置</span>}
          {actionType === 'selectRow' && (
            <span className="action-label">
              {pending?.context?.effect === 'jack_flip_row' && '点击展示区中的一行，将该行所有卡牌的明暗状态翻转'}
              {pending?.context?.effect === 'queen_reveal_shuffle' && '点击展示区中的一行，先全部展示，再将低分牌洗混暗置放回'}
              {!pending?.context?.effect && '选择一行'}
            </span>
          )}
          {actionType === 'pickCard' && <span className="action-label">从展示区选择一张卡牌加入你的手牌（放在最右侧）</span>}
        </div>
      )}
    </div>
    {/* Right Sidebar */}
    <CardSidebar revealedCard={gs.lastRevealedCard || gs.revealedCard} hoveredCard={hoveredCard} scoringMode={gs.scoringMode || 0} />
    </div>
  );
}
