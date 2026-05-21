import { Reference } from '../types'
import { deleteReference } from '../api'
import './ReferenceCard.css'

interface Props {
  reference: Reference
  onDeleted: (id: string) => void
}

export default function ReferenceCard({ reference, onDeleted }: Props) {
  async function handleDelete() {
    await deleteReference(reference.id)
    onDeleted(reference.id)
  }

  return (
    <div className="ref-card">
      <button className="ref-card-delete" onClick={handleDelete} title="Delete">×</button>
      {reference.type === 'image' && (
        <div className="ref-card-image-wrap">
          <img
            className="ref-card-img"
            src={reference.content}
            alt={reference.label || 'image reference'}
          />
        </div>
      )}
      {reference.type === 'color' && (
        <div
          className="ref-card-swatch"
          style={{ backgroundColor: reference.content }}
        />
      )}
      {reference.type === 'note' && (
        <div className="ref-card-note">{reference.content}</div>
      )}
      <div className="ref-card-meta">
        {reference.type === 'color' && (
          <span className="ref-card-hex">{reference.content}</span>
        )}
        {reference.label && (
          <span className="ref-card-label">{reference.label}</span>
        )}
      </div>
    </div>
  )
}
