import { Shield, Database, Bell } from 'lucide-react'

function SettingsPage() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold mb-2">Settings</h1>
        <p className="text-text-muted">Configure your Continuum instance</p>
      </div>

      <div className="space-y-4">
        <div className="bg-secondary rounded-xl p-6 border border-border">
          <div className="flex items-center gap-3 mb-4">
            <Shield className="text-accent" size={24} />
            <h2 className="text-lg font-semibold">Security</h2>
          </div>
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <div className="font-medium">API Key Rotation</div>
                <div className="text-sm text-text-muted">Automatically rotate API keys every 90 days</div>
              </div>
              <div className="w-12 h-6 bg-accent rounded-full relative cursor-pointer">
                <div className="absolute right-1 top-1 w-4 h-4 bg-white rounded-full" />
              </div>
            </div>
            <div className="flex items-center justify-between">
              <div>
                <div className="font-medium">Rate Limiting</div>
                <div className="text-sm text-text-muted">Limit requests to 1000 per minute per agent</div>
              </div>
              <div className="w-12 h-6 bg-accent rounded-full relative cursor-pointer">
                <div className="absolute right-1 top-1 w-4 h-4 bg-white rounded-full" />
              </div>
            </div>
          </div>
        </div>

        <div className="bg-secondary rounded-xl p-6 border border-border">
          <div className="flex items-center gap-3 mb-4">
            <Database className="text-accent" size={24} />
            <h2 className="text-lg font-semibold">Storage</h2>
          </div>
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <div className="font-medium">Memory Compression</div>
                <div className="text-sm text-text-muted">Auto-compress memories older than 7 days</div>
              </div>
              <div className="w-12 h-6 bg-accent rounded-full relative cursor-pointer">
                <div className="absolute right-1 top-1 w-4 h-4 bg-white rounded-full" />
              </div>
            </div>
            <div className="flex items-center justify-between">
              <div>
                <div className="font-medium">Encryption at Rest</div>
                <div className="text-sm text-text-muted">Encrypt all stored memories</div>
              </div>
              <div className="w-12 h-6 bg-accent rounded-full relative cursor-pointer">
                <div className="absolute right-1 top-1 w-4 h-4 bg-white rounded-full" />
              </div>
            </div>
          </div>
        </div>

        <div className="bg-secondary rounded-xl p-6 border border-border">
          <div className="flex items-center gap-3 mb-4">
            <Bell className="text-accent" size={24} />
            <h2 className="text-lg font-semibold">Notifications</h2>
          </div>
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <div className="font-medium">Budget Alerts</div>
                <div className="text-sm text-text-muted">Alert when spending reaches 80% of budget</div>
              </div>
              <div className="w-12 h-6 bg-accent rounded-full relative cursor-pointer">
                <div className="absolute right-1 top-1 w-4 h-4 bg-white rounded-full" />
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

export default SettingsPage
