import { useState, useEffect } from 'react'
import { Board, Reference } from '../types'
import { getReferences } from '../api'
import ReferenceCard from './ReferenceCard'
import AddReferenceForm from './AddReferenceForm'
import './BoardView.css'

interface Props {
  boardId: string
  boards: Board[]
}

export default function BoardView({ boardId, boards }: Props) {
  const [references, setReferences] = useState<Reference[]>([])
  const board = boards.find(b => b.id === boardId)

  useEffect(() => {
    setReferences([])
    getReferences(boardId).then(setReferences).catch(console.error)
  }, [boardId])

  function handleAdded(ref: Reference) {
    setReferences(prev => [...prev, ref])
  }

  function handleDeleted(id: string) {
    setReferences(prev => prev.filter(r => r.id !== id))
  }

  return (
    <div className="board-view">
      {board && (
        <div className="board-header">
          <h2 className="board-name">{board.name}</h2>
          {board.description && (
            <p className="board-description">{board.description}</p>
          )}
        </div>
      )}
      {references.length === 0 ? (
        <div className="refs-empty">No references yet — add one below.</div>
      ) : (
        <div className="refs-grid">
          {references.map(ref => (
            <ReferenceCard key={ref.id} reference={ref} onDeleted={handleDeleted} />
          ))}
        </div>
      )}
      <div className="board-form-bar">
        <AddReferenceForm boardId={boardId} onAdded={handleAdded} />
      </div>
    </div>
  )
}
