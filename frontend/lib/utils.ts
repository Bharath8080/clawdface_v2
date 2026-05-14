// Shared utility to generate unique emails for agents
export function generateAgentEmail(name: string): string {
  const cleanName = name.toLowerCase().replace(/[^a-z0-9]/g, '');
  
  // Format: name@agent.clawdface.ai
  return `${cleanName}@agent.clawdface.ai`;
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
