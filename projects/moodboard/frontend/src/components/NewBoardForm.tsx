import { useState, FormEvent } from 'react'
import { Board } from '../types'
import { createBoard } from '../api'
import './NewBoardForm.css'

interface Props {
  onCreated: (board: Board) => void
  onCancel: () => void
}

export default function NewBoardForm({ onCreated, onCancel }: Props) {
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [loading, setLoading] = useState(false)

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    if (!name.trim()) return
    setLoading(true)
    try {
      const board = await createBoard(name.trim(), description.trim())
      onCreated(board)
    } catch (err) {
      console.error(err)
    } finally {
      setLoading(false)
    }
  }

  return (
    <form className="new-board-form" onSubmit={handleSubmit}>
      <input
        className="nbf-input"
        placeholder="Board name *"
        value={name}
        onChange={e => setName(e.target.value)}
        required
        autoFocus
      />
      <input
        className="nbf-input"
        placeholder="Description"
        value={description}
        onChange={e => setDescription(e.target.value)}
      />
      <div className="nbf-actions">
        <button type="submit" className="nbf-btn nbf-btn--primary" disabled={loading}>
          {loading ? '…' : 'Create'}
        </button>
        <button type="button" className="nbf-btn nbf-btn--secondary" onClick={onCancel}>
          Cancel
        </button>
      </div>
    </form>
  )
}
