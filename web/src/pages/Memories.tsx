import { Search, Plus } from 'lucide-react'
import { useState } from 'react'

function Memories() {
  const [searchQuery, setSearchQuery] = useState('')
  const [filterType, setFilterType] = useState('all')

  const memories = [
    { id: '1', type: 'semantic', content: 'OAuth2 + PKCE for auth flow', agent: 'Hermes', project: 'anarudhan', visibility: 'shared', createdAt: '2026-05-22T10:30:00Z' },
    { id: '2', type: 'episodic', content: 'Decided to use httpOnly cookies instead of localStorage', agent: 'Claude Code', project: 'anarudhan', visibility: 'shared', createdAt: '2026-05-22T10:15:00Z' },
    { id: '3', type: 'procedural', content: 'Deploy workflow: 1) npm run build 2) npm audit 3) security scan 4) git commit 5) open PR', agent: 'Codex', project: 'anarudhan', visibility: 'shared', createdAt: '2026-05-22T09:00:00Z' },
    { id: '4', type: 'semantic', content: 'Tech stack: Go 1.24 + Gin + PostgreSQL 16 + React 19', agent: 'Hermes', project: 'continuum', visibility: 'shared', createdAt: '2026-05-21T14:20:00Z' },
  ]

  const filteredMemories = memories.filter(m => {
    const matchesSearch = m.content.toLowerCase().includes(searchQuery.toLowerCase())
    const matchesType = filterType === 'all' || m.type === filterType
    return matchesSearch && matchesType
  })

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold mb-2">Memories</h1>
          <p className="text-text-muted">Browse and search agent memories</p>
        </div>
        <button className="flex items-center gap-2 px-4 py-2 bg-accent hover:bg-accent-hover rounded-lg transition-colors">
          <Plus size={18} />
          <span>New Memory</span>
        </button>
      </div>

      {/* Search and Filter */}
      <div className="flex gap-4">
        <div className="flex-1 relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-text-muted" size={18} />
          <input
            type="text"
            placeholder="Search memories..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full pl-10 pr-4 py-2 bg-secondary border border-border rounded-lg focus:outline-none focus:border-accent"
          />
        </div>
        <select
          value={filterType}
          onChange={(e) => setFilterType(e.target.value)}
          className="px-4 py-2 bg-secondary border border-border rounded-lg focus:outline-none focus:border-accent"
        >
          <option value="all">All Types</option>
          <option value="episodic">Episodic</option>
          <option value="semantic">Semantic</option>
          <option value="procedural">Procedural</option>
        </select>
      </div>

      {/* Memories List */}
      <div className="bg-secondary rounded-xl border border-border">
        <div className="divide-y divide-border">
          {filteredMemories.map((memory) => (
            <div key={memory.id} className="p-4 hover:bg-primary/50 transition-colors">
              <div className="flex items-start gap-4">
                <div className={`px-2 py-1 rounded text-xs font-medium capitalize ${
                  memory.type === 'semantic' ? 'bg-accent/10 text-accent' :
                  memory.type === 'episodic' ? 'bg-warning/10 text-warning' :
                  'bg-success/10 text-success'
                }`}>
                  {memory.type}
                </div>
                <div className="flex-1">
                  <p className="text-sm mb-2">{memory.content}</p>
                  <div className="flex items-center gap-4 text-xs text-text-muted">
                    <span>{memory.agent}</span>
                    <span>{memory.project}</span>
                    <span className={`px-2 py-0.5 rounded ${
                      memory.visibility === 'shared' ? 'bg-success/10 text-success' : 'bg-warning/10 text-warning'
                    }`}>
                      {memory.visibility}
                    </span>
                  </div>
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}

export default Memories
