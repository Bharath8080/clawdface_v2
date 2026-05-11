// Shared utility to generate unique timestamped emails for agents
export function generateAgentEmail(name: string): string {
  const cleanName = name.toLowerCase().replace(/[^a-z0-9]/g, '');
  const now = new Date();
  const timestamp = now.toISOString()
    .slice(0, 16) // YYYY-MM-DDTHH:mm
    .replace(/:/g, '')
    .replace('T', '-');
  
  // Format: name-timestampclawdfaceai@agent.truhire.ai
  return `${cleanName}-${timestamp}clawdfaceai@agent.truhire.ai`;
}

export function formatDuration(seconds: number): string {
  if (isNaN(seconds) || seconds < 0) return '—';
  
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  const secs = Math.floor(seconds % 60);
  
  const parts: string[] = [];
  if (hours > 0) parts.push(`${hours}h`);
  if (minutes > 0) parts.push(`${minutes}m`);
  if (secs > 0 || parts.length === 0) parts.push(`${secs}s`);
  
  return parts.join(' ');
}
