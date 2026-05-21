import { useState, useEffect } from 'react'
import { Board } from './types'
import { getBoards } from './api'
import Sidebar from './components/Sidebar'
import BoardView from './components/BoardView'

export default function App() {
  const [boards, setBoards] = useState<Board[]>([])
  const [selectedBoardId, setSelectedBoardId] = useState<string | null>(null)

  useEffect(() => {
    getBoards().then(setBoards).catch(console.error)
  }, [])

  function handleBoardCreated(board: Board) {
    setBoards(prev => [...prev, board])
    setSelectedBoardId(board.id)
  }

  function handleBoardDeleted(id: string) {
    setBoards(prev => prev.filter(b => b.id !== id))
    if (selectedBoardId === id) setSelectedBoardId(null)
  }

  return (
    <div className="app-layout">
      <Sidebar
        boards={boards}
        selectedBoardId={selectedBoardId}
        onSelectBoard={setSelectedBoardId}
        onBoardCreated={handleBoardCreated}
        onBoardDeleted={handleBoardDeleted}
      />
      <main className="app-main">
        {selectedBoardId ? (
          <BoardView boardId={selectedBoardId} boards={boards} />
        ) : (
          <div className="empty-state">
            <h2>Welcome to Moodboard</h2>
            <p>Select a board from the sidebar or create a new one to get started.</p>
          </div>
        )}
      </main>
    </div>
  )
}
