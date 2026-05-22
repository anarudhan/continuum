import { Users, Clock } from 'lucide-react'

function Agents() {
  const agents = [
    { id: '1', name: 'Hermes', status: 'online', lastSeen: '2 min ago', memories: 456, sessions: 12 },
    { id: '2', name: 'Claude Code', status: 'online', lastSeen: '15 min ago', memories: 234, sessions: 8 },
    { id: '3', name: 'Codex', status: 'offline', lastSeen: '2 hours ago', memories: 189, sessions: 5 },
    { id: '4', name: 'OpenClaw', status: 'online', lastSeen: '5 min ago', memories: 123, sessions: 3 },
  ]

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold mb-2">Agents</h1>
        <p className="text-text-muted">Manage connected AI agents</p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {agents.map((agent) => (
          <div key={agent.id} className="bg-secondary rounded-xl p-6 border border-border">
            <div className="flex items-center justify-between mb-4">
              <div className="flex items-center gap-3">
                <Users className="text-accent" size={24} />
                <div>
                  <h3 className="font-semibold">{agent.name}</h3>
                  <div className="flex items-center gap-2 text-sm text-text-muted">
                    <div className={`w-2 h-2 rounded-full ${agent.status === 'online' ? 'bg-success' : 'bg-text-muted'}`} />
                    {agent.status}
                  </div>
                </div>
              </div>
            </div>

            <div className="grid grid-cols-3 gap-4 text-sm">
              <div>
                <div className="text-text-muted">Memories</div>
                <div className="font-semibold">{agent.memories}</div>
              </div>
              <div>
                <div className="text-text-muted">Sessions</div>
                <div className="font-semibold">{agent.sessions}</div>
              </div>
              <div>
                <div className="text-text-muted">Last Seen</div>
                <div className="font-semibold flex items-center gap-1">
                  <Clock size={12} />
                  {agent.lastSeen}
                </div>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

export default Agents
