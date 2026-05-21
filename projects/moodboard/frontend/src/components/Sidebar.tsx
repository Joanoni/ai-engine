import { useState } from 'react'
import { Board } from '../types'
import { deleteBoard } from '../api'
import NewBoardForm from './NewBoardForm'
import './Sidebar.css'

interface Props {
  boards: Board[]
  selectedBoardId: string | null
  onSelectBoard: (id: string) => void
  onBoardCreated: (board: Board) => void
  onBoardDeleted: (id: string) => void
}

export default function Sidebar({ boards, selectedBoardId, onSelectBoard, onBoardCreated, onBoardDeleted }: Props) {
  const [showNewForm, setShowNewForm] = useState(false)

  async function handleDelete(e: React.MouseEvent, id: string) {
    e.stopPropagation()
    await deleteBoard(id)
    onBoardDeleted(id)
  }

  return (
    <aside className="sidebar">
      <div className="sidebar-header">
        <h1 className="sidebar-title">Moodboard</h1>
      </div>
      <nav className="sidebar-boards">
        {boards.map(board => (
          <div
            key={board.id}
            className={`board-item ${selectedBoardId === board.id ? 'board-item--active' : ''}`}
            onClick={() => onSelectBoard(board.id)}
          >
            <span className="board-item-name">{board.name}</span>
            <button
              className="board-item-delete"
              onClick={(e) => handleDelete(e, board.id)}
              title="Delete board"
            >×</button>
          </div>
        ))}
        {boards.length === 0 && (
          <p className="sidebar-empty">No boards yet.</p>
        )}
      </nav>
      <div className="sidebar-footer">
        {showNewForm ? (
          <NewBoardForm
            onCreated={(board) => { onBoardCreated(board); setShowNewForm(false) }}
            onCancel={() => setShowNewForm(false)}
          />
        ) : (
          <button className="btn-new-board" onClick={() => setShowNewForm(true)}>
            + New Board
          </button>
        )}
      </div>
    </aside>
  )
}
