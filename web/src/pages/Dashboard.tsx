import { Database, Users, Activity, TrendingUp, Clock } from 'lucide-react'

function Dashboard() {
  const stats = [
    { label: 'Total Memories', value: '1,234', icon: Database, color: 'text-accent' },
    { label: 'Active Agents', value: '5', icon: Users, color: 'text-success' },
    { label: 'Active Sessions', value: '3', icon: Activity, color: 'text-warning' },
    { label: 'Today\'s Cost', value: '$2.45', icon: TrendingUp, color: 'text-danger' },
  ]

  const recentMemories = [
    { type: 'semantic', content: 'OAuth2 + PKCE for auth flow', agent: 'Hermes', time: '2 min ago' },
    { type: 'episodic', content: 'Decided to use httpOnly cookies', agent: 'Claude Code', time: '15 min ago' },
    { type: 'procedural', content: 'Deploy: test → build → push', agent: 'Codex', time: '1 hour ago' },
    { type: 'semantic', content: 'Project stack: Go + React', agent: 'Hermes', time: '2 hours ago' },
  ]

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold mb-2">Dashboard</h1>
        <p className="text-text-muted">Overview of your agent memory mesh</p>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {stats.map((stat) => (
          <div key={stat.label} className="bg-secondary rounded-xl p-6 border border-border">
            <div className="flex items-center justify-between mb-4">
              <stat.icon className={stat.color} size={24} />
            </div>
            <div className="text-2xl font-bold">{stat.value}</div>
            <div className="text-sm text-text-muted">{stat.label}</div>
          </div>
        ))}
      </div>

      {/* Recent Memories */}
      <div className="bg-secondary rounded-xl border border-border">
        <div className="p-6 border-b border-border">
          <h2 className="text-lg font-semibold">Recent Memories</h2>
        </div>
        <div className="divide-y divide-border">
          {recentMemories.map((memory, i) => (
            <div key={i} className="p-4 hover:bg-primary/50 transition-colors">
              <div className="flex items-start gap-4">
                <div className={`px-2 py-1 rounded text-xs font-medium ${
                  memory.type === 'semantic' ? 'bg-accent/10 text-accent' :
                  memory.type === 'episodic' ? 'bg-warning/10 text-warning' :
                  'bg-success/10 text-success'
                }`}>
                  {memory.type}
                </div>
                <div className="flex-1">
                  <p className="text-sm">{memory.content}</p>
                  <div className="flex items-center gap-4 mt-2 text-xs text-text-muted">
                    <span>{memory.agent}</span>
                    <span className="flex items-center gap-1">
                      <Clock size={12} />
                      {memory.time}
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

export default Dashboard
