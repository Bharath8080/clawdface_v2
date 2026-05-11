// Shared utility to generate unique emails for agents
export function generateAgentEmail(name: string): string {
  const cleanName = name.toLowerCase().replace(/[^a-z0-9]/g, '');
  
  // Format: name@agent.clawdface.ai
  return `${cleanName}@agent.clawdface.ai`;
}
