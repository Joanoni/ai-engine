import { useState, FormEvent } from 'react'
import { Reference } from '../types'
import { createReference } from '../api'
import './AddReferenceForm.css'

interface Props {
  boardId: string
  onAdded: (ref: Reference) => void
}

type RefType = Reference['type']

const PLACEHOLDERS: Record<RefType, string> = {
  image: 'Image URL (https://…)',
  color: 'Hex color (#7c3aed)',
  note: 'Note text…',
}

export default function AddReferenceForm({ boardId, onAdded }: Props) {
  const [type, setType] = useState<RefType>('image')
  const [content, setContent] = useState('')
  const [label, setLabel] = useState('')
  const [loading, setLoading] = useState(false)

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    if (!content.trim()) return
    setLoading(true)
    try {
      const ref = await createReference(boardId, type, content.trim(), label.trim())
      onAdded(ref)
      setContent('')
      setLabel('')
    } catch (err) {
      console.error(err)
    } finally {
      setLoading(false)
    }
  }

  return (
    <form className="add-ref-form" onSubmit={handleSubmit}>
      <select
        className="arf-select"
        value={type}
        onChange={e => setType(e.target.value as RefType)}
      >
        <option value="image">Image</option>
        <option value="color">Color</option>
        <option value="note">Note</option>
      </select>
      <input
        className="arf-input"
        placeholder={PLACEHOLDERS[type]}
        value={content}
        onChange={e => setContent(e.target.value)}
        required
      />
      <input
        className="arf-input"
        placeholder="Label (optional)"
        value={label}
        onChange={e => setLabel(e.target.value)}
      />
      <button type="submit" className="arf-btn" disabled={loading}>
        {loading ? '…' : 'Add'}
      </button>
    </form>
  )
}
