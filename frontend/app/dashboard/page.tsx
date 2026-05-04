"use client";

import { CloseIcon } from "@/components/CloseIcon";
import { NoAgentNotification } from "@/components/NoAgentNotification";
import TranscriptionView from "@/components/TranscriptionView";
import {
  BarVisualizer,
  DisconnectButton,
  RoomAudioRenderer,
  VideoTrack,
} from "@livekit/components-react";
// @ts-ignore - Internal context may not be exported in TS declaration
import { RoomContext, useVoiceAssistant, useRoomContext } from "@livekit/components-react";
import useCombinedTranscriptions from "@/hooks/useCombinedTranscriptions";
import { AnimatePresence, motion } from "framer-motion";
import { Room, RoomEvent, DisconnectReason } from "livekit-client";
import { useCallback, useEffect, useState, useRef, Suspense, createContext, useContext } from "react";
import type { ConnectionDetails } from "@/app/api/connection-details/route";
import { useRouter, useSearchParams } from "next/navigation";
import { useUser } from "@stackframe/stack";
import { initDefaultApiKey } from "@/app/services/apiKeyService";
import { createUserServiceServer } from "@/app/services/createUserService";
import { createAgent, updateAgent, getAgents, deleteAgent, type AgentBot } from "@/app/services/agentService";
import { getConversations, getConversationById } from "@/app/services/conversationService";
import { sendUsageData } from "@/app/services/usageService";
import { Sidebar } from "@/components/Sidebar";
import { SubscriptionView } from "@/components/SubscriptionView";
import Image from "next/image";
import {
  createBotAction as createBot,
  updateLastConfigAction as updateLastConfig,
  syncUserAction,
} from "@/lib/database-actions";
import { type AvatarItem, fetchAvatars } from "@/app/services/avatarService";

// Duplicate AvatarsContext removed


// ─── Session Config Defaults ────────────────────────────────────────────────
const DEFAULTS = {
  openclawUrl:     "",
  gatewayToken:    "",
  sessionKey:      "",
  avatarId:        "",
  botName:         "",
  thinkingEnabled: "true",
  thinkingDelay:   "5.0",
};

const stripSessionKey = (key: string) => {
  if (!key) return "";
  // Remove internal prefix agent:main:
  let clean = key.replace(/^agent:main:/, "");
  // Remove unique timestamp suffix (hyphen followed by 14 digits suffix like -20260314203015)
  clean = clean.replace(/-\d{14}$/, "");
  return clean;
};

const getBotAvatarId = (bot?: AgentBot | null) => {
  const avatar = bot?.avatars?.[0];
  return avatar?.avatar_key_id || avatar?.avatar_id || "";
};

const withPrimaryAvatar = (avatars: AgentBot["avatars"] | undefined, avatarId: string) => {
  if (!avatarId) return avatars ?? [];
  if (!avatars?.length) return [{ avatar_key_id: avatarId }];

  const [primary, ...rest] = avatars;
  const idField = primary.avatar_id && !primary.avatar_key_id ? "avatar_id" : "avatar_key_id";
  return [{ ...primary, [idField]: avatarId }, ...rest];
};

import { AVATARS } from "@/lib/constants";

const DASHBOARD_SESSIONS = new Set([
  "Library",
  "DirectCall",
  "AddBot",
  "My Bot",
  "Doctor",
  "Avatars",
  "Conversations",
]);


export const AvatarsContext = createContext<AvatarItem[]>(AVATARS);
export const useAvatars = () => useContext(AvatarsContext);

// ─── Icons ──────────────────────────────────────────────────────────────────
const UserIcon = ({ size = 15 }: { size?: number }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/>
  </svg>
);
const SmileIcon = ({ size = 15, className = "" }: { size?: number, className?: string }) => (
  <svg className={className} width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <circle cx="12" cy="12" r="10"/><path d="M8 14s1.5 2 4 2 4-2 4-2"/><line x1="9" y1="9" x2="9.01" y2="9"/><line x1="15" y1="9" x2="15.01" y2="9"/>
  </svg>
);
const LibraryIcon = ({ size = 15, className = "" }: { size?: number, className?: string }) => (
  <svg className={className} width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <rect width="16" height="20" x="4" y="2" rx="2" ry="2"/>
    <line x1="8" x2="16" y1="6" y2="6"/>
    <line x1="8" x2="16" y1="10" y2="10"/>
    <line x1="8" x2="16" y1="14" y2="14"/>
    <line x1="8" x2="16" y1="18" y2="18"/>
  </svg>
);
const LinkIcon = ({ size = 16 }: { size?: number }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/>
    <path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/>
  </svg>
);
const KeyIcon = ({ size = 16 }: { size?: number }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <path d="m21 2-2 2m-7.61 7.61a5.5 5.5 0 1 1-7.778 7.778 5.5 5.5 0 0 1 7.777-7.777zm0 0L15.5 7.5m0 0 3 3L22 7l-3-3m-3.5 3.5L19 4"/>
  </svg>
);
const HashIcon2 = ({ size = 16 }: { size?: number }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <line x1="4" x2="20" y1="9" y2="9"/><line x1="4" x2="20" y1="15" y2="15"/>
    <line x1="10" x2="8" y1="3" y2="21"/><line x1="16" x2="14" y1="3" y2="21"/>
  </svg>
);
const SettingsIcon = ({ size = 20, className = "" }: { size?: number, className?: string }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className={className}>
    <path d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.1a2 2 0 0 1-1-1.72v-.51a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z"/>
    <circle cx="12" cy="12" r="3"/>
  </svg>
);
const RefreshCwIcon = ({ size = 20, className = "" }: { size?: number, className?: string }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className={className}>
    <path d="M3 12a9 9 0 0 1 9-9 9.75 9.75 0 0 1 6.74 2.74L21 8"/>
    <path d="M21 3v5h-5"/>
    <path d="M21 12a9 9 0 0 1-9 9 9.75 9.75 0 0 1-6.74-2.74L3 16"/>
    <path d="M3 21v-5h5"/>
  </svg>
);
const TrashIcon = ({ size = 20 }: { size?: number }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <path d="M3 6h18"/><path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6"/><path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2"/><line x1="10" y1="11" x2="10" y2="17"/><line x1="14" y1="11" x2="14" y2="17"/>
  </svg>
);
const ClockIcon = ({ size = 20, className = "" }: { size?: number, className?: string }) => (
  <svg className={className} width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/>
  </svg>
);
const MicIcon = ({ size = 20 }: { size?: number }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <path d="M12 2a3 3 0 0 0-3 3v7a3 3 0 0 0 6 0V5a3 3 0 0 0-3-3Z"/><path d="M19 10v2a7 7 0 0 1-14 0v-2"/><line x1="12" x2="12" y1="19" y2="22"/>
  </svg>
);
const MicOffIcon = ({ size = 20 }: { size?: number }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <line x1="2" x2="22" y1="2" y2="22"/><path d="M18.89 13.23A7.12 7.12 0 0 0 19 12v-2"/><path d="M5 10v2a7 7 0 0 0 12 5"/><path d="M15 9.34V5a3 3 0 0 0-5.68-1.33"/><path d="M9 9v3a3 3 0 0 0 5.12 2.12"/><line x1="12" x2="12" y1="19" y2="22"/>
  </svg>
);
const MessageIcon = ({ size = 20 }: { size?: number }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <path d="M7.9 20A9 9 0 1 0 4 16.1L2 22Z"/>
  </svg>
);
const MailIcon = ({ size = 20 }: { size?: number }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <rect width="20" height="16" x="2" y="4" rx="2"/><path d="m22 7-8.97 5.7a1.94 1.94 0 0 1-2.06 0L2 7"/>
  </svg>
);
const CheckIcon = ({ size = 16, className = "" }: { size?: number, className?: string }) => (
  <svg className={className} width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round">
    <polyline points="20 6 9 17 4 12"/>
  </svg>
);
const ChevronDownIcon = ({ className = "", size = 14 }: { className?: string, size?: number }) => (
  <svg className={className} width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
    <path d="m6 9 6 6 6-6"/>
  </svg>
);
const CrossIcon = ({ size = 24 }: { size?: number }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <path d="M18 6 6 18"/><path d="m6 6 12 12"/>
  </svg>
);
const ActivityIcon = ({ size = 20, className = "" }: { size?: number, className?: string }) => (
  <svg className={className} width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <path d="M22 12h-4l-3 9L9 3 l-3 9H2"/>
  </svg>
);
const AlertCircleIcon = ({ size = 20, className = "" }: { size?: number, className?: string }) => (
  <svg className={className} width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/>
  </svg>
);
const CopyIcon = ({ size = 16, className = "" }: { size?: number, className?: string }) => (
  <svg className={className} width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <rect width="14" height="14" x="8" y="8" rx="2" ry="2"/><path d="M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2"/>
  </svg>
);
const ShieldIcon = ({ size = 20, className = "" }: { size?: number, className?: string }) => (
  <svg className={className} width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10Z"/>
  </svg>
);
const TerminalIcon = ({ size = 20, className = "" }: { size?: number, className?: string }) => (
  <svg className={className} width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <polyline points="4 17 10 11 4 5"/><line x1="12" y1="19" x2="20" y2="19"/>
  </svg>
);
const ShieldCheckIcon = ({ size = 20, className = "" }: { size?: number, className?: string }) => (
  <svg className={className} width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10Z"/><path d="m9 12 2 2 4-4"/>
  </svg>
);

const RecallUrlModal = ({
  isOpen,
  onClose,
  config,
}: {
  isOpen: boolean;
  onClose: () => void;
  config: any;
}) => {
  const [copied, setCopied] = useState(false);
  const [roomName, setRoomName] = useState("");
  const [recallSessionKey, setRecallSessionKey] = useState("");

  useEffect(() => {
    if (isOpen) {
      const generateId = () => {
        const now = new Date();
        const pad = (n: number) => n.toString().padStart(2, '0');
        const year = now.getFullYear();
        const month = pad(now.getMonth() + 1);
        const day = pad(now.getDate());
        const hours = pad(now.getHours());
        const minutes = pad(now.getMinutes());
        const seconds = pad(now.getSeconds());
        return `${year}-${month}-${day}T${hours}-${minutes}-${seconds}`;
      };
      
      const uniqueId = generateId();
      setRoomName(`room-${uniqueId}`);
      setRecallSessionKey(`session-${uniqueId}`);
      setCopied(false);
    }
  }, [isOpen]);

  const baseUrl = typeof window !== "undefined" ? window.location.origin : "";
  const recallUrl = `${baseUrl}/avatar?room=${roomName}&avatarId=${config.avatarId}&openclawUrl=${encodeURIComponent(config.openclawUrl)}&gatewayToken=${encodeURIComponent(config.gatewayToken)}&sessionKey=${encodeURIComponent(recallSessionKey)}&connection_type=recall`;

  const handleCopy = () => {
    navigator.clipboard.writeText(recallUrl);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <AnimatePresence>
      {isOpen && (
        <div className="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm">
          <motion.div
            initial={{ opacity: 0, scale: 0.95, y: 20 }}
            animate={{ opacity: 1, scale: 1, y: 0 }}
            exit={{ opacity: 0, scale: 0.95, y: 20 }}
            className="w-full max-w-lg bg-surface-secondary border border-white/10 rounded-2xl shadow-2xl overflow-hidden"
          >
            <div className="p-6 border-b border-white/5 flex items-center justify-between">
              <h3 className="text-xl font-bold text-white flex items-center gap-2">
                <LinkIcon size={20} />
                Recall.ai Integration
              </h3>
              <button onClick={onClose} className="text-neutral-500 hover:text-white transition-colors">
                <CloseIcon />
              </button>
            </div>
            <div className="p-6 space-y-6">
              <div className="space-y-2">
                <label className="text-xs font-bold text-neutral-500 uppercase tracking-widest">Room Name</label>
                <input
                  type="text"
                  value={roomName}
                  onChange={(e) => setRoomName(e.target.value)}
                  className="w-full bg-white/5 border border-white/10 rounded-xl px-4 py-3 text-sm font-mono text-white focus:outline-none focus:border-brand/50 transition-colors"
                />
              </div>
              <div className="space-y-2">
                <label className="text-xs font-bold text-neutral-500 uppercase tracking-widest">Public Video URL</label>
                <div className="relative group">
                  <textarea
                    readOnly
                    value={recallUrl}
                    className="w-full bg-black border border-white/10 rounded-xl px-4 py-3 text-xs font-mono text-neutral-400 h-32 resize-none break-all"
                  />
                  <button
                    onClick={handleCopy}
                    className="absolute bottom-3 right-3 px-4 py-2 bg-brand hover:bg-brand-hover text-black text-xs font-bold rounded-lg transition-all shadow-lg active:scale-95"
                  >
                    {copied ? "Copied!" : "Copy URL"}
                  </button>
                </div>
              </div>
              <p className="text-xs text-neutral-500 leading-relaxed italic">
                Use this URL when creating a Recall bot. The bot will join this room and display the avatar.
              </p>
            </div>
          </motion.div>
        </div>
      )}
    </AnimatePresence>
  );
};

// ─── Gateway Checklist Component ──────────────────────────────────────────────
const GatewayChecklist = () => (
  <div className="bg-black/40 rounded-2xl p-6 border border-white/5 space-y-4">
    <h4 className="text-xs font-bold text-white uppercase tracking-widest flex items-center gap-2 mb-2">
       Verification Checklist
    </h4>
    <ul className="text-[14px] text-neutral-400 space-y-3 list-none font-medium">
      <li className="flex items-center gap-3">
        <div className="w-6 h-6 rounded-lg bg-white/5 flex items-center justify-center text-[10px] text-neutral-500 shrink-0">1</div>
        Is your Ngrok tunnel running?
      </li>
      <li className="flex items-center gap-3">
        <div className="w-6 h-6 rounded-lg bg-white/5 flex items-center justify-center text-[10px] text-neutral-500 shrink-0">2</div>
        Is the OpenClaw Gateway service started?
      </li>
      <li className="flex items-center gap-3">
        <div className="w-6 h-6 rounded-lg bg-white/5 flex items-center justify-center text-[10px] text-neutral-500 shrink-0">3</div>
        Check your Target Gateway URL for typos.
      </li>
    </ul>
  </div>
);

// ─── Doctor View ─────────────────────────────────────────────────────────────
function DoctorView({ bots, onHealthUpdate }: { bots: AgentBot[], onHealthUpdate?: (key: string, status: 'healthy' | 'unhealthy') => void }) {
  const [url, setUrl] = useState("");
  const [status, setStatus] = useState<"idle" | "checking" | "healthy" | "error_404" | "error_connection">( "idle" );
  const [lastCheck, setLastCheck] = useState<Date | null>(null);
  const [apiError, setApiError] = useState<string | null>(null);

  useEffect(() => {
    if (bots.length > 0 && !url) {
      setUrl(bots[0].config?.openclaw_url ?? "");
    }
  }, [bots, url]);

  const runDiagnostics = async () => {
    if (!url) return;
    setStatus("checking");
    setApiError(null);

    const activeBot = bots.find(b => (b.config?.openclaw_url ?? "").replace(/\/$/, "") === url.replace(/\/$/, "")) || (bots.length > 0 ? bots[0] : null);
    const token = activeBot?.config?.gateway_token || "";
    const sessionKey = activeBot?.config?.session_key || "";
    
    try {
      // Use the server-side API route to avoid CORS
      const response = await fetch('/api/doctor/check', {
        method: "POST",
        headers: {
          "Content-Type": "application/json"
        },
        body: JSON.stringify({
          url,
          token,
          sessionKey
        })
      });

      const data = await response.json();
      setLastCheck(new Date());

      if (data.status === 200) {
        setStatus("healthy");
        // Always pass the ID and status to the handler
        if (onHealthUpdate && activeBot) onHealthUpdate(activeBot.id, 'healthy');
      } else if (data.status === 404) {
        // If it's a 404, check if it's an ngrok URL
        const isNgrok = url.includes("ngrok-free.dev") || 
                        url.includes("ngrok.io") || 
                        url.includes("ngrok-free.app") ||
                        (data.server && data.server.toLowerCase().includes("ngrok"));
        
        if (isNgrok) {
          setApiError("Ngrok returned a 404. Your tunnel is likely not running or the URL has expired.");
          setStatus("error_connection");
          if (onHealthUpdate && activeBot) onHealthUpdate(activeBot.id, 'unhealthy');
        } else {
          // If not ngrok, it might be the gateway but the endpoint is off
          setStatus("error_404");
          if (onHealthUpdate && activeBot) onHealthUpdate(activeBot.id, 'unhealthy');
        }
      } else if (data.error) {
        setApiError(data.error);
        setStatus("error_connection");
        if (onHealthUpdate && activeBot) onHealthUpdate(activeBot.id, 'unhealthy');
      } else {
        setStatus("error_connection");
        if (onHealthUpdate && activeBot) onHealthUpdate(activeBot.id, 'unhealthy');
      }
    } catch (error: any) {
      console.error("DIAGNOSTICS_ERROR:", error);
      setLastCheck(new Date());
      setApiError(error.message || "Unknown error");
      setStatus("error_connection");
    }
  };

  return (
    <div className="absolute inset-0 overflow-y-auto p-6 md:p-10 custom-scrollbar bg-canvas z-10">
      <div className="max-w-4xl mx-auto pb-20">
        <header className="mb-10">
          <h1 className="text-3xl font-bold text-white tracking-tight flex items-center gap-3">
            <ActivityIcon size={32} className="text-brand" />
            Gateway Doctor
          </h1>
          <p className="text-[#6b7280] mt-2 text-sm">Diagnose and fix connectivity between ClawdFace and your OpenClaw Gateway.</p>
        </header>

        <div className="space-y-6">
          {/* URL Input Card */}
          <div className="bg-surface-secondary border border-white/5 rounded-3xl p-8 shadow-2xl overflow-hidden relative">
            <div className="absolute top-0 right-0 w-64 h-64 bg-brand/5 rounded-full blur-[80px] -translate-y-1/2 translate-x-1/2" />
            
            <label className="text-xs font-bold text-neutral-500 uppercase tracking-widest block mb-4 relative z-10">Target Gateway URL</label>
            <div className="flex flex-col md:flex-row gap-4 relative z-10">
              <div className="relative flex-1 group">
                <div className="absolute left-4 top-1/2 -translate-y-1/2 text-neutral-600 group-focus-within:text-brand transition-colors">
                  <LinkIcon size={18} />
                </div>
                <input
                  type="text"
                  placeholder="https://your-ngrok-id.ngrok-free.app"
                  value={url}
                  onChange={(e) => setUrl(e.target.value)}
                  className="w-full bg-white/5 border border-white/10 rounded-2xl pl-12 pr-4 py-4 text-white font-medium focus:outline-none focus:border-brand/40 transition-all placeholder:text-neutral-700"
                />
              </div>
              <button
                onClick={runDiagnostics}
                disabled={status === "checking"}
                className="px-8 py-4 bg-brand hover:bg-[#00ffd0] disabled:bg-neutral-800 disabled:text-neutral-500 text-black font-bold rounded-2xl transition-all shadow-lg active:scale-95 flex items-center justify-center gap-2 min-w-[200px]"
              >
                {status === "checking" ? (
                  <>
                    <RefreshCwIcon size={20} className="animate-spin" />
                    Checking...
                  </>
                ) : (
                  <>
                    <ShieldCheckIcon size={20} />
                    Run Diagnostics
                  </>
                )}
              </button>
            </div>
            {lastCheck && (
              <p className="mt-4 text-[11px] text-neutral-600 font-medium relative z-10 flex items-center gap-1.5">
                <ClockIcon size={12} className="opacity-50" />
                Last check: {lastCheck.toLocaleTimeString()}
              </p>
            )}
          </div>

          {/* Status Display */}
          <AnimatePresence mode="wait">
            {status === "healthy" && (
              <motion.div
                key="healthy"
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: -20 }}
                className="bg-green-500/10 border border-green-500/20 rounded-3xl p-8 flex items-start gap-6 shadow-[0_0_40px_rgba(34,197,94,0.05)]"
              >
                <div className="w-12 h-12 rounded-2xl bg-green-500/20 flex items-center justify-center text-green-500 shrink-0 shadow-[0_0_20px_rgba(34,197,94,0.2)]">
                  <CheckIcon size={24} />
                </div>
                <div className="flex-1">
                  <h3 className="text-lg font-bold text-brand mb-2 font-outfit">Gateway Operational</h3>
                  <p className="text-[15px] text-brand/70 font-medium leading-relaxed">
                    Diagnostics passed! Your OpenClaw gateway is active and the Chat Completions endpoint is correctly enabled.
                  </p>
                </div>
              </motion.div>
            )}

            {status === "error_404" && (
              <motion.div
                key="error_404"
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: -20 }}
                className="space-y-6"
              >
                <div className="bg-red-500/10 border border-red-500/20 rounded-3xl p-8 flex items-start gap-6 shadow-[0_0_40px_rgba(239,68,68,0.05)]">
                  <div className="w-12 h-12 rounded-2xl bg-red-500/20 flex items-center justify-center text-red-500 shrink-0 shadow-[0_0_20px_rgba(239,68,68,0.2)]">
                    <AlertCircleIcon size={24} />
                  </div>
                  <div>
                    <h3 className="text-lg font-bold text-red-400 mb-2 font-outfit">Endpoint Disabled (404)</h3>
                    <p className="text-[15px] text-red-400/70 font-medium leading-relaxed">
                      The gateway is reachable, but the <code className="text-red-400 bg-red-400/10 px-1.5 py-0.5 rounded font-mono">chat/completions</code> endpoint is toggled OFF.
                    </p>
                  </div>
                </div>

                <div className="bg-surface-card border border-white/5 rounded-3xl overflow-hidden shadow-2xl">
                  <div className="p-6 border-b border-white/5 bg-white/[0.02]">
                    <h4 className="text-sm font-bold text-white uppercase tracking-widest flex items-center gap-2">
                      <TerminalIcon size={16} className="text-brand" />
                      Resolution Steps
                    </h4>
                  </div>
                  <div className="p-8 space-y-8">
                    <div className="space-y-4">
                      <p className="text-[14px] text-neutral-400 font-medium">1. Run this command to enable the endpoint:</p>
                      <CopyableCommand command="openclaw config set gateway.http.endpoints.chatCompletions.enabled true" />
                    </div>
                    
                    <div className="space-y-4">
                      <p className="text-[14px] text-neutral-400 font-medium">2. Restart your gateway:</p>
                      <CopyableCommand command="openclaw gateway restart" />
                    </div>
                  </div>
                </div>
              </motion.div>
            )}

            {status === "error_connection" && (
              <motion.div
                key="error_connection"
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: -20 }}
                className="bg-red-500/10 border border-red-500/20 rounded-3xl p-8 flex items-start gap-6 shadow-[0_0_40px_rgba(239,68,68,0.05)]"
              >
                 <div className="w-12 h-12 rounded-2xl bg-red-500/20 flex items-center justify-center text-red-500 shrink-0 shadow-[0_0_20px_rgba(239,68,68,0.2)]">
                  <CrossIcon size={24} />
                </div>
                <div className="flex-1">
                  <h3 className="text-lg font-bold text-red-400 mb-2 font-outfit">Gateway Unreachable</h3>
                  <p className="text-[15px] text-red-400/70 font-medium leading-relaxed mb-4">
                    {apiError || "Could not reach the gateway. Please verify your connection settings."}
                  </p>
                  
                  <GatewayChecklist />
                </div>
              </motion.div>
            )}
          </AnimatePresence>
        </div>
      </div>
    </div>
  );
}



// ─── Copyable Command Component ──────────────────────────────────────────────
const CopyableCommand = ({ command }: { command: string }) => {
  const [copied, setCopied] = useState(false);
  const handleCopy = () => {
    navigator.clipboard.writeText(command);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };
  return (
    <div className="group relative flex items-center justify-between bg-black/40 border border-white/5 rounded-xl p-3.5 transition-all hover:bg-black/60 shadow-inner">
      <code className="text-[15px] font-mono text-brand break-all pr-10 font-medium leading-tight">{command}</code>
      <button 
        onClick={handleCopy}
        className="absolute right-3 p-2 rounded-lg hover:bg-white/5 text-neutral-500 hover:text-white transition-all active:scale-90 border border-transparent hover:border-white/10"
        title="Copy to clipboard"
      >
        {copied ? (
          <motion.div initial={{ scale: 0.5 }} animate={{ scale: 1 }}>
            <CheckIcon size={16} className="text-brand" />
          </motion.div>
        ) : (
          <CopyIcon size={16} />
        )}
      </button>
    </div>
  );
};

// ─── Health Alert Notification ──────────────────────────────────────────────
function HealthAlertNotification({ 
  onClose, 
  onFix 
}: { 
  onClose: () => void, 
  onFix: () => void 
}) {
  return (
    <motion.div
      initial={{ opacity: 0, x: 100, scale: 0.9 }}
      animate={{ opacity: 1, x: 0, scale: 1 }}
      exit={{ opacity: 0, x: 100, scale: 0.9 }}
      transition={{ type: "spring", stiffness: 300, damping: 30 }}
      className="fixed bottom-8 right-8 z-[100] max-w-md w-full"
    >
      <div className="bg-surface-secondary/80 backdrop-blur-xl border border-red-500/20 rounded-2xl p-5 shadow-[0_20px_50px_rgba(0,0,0,0.5)] overflow-hidden relative group">
        <div className="absolute top-0 left-0 w-1 h-full bg-red-500/60" />
        
        <div className="flex gap-4">
          <div className="w-10 h-10 rounded-xl bg-red-500/10 flex items-center justify-center text-red-500 shrink-0 border border-red-500/20">
            <AlertCircleIcon size={20} className="animate-pulse" />
          </div>
          
          <div className="flex-1 space-y-4">
            <div>
              <h4 className="text-[16px] font-bold text-white tracking-tight">Gateway Connection Issue</h4>
            </div>

            <div className="pt-2">
              <p className="text-[13px] text-neutral-400 leading-relaxed font-medium">
                We detected connectivity issues with your gateway. Please fix them in the Gateway Doctor to resume your session.
              </p>
            </div>
            
            <div className="flex items-center gap-4 pt-2">
              <button
                onClick={onFix}
                className="px-5 py-2.5 bg-red-500/10 hover:bg-red-500/20 text-red-400 text-[12px] font-bold rounded-xl border border-red-500/20 transition-all active:scale-95 shadow-lg shadow-red-500/5"
              >
                Fix in Gateway Doctor
              </button>
              <button
                onClick={onClose}
                className="text-[12px] text-neutral-500 hover:text-white transition-colors font-semibold"
              >
                Dismiss
              </button>
            </div>
          </div>
        </div>

        <button 
          onClick={onClose}
          className="absolute top-4 right-4 text-neutral-600 hover:text-white transition-colors"
        >
          <CloseIcon size={14} />
        </button>
      </div>
    </motion.div>
  );
}

// ─── Main Page ───────────────────────────────────────────────────────────────

function ClientPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [room] = useState(new Room());
  const [activeSession, setActiveSession] = useState("Library");
  const initialSessionApplied = useRef(false);
  const botIdParamApplied = useRef(false);
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
  const [isAvatarPickerOpen, setIsAvatarPickerOpen] = useState(false);
  const [isRecallModalOpen, setIsRecallModalOpen] = useState(false);
  const [showCreditModal, setShowCreditModal] = useState(false);
  const [creditModalConfig, setCreditModalConfig] = useState<{ title: string; message: string; type: "credit" | "concurrency" }>({
    title: "Low Credits",
    message: "Your account balance is low. Please add credits to continue using our voice assistants.",
    type: "credit"
  });
  const [isValidatingCredit, setIsValidatingCredit] = useState(false);
  const user = useUser();
  const [authChecked, setAuthChecked] = useState(false);
  const [avatars, setAvatars] = useState<AvatarItem[]>(AVATARS);
  const [apiError, setApiError] = useState<string | null>(null);
  const apiKeyInitialized = useRef(false);

  // Session config state
  const [config, setConfig] = useState<typeof DEFAULTS>(DEFAULTS);
  const [bots, setBots] = useState<AgentBot[]>([]);
  const [profileId, setProfileId] = useState<string | null>(null);
  const [isLoadingBots, setIsLoadingBots] = useState(false);
  const [editingBotId, setEditingBotId] = useState<string | null>(null);
  const [editingAgent, setEditingAgent] = useState<AgentBot | null>(null);
  const [selectedAgentId, setSelectedAgentId] = useState<string | null>(null);
  const selectedAgentIdRef = useRef<string | null>(null);
  const externalAgentIdRef = useRef<string | null>(null);
  const [dbLastConfig, setDbLastConfig] = useState<any>(null);
  const [botHealth, setBotHealth] = useState<Record<string, 'healthy' | 'unhealthy' | 'checking' | 'unknown'>>({});
  const [showHealthAlert, setShowHealthAlert] = useState(false);

  useEffect(() => {
    if (initialSessionApplied.current) return;

    const sessionParam = searchParams.get("session");
    const botIdParam = searchParams.get("botId");

    if (sessionParam && DASHBOARD_SESSIONS.has(sessionParam)) {
      // If we have a botId, we want the botId useEffect to handle the session transition
      // after it has configured the bot. This prevents a race condition where DirectCall
      // starts with empty config.
      if (sessionParam === "DirectCall" && botIdParam) {
        console.log("⏳ DirectCall with botId detected, waiting for bot metadata...");
      } else {
        setActiveSession(sessionParam);
      }
    }

    initialSessionApplied.current = true;
  }, [searchParams]);

  useEffect(() => {
    if (botIdParamApplied.current || bots.length === 0) return;

    const botId = searchParams.get("botId");
    if (botId) {
      const bot = bots.find(b => b.id === botId);
      if (bot) {
        console.log("🤖 Found bot for DirectCall:", bot.agent_name);
        handleQuickCallSelect(bot);
        botIdParamApplied.current = true;
      } else if (!isLoadingBots) {
        console.warn("⚠️ Bot not found for botId:", botId);
        // Fallback to library if bot not found
        setActiveSession("Library");
        botIdParamApplied.current = true;
      }
    }
  }, [searchParams, bots, isLoadingBots]);

  // Conversation tracking state
  const [sessionTranscript, setSessionTranscript] = useState<any[]>([]);
  const transcriptRef = useRef<any[]>([]);
  const [sessionStartTime, setSessionStartTime] = useState<number | null>(null);
  const startTimeRef = useRef<number | null>(null);
  const [conversations, setConversations] = useState<any[]>([]);
  const [isLoadingConversations, setIsLoadingConversations] = useState(false);
  const [selectedConversation, setSelectedConversation] = useState<any | null>(null);
  const finalSegmentIds = useRef<Set<string>>(new Set());
  const segmentsMapRef = useRef<Map<string, any>>(new Map());
  const configRef = useRef(config);
  const activeSessionRef = useRef(activeSession);
  const technicalSessionKeyRef = useRef<string>("");
  const botsRef = useRef<AgentBot[]>([]);
  const currentConversationIdRef = useRef<string | null>(null);
  const currentJobIdRef = useRef<string | null>(null);
  const connectionLockRef = useRef(false);

  // Sync refs with state to ensure handleDisconnected sees latest values
  useEffect(() => {
    configRef.current = config;
  }, [config]);

  useEffect(() => {
    botsRef.current = bots;
  }, [bots]);

  useEffect(() => {
    activeSessionRef.current = activeSession;
  }, [activeSession]);

  // Robust Transcription Tracking via Hook
  // (Moved to TranscriptSynchronizer component to stay within RoomContext)

  // ─── Automated Health Checks ────────────────────────────────────────────────
  const checkAllBotsHealth = useCallback(async (currentBots: AgentBot[]) => {
    if (currentBots.length === 0) return;

    // Group unique gateways to avoid redundant pings
    const uniqueConfigs = new Set<string>();
    const tasks: Promise<any>[] = [];

    currentBots.forEach(bot => {
      if (!bot.config?.openclaw_url) return;
      const configKey = `${bot.config.openclaw_url.replace(/\/$/, "")}|${bot.config.gateway_token ?? ""}`;
      if (!uniqueConfigs.has(configKey)) {
        uniqueConfigs.add(configKey);

        // Mark as checking
        setBotHealth(prev => ({ ...prev, [configKey]: 'checking' }));

        // Fire background check
        const check = async () => {
          try {
            const res = await fetch('/api/doctor/check', {
              method: "POST",
              headers: { "Content-Type": "application/json" },
              body: JSON.stringify({
                url: bot.config.openclaw_url,
                token: bot.config.gateway_token ?? "",
                sessionKey: bot.config.session_key || ""
              })
            });
            const data = await res.json();
            const isHealthy = data.status === 200;
            setBotHealth(prev => ({ 
              ...prev, 
              [configKey]: isHealthy ? 'healthy' : 'unhealthy' 
            }));
            
            // Trigger global alert if a bot is unhealthy
            if (!isHealthy) {
              setShowHealthAlert(true);
            }
          } catch (err) {
            setBotHealth(prev => ({ ...prev, [configKey]: 'unhealthy' }));
            setShowHealthAlert(true);
          }
        };
        tasks.push(check());
      }
    });
  }, []);

  // Trigger health check when bots are loaded
  useEffect(() => {
    if (bots.length > 0) {
      checkAllBotsHealth(bots);
    }
  }, [bots, checkAllBotsHealth]);

  // 1. Initial config from localStorage
  useEffect(() => {
    const saved = localStorage.getItem("openclaw_config");
    if (saved) {
      try {
        const parsed = JSON.parse(saved);
        if (parsed.sessionKey) {
          parsed.sessionKey = stripSessionKey(parsed.sessionKey);
        }
        setConfig({ ...DEFAULTS, ...parsed });
      } catch (e) {}
    }
  }, []);

  // 2. Fetch profile and bots once authenticated
  useEffect(() => {
    if (authChecked && user) {
      const initData = async () => {
        // Initialize backend API key once (guard prevents re-runs on user object refresh)
        if (!apiKeyInitialized.current) {
          apiKeyInitialized.current = true;
          try {
            const tokens = await user.currentSession.getTokens();
            if (tokens?.accessToken) {
              // Ensure user exists in backend before fetching workspace/API key.
              // POST /v1/user is idempotent — a 409 Conflict (already exists) is fine.
              const nameParts = (user.displayName ?? "").trim().split(/\s+/).filter(Boolean);
              const firstName = nameParts[0] ?? "";
              const email = user.primaryEmail ?? "";
              await createUserServiceServer({
                id: user.id,
                email,
                first_name: firstName,
                last_name: nameParts.slice(1).join(" "),
                // company is used as org name in the backend; fall back to firstName then email prefix
                company: firstName || email.split("@")[0] || "ClawdFace User",
                password_hash: "",
              });

              await initDefaultApiKey(user.id, tokens.accessToken);
            }
          } catch (err) {
            console.error("API key init error:", err);
          }
        }

        const email = user.primaryEmail || user.displayName;
        if (email) {
          try {
            const profile = await syncUserAction(email);
            if (profile) {
              setProfileId(profile.id);
              setIsLoadingBots(true);
              const initApiKey = localStorage.getItem("defaultApiKey") ?? "";
              const { data: agentData } = await getAgents(initApiKey);
              setBots(agentData ?? []);
              
              // Also fetch avatars here so it uses the same fresh key
              try {
                const { data: avatarData } = await fetchAvatars(initApiKey);
                if (avatarData && avatarData.length > 0) setAvatars(avatarData);
              } catch (avErr) {
                console.error("Avatar fetch error:", avErr);
              }

              setIsLoadingBots(false);
              
              if (profile.last_config) {
                setDbLastConfig(profile.last_config);
                // If nothing in localStorage, initialize form from DB config
                if (!localStorage.getItem("openclaw_config")) {
                  const lastCfg = { ...DEFAULTS, ...profile.last_config };
                  if (lastCfg.sessionKey) {
                    lastCfg.sessionKey = stripSessionKey(lastCfg.sessionKey);
                  }
                  setConfig(lastCfg);
                  localStorage.setItem("openclaw_config", JSON.stringify(lastCfg));
                }
              }
            }
          } catch (err) {
            console.error("Initialization error:", err);
            setIsLoadingBots(false);
          }
        }
      };
      initData();
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [authChecked, user?.id]);


  useEffect(() => {
    if (user === null) {
      router.replace("/log-in");
    } else if (user && !user.primaryEmailVerified) {
      router.replace("/email-not-verified");
    } else if (user) {
      setAuthChecked(true);
    }
  }, [user, router]);
  
  // Fetch conversations when switching to the Conversations section
  useEffect(() => {
    if (activeSession === "Conversations" && authChecked) {
      const loadConversations = async () => {
        setIsLoadingConversations(true);
        try {
          const apiKey = localStorage.getItem("defaultApiKey") ?? "";
          const { data } = await getConversations(apiKey);
          setConversations(data ?? []);
        } catch (err) {
          console.error("Failed to fetch conversations:", err);
        } finally {
          setIsLoadingConversations(false);
        }
      };
      loadConversations();
    }
  }, [activeSession, authChecked]);

  const generateSessionId = () => {
    const now = new Date();
    const pad = (n: number) => n.toString().padStart(2, '0');
    
    const year = now.getFullYear();
    const month = pad(now.getMonth() + 1);
    const day = pad(now.getDate());
    const hours = pad(now.getHours());
    const minutes = pad(now.getMinutes());
    const seconds = pad(now.getSeconds());
    
    return `session-${year}-${month}-${day}T${hours}-${minutes}-${seconds}`;
  };

  const onConnectButtonClicked = useCallback(async (forcedSessionKey?: string, forcedConfig?: typeof DEFAULTS) => {
    if (room.state !== "disconnected" || isValidatingCredit || connectionLockRef.current) {
      return;
    }
    
    connectionLockRef.current = true;
    setIsValidatingCredit(true);
    setApiError(null);

    try {
      // HARD RESET: Clear all previous session data before starting a new one
      setSessionTranscript([]);
      transcriptRef.current = [];
      setSessionStartTime(null);
      startTimeRef.current = null;
      segmentsMapRef.current.clear();
      finalSegmentIds.current.clear();
      currentConversationIdRef.current = null;
      currentJobIdRef.current = null;

      const activeConfig = forcedConfig || config;

      // 1. Persist config to localStorage (Works on Vercel)
      localStorage.setItem("openclaw_config", JSON.stringify(activeConfig));

      // 2. Sync to Supabase & local files
      const email = user?.primaryEmail || user?.displayName;
      if (email) {
        await updateLastConfig(email, activeConfig);
        try {
          await fetch("/api/user-config", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ email: email, config: activeConfig }),
          });
        } catch (err) {
          console.warn("Local sync skipped (expected on production)");
        }
      }

      const finalSessionKey = `agent:main:${generateSessionId()}`;

      const finalConfig = {
        ...activeConfig,
        avatarId: activeConfig.avatarId || (avatars && avatars[0]?.id) || "",
        sessionKey: finalSessionKey,
        botName: activeConfig.botName || (avatars && avatars.find(a => a.id === activeConfig.avatarId)?.name) || "Bot",
        enable_thinking: activeConfig.thinkingEnabled,
        thinking_delay: activeConfig.thinkingDelay,
      };

      // Update config state but keep sessionKey clean for UI
      setConfig({
        ...finalConfig,
        sessionKey: stripSessionKey(finalSessionKey)
      });

      // Store the full technical session key for history persistence
      technicalSessionKeyRef.current = finalSessionKey;

      // --- Credit Validation Step ---
      const validationEmail = user?.primaryEmail || user?.displayName || "";
      
      if (!validationEmail) {
        console.error("❌ Credit Validation - Missing user email. Validation cannot proceed.");
        setApiError("Authentication required. Please sign in again.");
        return;
      }

      const resolvedAgentId = selectedAgentIdRef.current || externalAgentIdRef.current || "";

      if (!resolvedAgentId) {
        setApiError("Unable to identify the agent. Please re-select your bot from the library.");
        return;
      }

      const validationPayload = {
        agentId: resolvedAgentId,
        userName: validationEmail,
        userId: validationEmail,
        context: {
          text: ""
        },
        mode: "vtva",
        metadata: {
          active: "false" // FIX: Prevent premature session creation on backend during validation
        }
      };

      console.log("🔍 Credit Validation - Outgoing Payload:", JSON.stringify(validationPayload, null, 2));

      const baseUrl = (process.env.NEXT_PUBLIC_BASE_URL || "https://qaapi.clawdface.ai").replace(/\/$/, "");
      const apiKey = localStorage.getItem("defaultApiKey") ?? "";

      if (!apiKey) {
        setApiError("Missing API key. Please sign in again.");
        return;
      }
      
      let token = "";
      try {
        // @ts-ignore
        token = await user?.getAccessToken() || "";
      } catch (e) {
        console.warn("⚠️ Credit Validation - Failed to get access token:", e);
      }

      const validationResponse = await fetch(`${baseUrl}/v1/conversation`, {
        method: "POST",
        headers: {
          "X-API-Key": apiKey,
          "Content-Type": "application/json",
          ...(token ? { "Authorization": `Bearer ${token}` } : {})
        },
        body: JSON.stringify(validationPayload)
      });

      console.log("🔍 Credit Validation - Status:", validationResponse.status);
      
      if (validationResponse.status === 403) {
        console.log("🚫 Credit Validation - 403 Forbidden: Redirecting to Billing");
        setCreditModalConfig({
          title: "Low Credits",
          message: "Your account balance is low. Please add credits in the Payments section to continue using our voice assistants.",
          type: "credit"
        });
        setShowCreditModal(true);
        return;
      }

      if (validationResponse.status !== 200) {
        const errorText = await validationResponse.text().catch(() => "");
        
        if (validationResponse.status === 500 && errorText.includes("concurrent session limit")) {
          setCreditModalConfig({
            title: "Session Limit Reached",
            message: "You already have an active session. Please close your other session or wait 10 seconds and try again.",
            type: "concurrency"
          });
          setShowCreditModal(true);
        } else {
          console.error(`❌ Validation failed (${validationResponse.status}):`, errorText);
          setApiError(`Session start failed (Error ${validationResponse.status}). Please try again later.`);
        }
        return;
      }

      // If we reach here, validation passed
      const validationData = await validationResponse.json().catch(() => ({}));
      console.log("🔍 Credit Validation - Success Response:", JSON.stringify(validationData, null, 2));

      // Capture conversation and job IDs for usage tracking
      const receivedConvId = validationData.conversation_id || validationData.conversationId;
      const receivedJobId = validationData.job_id || validationData.jobId;

      if (!receivedConvId) {
        console.warn("⚠️ Credit Validation - No conversation_id returned. Reverting to fallback mode.");
      } else {
        currentConversationIdRef.current = receivedConvId;
        console.log(`✅ Credit Validation - Assigned ID: ${receivedConvId}`);
      }
      
      if (receivedJobId) {
        currentJobIdRef.current = receivedJobId;
      }

      // --- End Credit Validation ---

      // --- Connection Step ---
      // Merge the conversation_id/job_id from credit validation into the config
      // so the agent receives them in ctx.job.metadata and can post accurate usage.
      const connectionConfig = {
        ...finalConfig,
        conversation_id: currentConversationIdRef.current || undefined,
        job_id: currentJobIdRef.current || undefined,
      };
      console.log("🚀 Connecting with config:", connectionConfig);
      const response = await fetch("/api/connection-details", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(connectionConfig),
      });

      const connectionDetailsData: ConnectionDetails = await response.json();
      await room.connect(connectionDetailsData.serverUrl, connectionDetailsData.participantToken, {
        // @ts-ignore
        signalTimeout: 30000, 
        connectTimeout: 30000,
      });
      await room.localParticipant.setMicrophoneEnabled(true);

    } catch (err: any) {
      console.error("❌ Connection Error:", err);
      setApiError("An error occurred while starting the session.");
    } finally {
      connectionLockRef.current = false;
      setIsValidatingCredit(false);
    }
  }, [room, config, user, avatars]);


  const handleQuickCallSelect = async (bot: AgentBot) => {
    setSelectedAgentId(bot.id);
    selectedAgentIdRef.current = bot.id;
    
    // Store external agent identifiers separately; backend conversation calls use the internal UUID.
    // @ts-ignore
    const externalId = bot.agent_id || stripSessionKey(bot.config?.session_key || "");
    
    externalAgentIdRef.current = externalId || "";

    setConfig({
      ...config,
      openclawUrl: bot.config.openclaw_url,
      gatewayToken: bot.config.gateway_token,
      sessionKey: bot.config.session_key ? stripSessionKey(bot.config.session_key) : "",
      avatarId: getBotAvatarId(bot),
      botName: bot.agent_name,
      thinkingEnabled: String(bot.config.thinking_enabled ?? true),
      thinkingDelay: String(bot.config.thinking_delay ?? 5.0),
    });
    setActiveSession("DirectCall");
    setIsMobileMenuOpen(false);
  };

  useEffect(() => {
    room.on(RoomEvent.MediaDevicesError, onDeviceFailure);
    
    // Manual Transcription tracking removed in favor of useCombinedTranscriptions hook

    const handleConnected = () => {
      console.log("🚀 Session Connected Logic Triggered");
      const now = Date.now();
      setSessionStartTime(now);
      startTimeRef.current = now;
      setSessionTranscript([]);
      transcriptRef.current = [];
      segmentsMapRef.current.clear();
      finalSegmentIds.current.clear();
    };

    const handleDisconnected = async (reason?: DisconnectReason) => {
      console.log("📡 handleDisconnected Logic Triggered, Reason:", reason);
      
      const currentSessionType = activeSessionRef.current;
      const email = user?.primaryEmail || user?.displayName;

      // Usage reporting is handled exclusively by agent.py at session teardown.
      // It includes full STT/LLM/TTS metrics from the LiveKit usage collector.

      if (email) {
        try {
          const convApiKey = localStorage.getItem("defaultApiKey") ?? "";
          if (convApiKey) {
            const { data: convData } = await getConversations(convApiKey);
            setConversations(convData ?? []);
          }
        } catch (err) {
          console.error("⛔ Conversation refresh error:", err);
        }
      }

      // REDIRECTION: If this was a Direct Call, go back to the library automatically
      if (currentSessionType === "DirectCall") {
        console.log("↩️ Direct Call ended, returning to Library");
        setActiveSession("Library");
      }
      
      // Cleanup for next session
      setSessionStartTime(null);
      startTimeRef.current = null;
      segmentsMapRef.current.clear();
      finalSegmentIds.current.clear();
      setSessionTranscript([]);
      transcriptRef.current = [];
    };

    const handleDataReceived = (payload: Uint8Array) => {
      try {
        const strData = new TextDecoder().decode(payload);
        const data = JSON.parse(strData);
        if (data.type === "usage_status") {
          if (data.status === "success") {
            console.log("✅ [Backend Telemetry] Usage reported successfully:", data.id);
          } else {
            console.warn("❌ [Backend Telemetry] Usage reporting failed:", data.error || "Unknown error");
          }
        }
      } catch (e) {
        // Not our message or not JSON
      }
    };

    room.on(RoomEvent.Connected, handleConnected);
    room.on(RoomEvent.Disconnected, handleDisconnected);
    room.on(RoomEvent.DataReceived, handleDataReceived);

    // Close session when user closes/refreshes the tab
    const handleBeforeUnload = async () => {
      if (room.state === "connected" && currentConversationIdRef.current) {
        const baseUrl = (process.env.NEXT_PUBLIC_BASE_URL || "https://qaapi.clawdface.ai").replace(/\/$/, "");
        navigator.sendBeacon(
          `${baseUrl}/v1/usage/update`,
          JSON.stringify({
            conversation_id: currentConversationIdRef.current,
            status: "TERMINATED",
            message: "client_unload"
          })
        );
      }
    };

    window.addEventListener("beforeunload", handleBeforeUnload);

    return () => {
      console.log("🧹 Cleaning up Room listeners");
      room.off(RoomEvent.MediaDevicesError, onDeviceFailure);
      room.off(RoomEvent.Connected, handleConnected);
      room.off(RoomEvent.Disconnected, handleDisconnected);
      room.off(RoomEvent.DataReceived, handleDataReceived);
      window.removeEventListener("beforeunload", handleBeforeUnload);
    };
  }, [room, user]);

  // ─── Auto-disconnect call on navigation ────────────────────────────────────
  useEffect(() => {
    // Define sessions that are allowed to maintain an active call
    const callCapableSessions = ["DirectCall", "My Bot"];
    
    // If the room is active and we navigate to a non-call session, disconnect.
    if (room && (room.state === "connected" || room.state === "connecting" || room.state === "reconnecting")) {
      if (!callCapableSessions.includes(activeSession)) {
        console.log("🚶 Navigation event: User moved to non-call section:", activeSession, "- auto-disconnecting call");
        room.disconnect();
      }
    }
  }, [activeSession, room]);

  const refreshBots = async () => {
    const refreshKey = localStorage.getItem("defaultApiKey") ?? "";
    setIsLoadingBots(true);
    const { data: agentData } = await getAgents(refreshKey);
    setBots(agentData ?? []);
    setIsLoadingBots(false);
  };

  if (!authChecked) {
    return (
      <div className="h-screen w-screen bg-surface-secondary flex items-center justify-center">
        <svg className="animate-spin text-brand" width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <circle cx="12" cy="12" r="10" opacity="0.25"/>
          <path d="M22 12a10 10 0 0 1-10 10" opacity="0.9"/>
        </svg>
      </div>
    );
  }


  const handleSaveBot = async () => {
    if (!profileId && !editingBotId) return;
    setIsLoadingBots(true);
    try {
      const apiKey = localStorage.getItem("defaultApiKey") ?? "";

      const selectedAvatar = avatars.find(a => a.id === config.avatarId);
      const botName = config.botName || (selectedAvatar ? `${selectedAvatar.name}'s Bot` : "My New Bot");
      
      let botToUse;
      if (editingBotId) {
        const agentToUpdate = editingAgent ?? bots.find(bot => bot.id === editingBotId);

        if (!agentToUpdate) {
          throw new Error("Unable to find the selected agent to update.");
        }

        if (apiKey) {
          const agentPayload = {
            agent_name: botName,
            agent_system_prompt: agentToUpdate.agent_system_prompt ?? "",
            default_system_prompt: agentToUpdate.default_system_prompt ?? false,
            email: agentToUpdate.email || user?.primaryEmail || user?.displayName || "",
            config: {
              ...(agentToUpdate.config ?? {}),
              openclaw_url: config.openclawUrl,
              gateway_token: config.gatewayToken,
              session_key: config.sessionKey,
              thinking_enabled: config.thinkingEnabled === "true",
              thinking_delay: parseFloat(config.thinkingDelay || "5.0"),
            },
            tools: agentToUpdate.tools ?? {},
            avatars: withPrimaryAvatar(agentToUpdate.avatars, config.avatarId),
            knowledge_base: agentToUpdate.knowledge_base ?? [],
            mcp: agentToUpdate.mcp ?? [],
            tool: agentToUpdate.tool ?? [],
            integration: agentToUpdate.integration ?? [],
            record: agentToUpdate.record ?? false,
            callback_url: agentToUpdate.callback_url ?? "",
            callback_events: agentToUpdate.callback_events ?? [],
            is_public: agentToUpdate.is_public ?? true,
            is_active: true,
            type: agentToUpdate.type ?? "etev",
            add_on: agentToUpdate.add_on ?? [],
          };

          const { error } = await updateAgent(apiKey, agentToUpdate.id, agentPayload);
          if (error) throw new Error(error);
        }
      } else {
        // Create new bot in Supabase
        botToUse = await createBot({
          user_id: profileId ?? "",
          name: botName,
          avatar_id: config.avatarId,
          openclaw_url: config.openclawUrl,
          gateway_token: config.gatewayToken,
          session_key: config.sessionKey,
          voice_id: "default",
          thinking_enabled: config.thinkingEnabled,
          thinking_delay: config.thinkingDelay,
        });
      }

      // Sync with API
      if (!editingBotId && apiKey && botToUse?.agent_email) {
        const agentPayload = {
          agent_name: botName,
          agent_system_prompt: "",
          email: botToUse.agent_email,
          config: {
            openclaw_url: config.openclawUrl,
            gateway_token: config.gatewayToken,
            session_key: config.sessionKey,
            thinking_enabled: config.thinkingEnabled === "true",
            thinking_delay: parseFloat(config.thinkingDelay || "5.0"),
          },
          tools: {},
          avatars: [{ avatar_key_id: config.avatarId }],
          is_active: true,
          is_public: false,
          type: "vtva",
          add_on: [],
        };

        await createAgent(apiKey, agentPayload);
      }
      
      // Refresh and reset
      await refreshBots();
      setEditingBotId(null);
      setEditingAgent(null);
      setConfig(DEFAULTS);
      setActiveSession("Library");
    } catch (err: any) {
      console.error("Failed to save/update bot:", err.message || err);
    } finally {
      setIsLoadingBots(false);
    }
  };



  return (
    <AvatarsContext.Provider value={avatars}>
    <main data-lk-theme="default" className="h-[100dvh] w-screen bg-canvas flex overflow-hidden font-[Inter] text-white">
        <Sidebar
          activeSession={activeSession}
          setActiveSession={(session) => {
            setActiveSession(session);
            // Clear editing state when navigating via sidebar
            if (session !== "AddBot") {
              setEditingBotId(null);
              setEditingAgent(null);
              // Only reset config if we're not moving to a call-related session
              if (session !== "DirectCall" && session !== "My Bot") {
                setConfig(DEFAULTS);
              }
            }
          }}
          isMobileMenuOpen={isMobileMenuOpen}
          setIsMobileMenuOpen={setIsMobileMenuOpen}
          bots={bots}
          onQuickCall={handleQuickCallSelect}
          avatars={avatars}
        />

      <div className="flex-1 h-full w-full overflow-hidden flex flex-col relative z-0">
        {/* Mobile Header */}
        <div className="md:hidden flex items-center justify-between px-4 h-14 border-b border-white/5 bg-surface-secondary shrink-0 z-10 shadow-sm transition-all duration-300">
          <div className="flex items-center gap-3">
            <div className="w-7 h-7 shrink-0 relative flex items-center justify-center rounded-lg bg-brand/10 text-brand">
              <Image src="/openclaw.png" alt="Logo" width={18} height={18} className="object-contain drop-shadow-[0_0_4px_rgba(0,227,170,0.5)]" />
            </div>
            <span className="text-white font-bold text-lg leading-none tracking-tight mt-1 font-outfit">ClawdFace</span>
          </div>
          <button 
            onClick={() => setIsMobileMenuOpen(true)} 
            className="text-white/70 hover:text-white p-2.5 rounded-xl bg-white/[0.03] hover:bg-white/[0.08] transition-all border border-white/5 active:scale-95"
          >
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
              <line x1="3" x2="21" y1="12" y2="12"/><line x1="3" x2="21" y1="6" y2="6"/><line x1="3" x2="21" y1="18" y2="18"/>
            </svg>
          </button>
        </div>

        <div className="flex-1 overflow-hidden relative">
          {/* @ts-ignore */}
          <RoomContext.Provider value={room}>
            <RoomAudioRenderer />
            <TranscriptSynchronizer transcriptRef={transcriptRef} startTimeRef={startTimeRef} />
            {activeSession === "My Bot" || activeSession === "AddBot" ? (
              <SimpleVoiceAssistant
                onConnectButtonClicked={onConnectButtonClicked}
                config={config}
                setConfig={setConfig}
                onOpenPicker={() => setIsAvatarPickerOpen(true)}
                onSaveAsBot={handleSaveBot}
                isSavingBot={isLoadingBots}
                isEditing={!!editingBotId}
                onCancelEdit={() => {
                  setEditingBotId(null);
                  setEditingAgent(null);
                  setConfig(DEFAULTS);
                }}
                bots={activeSession === "My Bot" ? bots : []}
                showConnectButton={activeSession === "My Bot"}
                isConnecting={isValidatingCredit || room.state === "connecting"}
              />
            ) : activeSession === "DirectCall" ? (
              <DirectCallDashboard
                config={config}
                autoStart={true}
                onStartCall={async () => {
                  await onConnectButtonClicked();
                }}
                onBack={() => {
                  setEditingBotId(null);
                  setEditingAgent(null);
                  setConfig(DEFAULTS);
                  setActiveSession("Library");
                }}
                isValidating={isValidatingCredit}
              />
            ) : activeSession === "Avatars" ? (
              <AvatarGallery />
            ) : activeSession === "Doctor" ? (
              <DoctorView 
                bots={bots} 
                onHealthUpdate={(id, status) => {
                  const bot = bots.find(b => b.id === id);
                  if (bot) {
                    const healthKey = `${(bot.config?.openclaw_url ?? "").replace(/\/$/, "")}|${bot.config?.gateway_token ?? ""}`;
                    setBotHealth(prev => ({ ...prev, [healthKey]: status }));
                  }
                }}
              />
            ) : activeSession === "Library" ? (
              <BotLibraryView 
                bots={bots} 
                profileId={profileId} 
                onRefresh={refreshBots}
                onSelectBot={(bot) => {
                  setSelectedAgentId(bot.id);
                  selectedAgentIdRef.current = bot.id;
                  
                  // Store external agent identifiers separately; backend conversation calls use the internal UUID.
                  // @ts-ignore
                  const externalId = bot.agent_id || stripSessionKey(bot.config?.session_key || "");
                  
                  externalAgentIdRef.current = externalId || "";

                  const newConfig = {
                    openclawUrl: bot.config.openclaw_url,
                    gatewayToken: bot.config.gateway_token,
                    sessionKey: stripSessionKey(bot.config.session_key || ""),
                    avatarId: getBotAvatarId(bot),
                    botName: bot.agent_name,
                    thinkingEnabled: String(bot.config.thinking_enabled ?? true),
                    thinkingDelay: String(bot.config.thinking_delay ?? 5.0),
                  };
                  setConfig(newConfig);
                  setActiveSession("DirectCall");
                }}
                onEditBot={(bot) => {
                  setEditingBotId(bot.id);
                  setEditingAgent(bot);
                  const editConfig = {
                    openclawUrl: bot.config.openclaw_url,
                    gatewayToken: bot.config.gateway_token,
                    sessionKey: stripSessionKey(bot.config.session_key || ""),
                    avatarId: getBotAvatarId(bot),
                    botName: bot.agent_name,
                    thinkingEnabled: String(bot.config.thinking_enabled ?? true),
                    thinkingDelay: String(bot.config.thinking_delay ?? 5.0),
                  };
                  setConfig(editConfig);
                  setActiveSession("AddBot");
                }}
                botHealth={botHealth}
                isConnecting={isValidatingCredit || room.state === "connecting"}
              />
            ) : activeSession === "Subscription" ? (
              <SubscriptionView />
            ) : activeSession === "Conversations" ? (
              selectedConversation ? (
                <ConversationDetailView 
                  conversation={selectedConversation} 
                  onBack={() => setSelectedConversation(null)} 
                />
              ) : (
                <ConversationsListView
                  isLoading={isLoadingConversations}
                  conversations={conversations}
                  onSelect={async (conv) => {
                    // Fetch full conversation details including transcript
                    try {
                      const apiKey = localStorage.getItem("defaultApiKey") ?? "";
                      const { data, error } = await getConversationById(apiKey, conv.id);
                      if (data) {
                        setSelectedConversation(data);
                      } else {
                        console.error("Failed to fetch conversation details:", error);
                        // Fallback to basic conversation data
                        setSelectedConversation(conv);
                      }
                    } catch (err) {
                      console.error("Error fetching conversation:", err);
                      // Fallback to basic conversation data
                      setSelectedConversation(conv);
                    }
                  }}
                />
              )
            ) : (
              <div className="flex flex-col items-center justify-center h-full text-neutral-400 bg-canvas p-6">
                <div className="text-center space-y-4 max-w-md p-8 border border-white/5 rounded-2xl bg-surface-secondary shadow-2xl relative overflow-hidden">
                  <div className="absolute top-0 left-1/2 -translate-x-1/2 w-32 h-32 bg-brand/5 rounded-full blur-3xl mix-blend-screen pointer-events-none" />
                  <div className="w-16 h-16 bg-white/5 rounded-full flex items-center justify-center mx-auto mb-6 text-brand relative z-10">
                    <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
                      <circle cx="12" cy="12" r="10"/><path d="m4.9 4.9 14.2 14.2"/>
                    </svg>
                  </div>
                  <h2 className="text-2xl font-bold text-white tracking-tight relative z-10">Session Empty</h2>
                  <p className="text-[15px] leading-relaxed relative z-10">
                    The <span className="text-white font-medium">&quot;{activeSession}&quot;</span> session is currently under development.
                  </p>
                  <button
                    onClick={() => setActiveSession("My Bot")}
                    className="relative z-10 mt-6 px-5 py-2.5 bg-brand/10 hover:bg-brand/20 text-brand rounded-lg font-medium transition-all duration-300 text-sm border border-brand/20"
                  >
                    Return to My Agent
                  </button>
                </div>
              </div>
            )}
          </RoomContext.Provider>
        </div>
      </div>
        <AvatarPickerModal
          isOpen={isAvatarPickerOpen}
          onClose={() => setIsAvatarPickerOpen(false)}
          currentId={config.avatarId}
          onSelect={(id) => {
            setConfig({ 
              ...config, 
              avatarId: id,
              botName: (!config.botName || config.botName === "Bot" || config.botName === "My Bot" || config.botName === "Agent" || config.botName === "My Agent") 
                ? (avatars.find(a => a.id === id)?.name || "")
                : config.botName
            });
          }}
        />
        <RecallUrlModal
          isOpen={isRecallModalOpen}
          onClose={() => setIsRecallModalOpen(false)}
          config={config}
        />
        <AnimatePresence>
          {showHealthAlert && (
            <HealthAlertNotification 
              onClose={() => setShowHealthAlert(false)} 
              onFix={() => {
                setActiveSession("Doctor");
                setShowHealthAlert(false);
              }} 
            />
          )}
        </AnimatePresence>

        <CreditModal 
          isOpen={showCreditModal} 
          onClose={() => setShowCreditModal(false)} 
          config={creditModalConfig}
        />
    </main>
    </AvatarsContext.Provider>
  );
}

// ─── Launcher / Session Config Form ──────────────────────────────────────────
function SessionConfigForm({
  onConnect,
  config,
  setConfig,
  onOpenPicker,
  onSaveAsBot,
  isSavingBot,
  isEditing,
  onCancelEdit,
  bots = [],
  isConnecting = false,
  showConnectButton = true,
}: {
  onConnect: (e: React.FormEvent) => void;
  config: any;
  setConfig: (cfg: any) => void;
  onOpenPicker: () => void;
  onSaveAsBot?: () => void;
  isSavingBot?: boolean;
  isEditing?: boolean;
  onCancelEdit?: () => void;
  bots?: AgentBot[];
  isConnecting?: boolean;
  showConnectButton?: boolean;
}) {
  const avatars = useAvatars();
  const selectedAvatar = avatars.find((a) => a.id === config.avatarId);

  const handleConnect = (e: React.MouseEvent | React.FormEvent) => {
    e.preventDefault();
    onConnect(e as any);
  };

  const field = (id: string, label: string, icon: React.ReactNode, placeholder: string, type: string = "text") => (
    <div className="flex flex-col gap-1.5">
      <label className="text-[12px] font-semibold uppercase tracking-[0.08em] text-[#6b7280] flex items-center gap-1.5">
        <span className="text-[#9ca3af]">{icon}</span>
        {label} <span className="text-brand ml-0.5">*</span>
      </label>
      <div className="relative group">
        <input
          type={type}
          id={id}
          value={(config as any)[id]}
          onChange={(e) => setConfig({ ...config, [id]: e.target.value })}
          placeholder={placeholder}
          className="w-full bg-surface border border-[#242424] hover:border-brand/40 rounded-xl py-3 px-4 text-[14px] text-white focus:outline-none focus:border-brand transition-all placeholder:text-[#3a3a3a]"
        />
      </div>
    </div>
  );

  return (
    <motion.div
      initial={{ opacity: 0, y: 24 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0, y: -24 }}
      transition={{ duration: 0.35, ease: [0.09, 1.04, 0.245, 1.055] }}
      className="h-full overflow-y-auto custom-scrollbar"
    >
      <div className="w-full max-w-[620px] mx-auto px-6 py-8">
        <div className="mb-6 text-center">
          <h2 className="text-[22px] font-bold text-white tracking-tight">
            {isEditing ? "Edit Agent Configuration" : (isSavingBot ? "Save Agent to Library" : "Quick Call")}
          </h2>
          <p className="text-[#6b7280] text-[13px] mt-1">
            {isEditing 
              ? "Update your agent settings below" 
              : "Manual configuration for a one-time connection"}
          </p>
        </div>

        <div className={`bg-surface-card border border-[#1f1f1f] rounded-2xl p-5 flex flex-col gap-4 shadow-2xl mx-auto`}>
          {!isEditing && bots.length > 0 && (
            <div className="flex flex-col gap-1.5 pb-4 border-b border-[#1f1f1f]">
              <label className="text-[12px] font-semibold uppercase tracking-[0.08em] text-brand flex items-center gap-1.5">
                <LibraryIcon size={14} className="text-brand" />
                Quick Fill from Library
              </label>
                <select
                  onChange={async (e) => {
                    const botId = e.target.value;
                    if (!botId) return;
                    const selected = bots.find(b => b.id === botId);
                    if (selected) {
                      const newConfig = {
                        openclawUrl: selected.config.openclaw_url,
                        gatewayToken: selected.config.gateway_token,
                        sessionKey: stripSessionKey(selected.config.session_key),
                        avatarId: getBotAvatarId(selected),
                        botName: selected.agent_name,
                      };
                      setConfig(newConfig);
                    }
                  }}
                  className="w-full bg-surface border border-brand/30 rounded-xl py-3 px-4 text-[14px] text-white focus:outline-none focus:border-brand transition-all cursor-pointer font-medium"
                  defaultValue=""
                >
                  <option value="" disabled>Select an agent to fill fields...</option>
                  {bots.map(bot => {
                    const date = bot.created_at ? new Date(bot.created_at) : null;
                    const timestamp = date ? `${date.getMonth() + 1}/${date.getDate()}/${date.getFullYear().toString().slice(-2)} ${date.getHours()}:${date.getMinutes().toString().padStart(2, '0')}` : "";
                    return (
                      <option key={bot.id} value={bot.id}>
                        {bot.agent_name}{timestamp ? ` (${timestamp})` : ""}
                      </option>
                    );
                  })}
                </select>
              </div>
            )}

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                {field("openclawUrl",  "URL",     <LinkIcon />,   "http://localhost:18789")}
                {field("gatewayToken", "Token",    <KeyIcon />,    "Enter token")}
              </div>
              {field("botName",      "Agent Name",         <UserIcon />,   "Enter a custom name for this agent")}

              <div className="flex flex-col gap-1.5">
                <label className="text-[12px] font-semibold uppercase tracking-[0.08em] text-[#6b7280] flex items-center gap-1.5">
                  <span className="text-[#9ca3af]"><UserIcon size={14} /></span>
                  Avatar <span className="text-brand ml-0.5">*</span>
                </label>
                <button
                  onClick={onOpenPicker}
                  className="group relative w-full h-80 rounded-xl bg-surface border-2 border-dashed border-[#242424] hover:border-brand/40 transition-all duration-300 overflow-hidden flex flex-col items-center justify-center gap-2"
                >
                  {selectedAvatar ? (
                    <>
                      <Image 
                        src={selectedAvatar.image} 
                        alt={selectedAvatar.name} 
                        fill 
                        className="object-cover opacity-60 group-hover:opacity-80 transition-opacity"
                      />
                      <div className="absolute inset-0 bg-gradient-to-t from-black/80 via-transparent to-transparent" />
                      <div className="relative z-10 flex flex-col items-center gap-1">
                        <span className="text-white font-bold text-sm tracking-tight">{selectedAvatar.name}</span>
                        <span className="text-[11px] text-brand font-medium uppercase tracking-wider">Change Avatar</span>
                      </div>
                    </>
                  ) : (
                    <>
                      <div className="w-10 h-10 rounded-full bg-white/5 flex items-center justify-center text-[#4b5563] group-hover:text-brand group-hover:bg-brand/10 transition-colors">
                        <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                          <line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/>
                        </svg>
                      </div>
                      <span className="text-[13px] font-bold text-[#4b5563] group-hover:text-white transition-colors">Choose From Existing Avatars</span>
                    </>
                  )}
                </button>
              </div>

              <div className="flex flex-col gap-3 p-4 rounded-2xl bg-white/[0.02] border border-white/5 mt-0 transition-all hover:bg-white/[0.04] shadow-inner">
                <div className="flex items-center justify-between">
                  <div className="flex flex-col gap-0.5">
                    <label className="text-[12px] font-bold uppercase tracking-[0.1em] text-neutral-400 flex items-center gap-2">
                      <SmileIcon size={14} className="text-brand" />
                      Dynamic Thinking
                    </label>
                    <p className="text-[10px] text-neutral-600 font-medium">Auto-trigger filler phrases</p>
                  </div>
                  <label className="relative inline-flex items-center cursor-pointer">
                    <input 
                      type="checkbox" 
                      className="sr-only peer"
                      checked={config.thinkingEnabled === "true"}
                      onChange={(e) => setConfig({ ...config, thinkingEnabled: e.target.checked ? "true" : "false" })}
                    />
                    <div className="w-11 h-5 bg-neutral-800 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-4 after:w-5 after:transition-all peer-checked:bg-brand"></div>
                  </label>
                </div>

                <AnimatePresence mode="wait">
                  {config.thinkingEnabled === "true" && (
                    <motion.div 
                      initial={{ height: 0, opacity: 0 }}
                      animate={{ height: "auto", opacity: 1 }}
                      exit={{ height: 0, opacity: 0 }}
                      transition={{ duration: 0.2, ease: "easeInOut" }}
                      className="overflow-hidden"
                    >
                      <div className="pt-2 flex flex-col gap-4 border-t border-white/5">
                        <div className="flex flex-col gap-3">
                          <div className="flex justify-between items-center text-[11px] font-bold text-neutral-500 uppercase tracking-widest pl-1">
                            <span>Thinking Delay</span>
                          </div>
                          
                          <div className="flex items-center gap-2 bg-black/40 p-1.5 rounded-lg border border-white/5">
                            <button
                              onClick={(e) => {
                                e.preventDefault();
                                const current = parseFloat(config.thinkingDelay || "5.0");
                                if (current > 0.5) {
                                  setConfig({ ...config, thinkingDelay: (current - 0.5).toFixed(1) });
                                }
                              }}
                              className="w-8 h-8 flex items-center justify-center rounded-md bg-white/5 hover:bg-white/10 active:scale-95 transition-all text-brand border border-white/5"
                            >
                              <span className="text-lg font-bold">−</span>
                            </button>
                            
                            <div className="flex-1 flex items-center justify-center gap-1">
                              <input
                                type="number"
                                step="0.5"
                                min="0.5"
                                max="15"
                                value={config.thinkingDelay}
                                onChange={(e) => setConfig({ ...config, thinkingDelay: e.target.value })}
                                className="w-12 bg-transparent text-center text-[15px] font-bold text-white focus:outline-none [appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none"
                              />
                              <span className="text-[11px] font-bold text-neutral-500 uppercase tracking-tight">sec</span>
                            </div>

                            <button
                              onClick={(e) => {
                                e.preventDefault();
                                const current = parseFloat(config.thinkingDelay || "5.0");
                                if (current < 15) {
                                  setConfig({ ...config, thinkingDelay: (current + 0.5).toFixed(1) });
                                }
                              }}
                              className="w-8 h-8 flex items-center justify-center rounded-md bg-white/5 hover:bg-white/10 active:scale-95 transition-all text-brand border border-white/5"
                            >
                              <span className="text-lg font-bold">+</span>
                            </button>
                          </div>
                        </div>
                      </div>
                    </motion.div>
                  )}
                </AnimatePresence>
              </div>

              <div className="flex flex-col gap-3 mt-4">
                  {showConnectButton && (
                    <button
                    onClick={handleConnect}
                    disabled={isConnecting || !config.openclawUrl || !config.gatewayToken}
                    className="w-full py-3.5 rounded-xl font-bold text-[15px] tracking-wide transition-all duration-200
                      bg-brand text-black hover:bg-[#00c994] active:scale-[0.98]
                      disabled:opacity-40 disabled:cursor-not-allowed disabled:active:scale-100
                      shadow-[0_0_24px_rgba(0,227,170,0.25)] hover:shadow-[0_0_32px_rgba(0,227,170,0.35)]
                      flex items-center justify-center gap-2"
                  >
                    {isConnecting ? (
                      <>
                        <svg className="animate-spin" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
                          <circle cx="12" cy="12" r="10" opacity="0.25"/>
                          <path d="M22 12a10 10 0 0 1-10 10" opacity="0.9"/>
                        </svg>
                        Connecting…
                      </>
                    ) : (
                      <>
                        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                          <polygon points="5 3 19 12 5 21 5 3"/>
                        </svg>
                        Start Session
                      </>
                    )}
                  </button>
                )}

                <button
                  onClick={onSaveAsBot}
                  disabled={isSavingBot || !config.openclawUrl}
                  className="w-full py-3 bg-white/[0.03] hover:bg-white/[0.08] disabled:opacity-40 text-white/90 font-semibold rounded-xl transition-all border border-white/5 hover:border-white/10 text-[14px] flex items-center justify-center gap-2 shadow-sm"
                >
                  {isSavingBot ? (
                    <svg className="animate-spin" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3">
                      <circle cx="12" cy="12" r="10" opacity="0.25"/><path d="M22 12a10 10 0 0 1-10 10" opacity="0.9"/>
                    </svg>
                  ) : isEditing ? (
                    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                      <polyline points="20 6 9 17 4 12"/>
                    </svg>
                  ) : (
                    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                      <path d="m19 21-7-4-7 4V5a2 2 0 0 1 2-2h10a2 2 0 0 1 2 2v16z"/>
                    </svg>
                  )}
                  {isEditing ? "Update Agent Configuration" : "Save as new Agent"}
                </button>
                {isEditing && (
                  <button
                    onClick={onCancelEdit}
                    className="w-full py-3 bg-red-500/10 hover:bg-red-500/20 text-red-500 font-semibold rounded-xl transition-all border border-red-500/10 text-[14px]"
                  >
                    Cancel Edit
                  </button>
                )}
              </div>
            </div>

      </div>
    </motion.div>
  );
}

// ─── Credit Error Modal ──────────────────────────────────────────────────────
function CreditModal({
  isOpen,
  onClose,
  config,
}: {
  isOpen: boolean;
  onClose: () => void;
  config: { title: string; message: string; type: "credit" | "concurrency" };
}) {
  const router = useRouter();
  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-[110] flex items-center justify-center p-4">
      <div className="absolute inset-0 bg-black/80 backdrop-blur-sm" onClick={onClose} />
      <motion.div 
        initial={{ opacity: 0, scale: 0.95, y: 20 }}
        animate={{ opacity: 1, scale: 1, y: 0 }}
        className="w-full max-w-md bg-surface-secondary rounded-3xl border border-white/5 shadow-2xl overflow-hidden relative p-8 text-center"
      >
        <div className="w-16 h-16 bg-red-500/10 rounded-full flex items-center justify-center mx-auto mb-6">
          <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="#ef4444" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/>
          </svg>
        </div>
        
        <h2 className="text-2xl font-bold text-white mb-3 font-outfit">{config.title}</h2>
        <p className="text-neutral-400 text-[15px] leading-relaxed mb-8">
          {config.message}
        </p>

        <div className="flex flex-col gap-3">
          <button 
            onClick={() => {
              onClose();
              router.push("/dashboard/settings/billing-and-subscription");
            }}
            className="w-full py-4 bg-brand hover:bg-[#00c994] text-black font-bold rounded-2xl transition-all shadow-[0_0_20px_rgba(0,227,170,0.2)] active:scale-[0.98]"
          >
            Go to Payments
          </button>
          <button 
            onClick={onClose}
            className="w-full py-4 bg-white/5 hover:bg-white/10 text-white font-semibold rounded-2xl transition-all border border-white/5 active:scale-[0.98]"
          >
            Dismiss
          </button>
        </div>

        <button 
          onClick={onClose}
          className="absolute top-4 right-4 text-neutral-600 hover:text-white transition-colors"
        >
          <CloseIcon size={14} />
        </button>
      </motion.div>
    </div>
  );
}

// ─── Avatar Picker Modal ─────────────────────────────────────────────────────
function AvatarPickerModal({
  currentId,
  isOpen,
  onClose,
  onSelect
}: {
  currentId: string;
  isOpen: boolean;
  onClose: () => void;
  onSelect: (id: string) => void;
}) {
  const avatars = useAvatars();
  const [tempId, setTempId] = useState(currentId);
  if (!isOpen) return null;
  return (
    <div className="fixed inset-0 z-[100] flex items-center justify-center p-4 md:p-6">
      <div className="absolute inset-0 bg-black/90 backdrop-blur-md" onClick={onClose} />
      <motion.div 
        initial={{ opacity: 0, scale: 0.95, y: 20 }}
        animate={{ opacity: 1, scale: 1, y: 0 }}
        className="w-full max-w-5xl h-[85vh] bg-surface-secondary rounded-3xl border border-white/5 shadow-[0_0_50px_rgba(0,0,0,0.8)] overflow-hidden flex flex-col relative"
      >
        <div className="px-6 py-5 border-b border-white/5 flex items-center justify-between bg-surface-card/50">
          <div>
            <h2 className="text-xl font-bold text-white">Add Avatar</h2>
            <p className="text-[#6b7280] text-xs mt-0.5">Select an identity for your interaction</p>
          </div>
          <button onClick={onClose} className="p-2 hover:bg-white/5 rounded-full text-[#6b7280] hover:text-white transition-colors">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <line x1="18" x2="6" y1="6" y2="18"/><line x1="6" x2="18" y1="6" y2="18"/>
            </svg>
          </button>
        </div>
        <div className="flex-1 overflow-y-auto p-6 custom-scrollbar">
          <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-6">
            {avatars.map((avatar) => (
              <button
                key={avatar.id}
                onClick={() => setTempId(avatar.id)}
                className={`group relative rounded-2xl transition-all duration-200 overflow-hidden ${
                  tempId === avatar.id ? "ring-2 ring-[#00E3AA] shadow-[0_0_30px_rgba(0,227,170,0.2)]" : "border border-white/5 hover:border-white/10"
                }`}
              >
                <div className="relative w-full aspect-video rounded-xl overflow-hidden border border-white/5 shadow-inner">
                  <img src={avatar.image} alt={avatar.name} className={`w-full h-full object-cover transition-transform duration-300 ${tempId === avatar.id ? "scale-105" : "group-hover:scale-105"}`} loading="lazy" />
                  <div className="absolute top-2 left-2"><span className="px-2.5 py-1 rounded-full bg-black/50 backdrop-blur-md text-[11px] text-white font-semibold border border-white/10 shadow-lg">{avatar.name}</span></div>
                  <div className="absolute top-2 right-2"><span className="px-2.5 py-1 rounded-full bg-black/50 backdrop-blur-md text-[11px] text-white/80 font-medium border border-white/10 shadow-lg">Huma-2</span></div>
                  <div className="absolute bottom-3 left-3"><span className="text-[10px] text-white font-bold uppercase tracking-wider">PRO</span></div>
                  <div className="absolute bottom-3 right-3"><span className="text-[10px] text-white/70 font-mono">id:{avatar.id}</span></div>
                  {tempId === avatar.id && (
                    <div className="absolute inset-0 bg-brand/10 flex items-center justify-center backdrop-blur-[1px]">
                      <div className="w-10 h-10 rounded-full bg-brand text-black flex items-center justify-center shadow-xl ring-4 ring-brand/20">
                        <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="4" strokeLinecap="round" strokeLinejoin="round"><polyline points="20 6 9 17 4 12"/></svg>
                      </div>
                    </div>
                  )}
                </div>
              </button>
            ))}
          </div>
        </div>
        <div className="px-6 py-5 border-t border-white/5 flex items-center justify-between bg-surface-card/50">
          <button onClick={onClose} className="px-6 py-2.5 text-sm font-semibold text-[#9ca3af] hover:text-white transition-colors">Cancel</button>
          <button onClick={() => { onSelect(tempId); onClose(); }} className="px-8 py-2.5 rounded-xl bg-brand text-black font-bold text-sm tracking-wide shadow-lg hover:bg-[#00c994] transition-all">Save Selection</button>
        </div>
      </motion.div>
    </div>
  );
}

// ─── Avatar Gallery ──────────────────────────────────────────────────────────
function AvatarGallery() {
  const avatars = useAvatars();
  return (
    <div className="absolute inset-0 overflow-y-auto p-6 md:p-10 custom-scrollbar bg-canvas z-10">
      <div className="max-w-6xl mx-auto pb-20">
        <header className="mb-10 text-center md:text-left">
          <h1 className="text-3xl font-bold text-white tracking-tight flex items-center gap-3">
            <SmileIcon size={32} className="text-brand" />
            Stock Avatars
          </h1>
          <p className="text-[#6b7280] mt-2">Design your AI companions with advanced customization</p>
        </header>
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6 md:gap-8">
          {avatars.map((avatar) => (
            <div key={avatar.id} className="group relative rounded-2xl transition-all duration-200 overflow-hidden border border-white/5 hover:border-white/10">
              <div className="relative w-full aspect-video">
                <img src={avatar.image} alt={avatar.name} className="w-full h-full object-cover transition-transform duration-400 group-hover:scale-105" loading="lazy" />
                <div className="absolute top-2 left-2"><span className="px-2.5 py-1 rounded-full bg-black/50 backdrop-blur-md text-[11px] text-white font-semibold border border-white/10 shadow-lg">{avatar.name}</span></div>
                <div className="absolute top-2 right-2"><span className="px-2.5 py-1 rounded-full bg-black/50 backdrop-blur-md text-[11px] text-white/80 font-medium border border-white/10 shadow-lg">Huma-2</span></div>
                <div className="absolute bottom-3 left-3"><span className="text-[10px] text-white font-bold uppercase tracking-wider">PRO</span></div>
                <div className="absolute bottom-3 right-3"><span className="text-[10px] text-white/70 font-mono">id:{avatar.id}</span></div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

// ─── Transcript Synchronizer ─────────────────────────────────────────────────
function TranscriptSynchronizer({ 
  transcriptRef,
  startTimeRef
}: { 
  transcriptRef: React.MutableRefObject<any[]>,
  startTimeRef: React.MutableRefObject<number | null>
}) {
  const combinedTranscriptions = useCombinedTranscriptions();
  
  useEffect(() => {
    if (combinedTranscriptions.length > 0) {
      const startTime = startTimeRef.current || 0;
      
      // Only include segments that started AFTER the session began
      const filtered = combinedTranscriptions.filter(s => s.firstReceivedTime >= startTime);

      if (filtered.length > 0) {
        transcriptRef.current = filtered.map(s => ({
          text: s.text,
          isAgent: s.role === "assistant",
          timestamp: new Date(s.firstReceivedTime).toISOString(),
          participant: s.role === "assistant" ? "Agent" : "User"
        }));
      }
    }
  }, [combinedTranscriptions]);
  
  return null;
}

// ─── Active Voice Assistant View ─────────────────────────────────────────────
function ActiveVoiceAssistantView({ onConnectButtonClicked }: { onConnectButtonClicked: () => void }) {
  const { state: agentState, videoTrack } = useVoiceAssistant();
  const [isChatVisible, setIsChatVisible] = useState(false);
  const [chatWidth, setChatWidth] = useState(450);
  const [isDragging, setIsDragging] = useState(false);

  const MIN_WIDTH = 300;
  const MAX_WIDTH = 800;

  const handlePointerDown = (e: React.PointerEvent) => {
    e.preventDefault();
    setIsDragging(true);
    const handlePointerMove = (moveEvent: PointerEvent) => {
      const newWidth = document.documentElement.clientWidth - moveEvent.clientX;
      setChatWidth(Math.min(Math.max(newWidth, MIN_WIDTH), MAX_WIDTH));
    };
    const handlePointerUp = () => {
      setIsDragging(false);
      document.removeEventListener("pointermove", handlePointerMove);
      document.removeEventListener("pointerup", handlePointerUp);
    };
    document.addEventListener("pointermove", handlePointerMove);
    document.addEventListener("pointerup", handlePointerUp);
  };

  if (agentState === "disconnected") return null;

  const isAgentInteractive = ["listening", "thinking", "speaking", "idle"].includes(agentState) && !!videoTrack;

  return (
    <motion.div 
      initial={{ opacity: 0 }} 
      animate={{ opacity: 1 }} 
      exit={{ opacity: 0 }} 
      className="absolute inset-0 flex h-full w-full bg-canvas overflow-hidden z-20"
    >
      <main className="flex-1 h-full flex flex-col relative bg-[#000000]">
        <div className="flex-1 flex items-center justify-center p-12">
          {/* Only render visualizer when truly interactive, to prevent premature waving */}
          {isAgentInteractive && <AgentVisualizer />}
        </div>
        <div className="absolute bottom-12 left-0 right-0 flex justify-center">
          <ControlBar onConnectButtonClicked={onConnectButtonClicked} isChatVisible={isChatVisible} setIsChatVisible={setIsChatVisible} />
        </div>
      </main>
      <motion.aside
        initial={false}
        animate={{ width: isChatVisible ? chatWidth : 0, opacity: isChatVisible ? 1 : 0 }}
        transition={{ duration: isDragging ? 0 : 0.3, ease: "easeInOut" }}
        className="relative min-w-0 h-full border-l border-white/5 bg-black/10 backdrop-blur-md overflow-hidden flex-shrink-0"
      >
        {isChatVisible && (
          <div onPointerDown={handlePointerDown} className="absolute left-0 top-0 bottom-0 w-2 cursor-col-resize z-10 hover:bg-white/10 active:bg-white/20 transition-colors" />
        )}
        <div style={{ width: chatWidth }} className="h-full">
          <TranscriptionView />
        </div>
      </motion.aside>
      <NoAgentNotification state={agentState} />
    </motion.div>
  );
}

// ─── Voice Assistant (manages disconnected/connected states) ─────────────────
function SimpleVoiceAssistant({
  onConnectButtonClicked,
  config,
  setConfig,
  onOpenPicker,
  onSaveAsBot,
  isSavingBot,
  isEditing,
  onCancelEdit,
  bots = [],
  showConnectButton = true,
  isConnecting: isConnectingProp = false,
}: {
  onConnectButtonClicked: () => void;
  config: typeof DEFAULTS;
  setConfig: (c: typeof DEFAULTS) => void;
  onOpenPicker: () => void;
  onSaveAsBot?: () => void;
  isSavingBot?: boolean;
  isEditing?: boolean;
  onCancelEdit?: () => void;
  bots?: AgentBot[];
  showConnectButton?: boolean;
  isConnecting?: boolean;
}) {
  const { state: agentState } = useVoiceAssistant();
  const [internalIsConnecting, setInternalIsConnecting] = useState(false);

  const isConnecting = internalIsConnecting || isConnectingProp;

  const handleConnect = async () => {
    setInternalIsConnecting(true);
    try {
      await onConnectButtonClicked();
    } finally {
      setInternalIsConnecting(false);
    }
  };

  return (
    <div className="h-screen w-full bg-canvas">
      <AnimatePresence mode="wait">
        {!["listening", "thinking", "speaking", "idle"].includes(agentState) ? (
          <SessionConfigForm
            key="config"
            config={config}
            setConfig={setConfig}
            onConnect={handleConnect}
            isConnecting={isConnecting}
            onOpenPicker={onOpenPicker}
            onSaveAsBot={onSaveAsBot}
            isSavingBot={isSavingBot}
            isEditing={isEditing}
            onCancelEdit={onCancelEdit}
            bots={bots}
            showConnectButton={showConnectButton}
          />
        ) : (
          <ActiveVoiceAssistantView 
            key="active" 
            onConnectButtonClicked={onConnectButtonClicked} 
          />
        )}
      </AnimatePresence>
    </div>
  );
}

// ─── Agent Visualizer ────────────────────────────────────────────────────────
function AgentVisualizer() {
  const { state: agentState, videoTrack, audioTrack } = useVoiceAssistant();
  if (videoTrack) {
    return (
      <div className="w-full max-w-5xl mx-auto aspect-video rounded-2xl overflow-hidden border border-white/10 shadow-2xl bg-black/50 transition-all duration-300">
        <VideoTrack trackRef={videoTrack} className="w-full h-full object-cover" />
      </div>
    );
  }
  return (
    <div className="h-[300px] w-full max-w-2xl mx-auto flex items-center justify-center">
      <BarVisualizer state={agentState} barCount={5} trackRef={audioTrack} className="agent-visualizer" options={{ minHeight: 24 }} />
    </div>
  );
}

// ─── Control Bar ─────────────────────────────────────────────────────────────

function ControlBar(props: {
  onConnectButtonClicked: () => void;
  isChatVisible: boolean;
  setIsChatVisible: (v: boolean) => void;
}) {
  const { state: agentState } = useVoiceAssistant();
  const room = useRoomContext();
  const [isMicEnabled, setIsMicEnabled] = useState(true);
  const toggleMic = async () => {
    const enabled = !isMicEnabled;
    setIsMicEnabled(enabled);
    await room.localParticipant.setMicrophoneEnabled(enabled);
  };
  if (agentState === "disconnected" || agentState === "connecting") return null;
  return (
    <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} exit={{ opacity: 0, y: 20 }} transition={{ duration: 0.4, ease: [0.09, 1.04, 0.245, 1.055] }} className="flex items-center gap-4">
      <div className="control-pill">
        <button onClick={toggleMic} className="control-button-white">{isMicEnabled ? <MicIcon /> : <MicOffIcon />}</button>
        <div className="control-dropdown-part"><ChevronDownIcon /></div>
      </div>
      <button onClick={() => props.setIsChatVisible(!props.isChatVisible)} className={`control-circle ${props.isChatVisible ? "active" : ""}`}><MessageIcon /></button>
      <DisconnectButton className="disconnect-circle"><CrossIcon /></DisconnectButton>
    </motion.div>
  );
}

function onDeviceFailure(error: Error) {
  console.error(error);
  alert("Error acquiring microphone permissions. Please grant the necessary permissions and reload.");
}

// ─── Bot Library View ────────────────────────────────────────────────────────
function BotLibraryView({
  bots,
  profileId,
  onRefresh,
  onSelectBot,
  onEditBot,
  botHealth,
  isConnecting
}: {
  bots: AgentBot[],
  profileId: string | null,
  onRefresh: () => void,
  onSelectBot: (bot: AgentBot) => void,
  onEditBot: (bot: AgentBot) => void,
  botHealth: Record<string, 'healthy' | 'unhealthy' | 'checking' | 'unknown'>,
  isConnecting?: boolean
}) {
  const avatars = useAvatars();
  const [isDeleting, setIsDeleting] = useState<string | null>(null);
  const [copiedEmail, setCopiedEmail] = useState<string | null>(null);

  const handleDelete = async (id: string, e: React.MouseEvent) => {
    e.stopPropagation();
    if (!confirm("Are you sure you want to delete this agent?")) return;
    setIsDeleting(id);
    const apiKey = localStorage.getItem("defaultApiKey") ?? "";
    await deleteAgent(apiKey, id);
    onRefresh();
    setIsDeleting(null);
  };

  const containerVariants = {
    hidden: { opacity: 0 },
    visible: {
      opacity: 1,
      transition: {
        staggerChildren: 0.1,
        delayChildren: 0.2
      }
    }
  };

  const cardVariants = {
    hidden: { opacity: 0, y: 30, scale: 0.95 },
    visible: { 
      opacity: 1, 
      y: 0, 
      scale: 1,
      transition: { type: "spring", stiffness: 100, damping: 20 }
    }
  };

  return (
    <div className="absolute inset-0 overflow-y-auto p-6 md:p-12 custom-scrollbar bg-canvas z-10">
      <div className="max-w-7xl mx-auto pb-24">
        <motion.header 
          initial={{ opacity: 0, y: -20 }}
          animate={{ opacity: 1, y: 0 }}
          className="mb-14 flex items-end justify-between"
        >
          <div className="space-y-1">
            <div className="flex items-center gap-3 mb-2">
              <div className="w-10 h-10 rounded-2xl bg-brand/10 flex items-center justify-center border border-brand/20 shadow-[0_0_20px_rgba(0,227,170,0.1)]">
                <LibraryIcon size={22} className="text-brand" />
              </div>
              <h1 className="text-3xl font-bold text-white tracking-tight font-outfit">
                Agent <span className="text-brand">Library</span>
              </h1>
            </div>
            <p className="text-neutral-500 font-medium text-sm tracking-wide ml-1 font-outfit">
              Select and deploy your personalized AI agents to any meeting.
            </p>
          </div>
          
          <button 
            onClick={onRefresh} 
            className="group p-3 rounded-2xl bg-white/[0.03] hover:bg-brand/10 text-neutral-500 hover:text-brand transition-all border border-white/5 hover:border-brand/30 shadow-xl"
          >
            <RefreshCwIcon size={20} className="group-hover:rotate-180 transition-transform duration-500" />
          </button>
        </motion.header>

        {bots.length === 0 ? (
          <motion.div 
            initial={{ opacity: 0, scale: 0.9 }}
            animate={{ opacity: 1, scale: 1 }}
            className="flex flex-col items-center justify-center py-32 border border-white/5 rounded-[2.5rem] bg-gradient-to-b from-white/[0.03] to-transparent backdrop-blur-3xl shadow-2xl"
          >
            <div className="w-24 h-24 rounded-full bg-white/5 flex items-center justify-center text-neutral-700 mb-8 border border-white/5 relative overflow-hidden group">
              <div className="absolute inset-0 bg-gradient-to-tr from-brand/10 to-transparent opacity-0 group-hover:opacity-100 transition-opacity" />
              <LibraryIcon size={40} />
            </div>
            <h3 className="text-2xl font-bold text-white tracking-tight">Your vault is empty</h3>
            <p className="text-neutral-500 text-[15px] mt-3 max-w-sm text-center font-medium">
              Create and save your first AI companion from the <span className="text-brand">&quot;Add Agent&quot;</span> lab to see them listed here.
            </p>
          </motion.div>
        ) : (
          <motion.div 
            key={bots.length}
            variants={containerVariants}
            initial="hidden"
            animate="visible"
            className="grid grid-cols-1 min-[500px]:grid-cols-2 min-[1100px]:grid-cols-3 xl:grid-cols-3 gap-6 md:gap-8"
          >
            {bots.map((bot) => {
              const botAvatarId = getBotAvatarId(bot);
              const avatar = avatars.find(a => a.id === botAvatarId);
              return (
                <motion.div 
                  key={bot.id} 
                  variants={cardVariants}
                  whileHover={{ y: -12, scale: 1.02, transition: { type: "spring", stiffness: 260, damping: 25 } }}
                  onClick={() => onSelectBot(bot)} 
                  className="group relative rounded-[2rem] bg-surface-secondary/80 backdrop-blur-xl border border-white/5 hover:border-brand/40 transition-[border-color,box-shadow,background-color] duration-300 overflow-hidden cursor-pointer flex flex-col shadow-2xl hover:shadow-brand/20"
                >
                  {/* Decorative background glow */}
                  <div className="absolute top-0 right-0 w-24 h-24 bg-brand/5 rounded-full blur-[60px] pointer-events-none group-hover:bg-brand/20 transition-all duration-300" />
                  
                  <div className="relative aspect-[16/10] w-full overflow-hidden bg-black/40">
                    {avatar ? (
                      <img 
                        src={avatar.image}
                        alt={bot.agent_name}
                        className="w-full h-full object-cover transition-transform duration-500 scale-105 group-hover:scale-110 opacity-100 grayscale-0" 
                      />
                    ) : (
                      <div className="w-full h-full flex items-center justify-center bg-white/5 text-neutral-700">
                        <UserIcon size={48} />
                      </div>
                    )}
                    
                    {/* Health Status Indicator Badge */}
                    <div className="absolute top-4 left-4 z-20">
                      {(() => {
                        const healthKey = `${(bot.config?.openclaw_url ?? "").replace(/\/$/, "")}|${bot.config?.gateway_token ?? ""}`;
                        const status = botHealth[healthKey] || 'unknown';
                        
                        return (
                          <div className={`
                            flex items-center gap-2 px-2.5 py-1 rounded-full backdrop-blur-md border 
                            ${status === 'healthy' ? 'bg-green-500 border-green-500/30' : 
                              status === 'unhealthy' ? 'bg-red-500 border-red-500/30 ' : 
                              status === 'checking' ? 'bg-yellow-500 border-yellow-500/30' : 
                              'bg-neutral-500 border-neutral-500 '}
                          `}>
                            {status === 'checking' ? (
                              <div className="w-2 h-2 rounded-full bg-current animate-pulse shadow-[0_0_8px_rgba(234,179,8,0.5)]" />
                            ) : (
                              <div className={`w-2 h-2 rounded-full bg-current ${status === 'healthy' ? 'shadow-[0_0_8px_rgba(34,197,94,0.5)]' : status === 'unhealthy' ? 'shadow-[0_0_8px_rgba(239,68,68,0.5)]' : ''}`} />
                            )}
                            <span className="text-[10px] font-bold uppercase tracking-wider">
                              {status === 'healthy' ? 'Active' : 
                               status === 'unhealthy' ? 'Offline' : 
                               status === 'checking' ? 'Pinging' : 'Unknown'}
                            </span>
                          </div>
                        );
                      })()}
                    </div>
                    
                    <div className="absolute inset-0 bg-gradient-to-t from-surface-secondary via-surface-secondary/20 to-transparent opacity-90" />
                    
                    {/* Floating Controls */}
                    <div className="absolute top-4 right-4 flex gap-2 translate-y-2 opacity-0 group-hover:translate-y-0 group-hover:opacity-100 transition-all duration-200">
                      <button 
                        onClick={(e) => { e.stopPropagation(); onEditBot(bot); }} 
                        className="p-2.5 rounded-xl bg-black/40 backdrop-blur-xl border border-white/10 text-white/50 hover:text-white hover:bg-brand/20 hover:border-brand/40 transition-all"
                        title="Edit Configuration"
                      >
                        <SettingsIcon size={16} />
                      </button>
                      <button 
                        onClick={(e) => handleDelete(bot.id, e)} 
                        className="p-2.5 rounded-xl bg-black/40 backdrop-blur-xl border border-white/10 text-white/50 hover:text-red-500 hover:bg-red-500/10 hover:border-red-500/40 transition-all"
                        title="Delete Agent"
                      >
                        {isDeleting === bot.id ? <RefreshCwIcon size={16} className="animate-spin" /> : <TrashIcon size={16} />}
                      </button>
                    </div>

                    {/* Agent Name Badge (Bottom Left) */}
                    <div className="absolute bottom-4 left-6">
                      <h3 className="text-xl font-bold text-white tracking-tight group-hover:text-brand transition-colors duration-200 font-outfit">
                        {bot.agent_name || "Unnamed Agent"}
                      </h3>
                      <div className="flex items-center gap-1.5 mt-1">
                        <div className="w-1.5 h-1.5 rounded-full bg-brand shadow-[0_0_8px_rgba(0,227,170,0.6)]" />
                        <span className="text-[9px] text-neutral-400 font-semibold uppercase tracking-wider font-outfit">Configured Identity</span>
                      </div>
                    </div>
                  </div>

                  <div className="p-6 pt-2 relative">
                    <div className="space-y-3">
                      {/* Identity Section */}
                      <div className="space-y-4 px-1 pt-4">
                        {/* URL Source */}
                        <div className="flex items-center gap-3 px-1 text-[12px] text-neutral-500">
                          <span className="text-neutral-700"><LinkIcon size={14} /></span>
                          <span className="truncate italic font-medium">{bot.config?.openclaw_url ?? "—"}</span>
                        </div>

                        {/* Avatar Info */}
                        <div className="flex items-center gap-3">
                          <div className="w-6 h-6 rounded-lg bg-neutral-800 flex items-center justify-center text-neutral-400">
                            <UserIcon size={14} />
                          </div>
                          <div className="flex flex-col">
                            <span className="text-[9px] text-neutral-600 font-bold uppercase tracking-tighter leading-none font-outfit">Avatar Id</span>
                            <span className="text-[12px] text-neutral-300 font-jetbrains-mono font-medium truncate">{botAvatarId || "—"}</span>
                          </div>
                        </div>

                        {/* Email Info */}
                        {bot.email && (
                          <div className="flex items-center justify-between group/email py-2 px-3 rounded-xl bg-brand/5 border border-brand/10 hover:border-brand/30 transition-all shadow-inner relative">
                            <div className="flex items-center gap-3 min-w-0">
                              <div className="w-6 h-6 rounded-lg bg-brand/10 flex items-center justify-center text-brand shrink-0">
                                <MailIcon size={12} />
                              </div>
                              <span className="text-[12px] text-brand font-bold font-jetbrains-mono truncate lowercase tracking-tight">
                                {bot.email}
                              </span>
                            </div>
                            <button
                              onClick={(e) => {
                                e.stopPropagation();
                                navigator.clipboard.writeText(bot.email);
                                setCopiedEmail(bot.email);
                                setTimeout(() => setCopiedEmail(null), 2000);
                              }}
                              className={`p-1.5 rounded-lg transition-all ${
                                copiedEmail === bot.email
                                  ? "text-brand bg-brand/20 opacity-100"
                                  : "text-neutral-500 hover:text-white transition-all opacity-0 group-hover/email:opacity-100"
                              }`}
                            >
                              {copiedEmail === bot.email ? <CheckIcon size={14} /> : <span className="rotate-45 block"><LinkIcon size={14} /></span>}
                            </button>
                          </div>
                        )}
                      </div>
                    </div>

                    <div className="mt-8 flex flex-wrap items-end justify-between gap-y-4 gap-x-2 px-1">
                      <div className="flex flex-col">
                        <span className="text-[9px] text-neutral-600 font-bold uppercase tracking-tighter font-outfit">Creation Date</span>
                        <div className="flex items-center gap-1.5 text-[11px] text-neutral-400 font-bold font-jetbrains-mono whitespace-nowrap">
                          <span className="text-neutral-700"><ClockIcon size={12} /></span>
                          <span>{bot.created_at ? new Date(bot.created_at).toLocaleDateString() : "—"}</span>
                        </div>
                      </div>

                      <div className="flex items-center gap-2">
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            const newConfig = {
                              openclawUrl: bot.config.openclaw_url,
                              gatewayToken: bot.config.gateway_token,
                              sessionKey: "",
                              avatarId: botAvatarId,
                              botName: bot.agent_name,
                            };
                            (window as any).openRecallWithConfig?.(newConfig);
                          }}
                          className="p-3 rounded-2xl bg-white/5 hover:bg-white/10 text-neutral-500 hover:text-white border border-white/5 shadow-lg transition-all group/recall"
                          title="Generate Automated URL"
                        >
                          <span className="group-hover/recall:scale-110 transition-transform block"><LinkIcon size={16} /></span>
                        </button>
                        
                        <button 
                          disabled={isConnecting}
                          onClick={(e) => {
                            e.stopPropagation();
                            onSelectBot(bot);
                          }}
                          className={`h-10 px-5 rounded-xl text-black text-[12px] font-bold uppercase tracking-widest transition-all transform hover:scale-[1.02] active:scale-95 shadow-[0_4px_12px_rgba(0,227,170,0.2)] flex items-center gap-2 shrink-0 whitespace-nowrap ${
                            isConnecting 
                              ? "bg-neutral-800 text-neutral-500 cursor-not-allowed opacity-50 shadow-none scale-100" 
                              : "bg-brand hover:bg-[#00ffd0]"
                          }`}
                        >
                          {isConnecting ? (
                            <RefreshCwIcon size={14} className="animate-spin" />
                          ) : (
                            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="4" strokeLinecap="round" strokeLinejoin="round">
                              <polygon points="5 3 19 12 5 21 5 3"/>
                            </svg>
                          )}
                          {isConnecting ? "Connecting..." : "Connect"}
                        </button>
                      </div>
                    </div>
                  </div>
                </motion.div>
              );
            })}
          </motion.div>
        )}
      </div>
    </div>
  );
}

// ─── Conversations List View ────────────────────────────────────────────────
function ConversationsListView({
  isLoading,
  conversations,
  onSelect
}: {
  isLoading: boolean;
  conversations: any[];
  onSelect: (conv: any) => void;
}) {
  const avatars = useAvatars();
  return (
    <div className="absolute inset-0 overflow-y-auto p-6 md:p-10 custom-scrollbar bg-canvas z-10">
      <div className="max-w-6xl mx-auto pb-20">
        <header className="mb-10">
          <h1 className="text-3xl font-bold text-white tracking-tight flex items-center gap-3">
            <RefreshCwIcon size={32} className="text-brand" />
            Conversations
          </h1>
          <p className="text-[#6b7280] mt-2 text-sm">Review past interactions and transcripts</p>
        </header>

        {isLoading ? (
          <div className="flex items-center justify-center py-20">
            <RefreshCwIcon className="animate-spin text-brand" size={32} />
          </div>
        ) : conversations.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-20 border-2 border-dashed border-white/5 rounded-3xl bg-white/[0.02]">
            <div className="w-16 h-16 rounded-full bg-white/5 flex items-center justify-center text-[#4b5563] mb-4"><MessageIcon size={32} /></div>
            <h3 className="text-lg font-semibold text-white">No Conversations Found</h3>
            <p className="text-[#6b7280] text-[13px] mt-1 max-w-xs text-center">Your interaction history will appear here after your first call.</p>
          </div>
        ) : (
          <div className="bg-surface border border-white/5 rounded-2xl overflow-hidden">
            <table className="w-full text-left border-collapse">
              <thead>
                <tr className="border-b border-white/5 bg-white/5">
                  <th className="px-6 py-4 text-[12px] font-bold uppercase tracking-wider text-[#6b7280]">Status</th>
                  <th className="px-6 py-4 text-[12px] font-bold uppercase tracking-wider text-[#6b7280]">Agent Detail</th>
                  <th className="px-6 py-4 text-[12px] font-bold uppercase tracking-wider text-[#6b7280]">Duration</th>
                  <th className="px-6 py-4 text-[12px] font-bold uppercase tracking-wider text-[#6b7280]">Date/Time</th>
                  <th className="px-6 py-4 text-[12px] font-bold uppercase tracking-wider text-[#6b7280]">Action</th>
                </tr>
              </thead>
              <tbody>
                {conversations.map((conv) => {
                  const getStatusStyles = (status: string) => {
                    const normalized = status?.toLowerCase();
                    if (normalized === "completed") return "bg-green-500/10 text-green-500 border-green-500/20";
                    if (normalized === "terminated") return "bg-red-500/10 text-red-500 border-red-500/20";
                    if (normalized === "failed" || normalized === "interrupted") return "bg-yellow-500/10 text-yellow-500 border-yellow-500/20";
                    return "bg-gray-500/10 text-gray-500 border-gray-500/20";
                  };
                  // Map API response fields — agentId links back to the bot in the library
                  const matchedBot = avatars.find(a => a.id === (conv.bot_avatar ?? conv.agentId));
                  const displayAvatar = matchedBot ?? avatars[0];
                  const displayName = conv.bot_name || conv.userName || conv.userId || "Unknown";
                  const displayId = conv.session_key || conv.agentId || conv.id;
                  const displayDate = conv.created_at ? new Date(conv.created_at) : null;
                  return (
                    <tr key={conv.id} className="border-b border-white/5 hover:bg-white/[0.02] transition-colors cursor-pointer group" onClick={() => onSelect(conv)}>
                      <td className="px-6 py-4">
                        <span className={`px-2 py-1 rounded-full text-[10px] font-bold uppercase border ${getStatusStyles(conv.status || "Completed")}`}>
                          {conv.status || "Completed"}
                        </span>
                      </td>
                      <td className="px-6 py-4">
                        <div className="flex items-center gap-3">
                          <div className="w-8 h-8 rounded-lg overflow-hidden bg-white/10 flex flex-shrink-0 items-center justify-center text-[#9ca3af]">
                            {displayAvatar ? (
                              <img src={displayAvatar.image} className="w-full h-full object-cover" alt={displayAvatar.name} />
                            ) : (
                              <UserIcon size={16} />
                            )}
                          </div>
                          <div className="flex flex-col min-w-0">
                            <span className="text-[#6b7280] text-[10px] font-bold uppercase tracking-widest mb-0.5 opacity-60">Video Companion</span>
                            <span className="text-white text-[14px] font-bold truncate leading-tight">{displayName}</span>
                            <span className="text-[#3a3a3a] text-[9px] font-mono truncate mt-0.5">ID: {displayId}</span>
                          </div>
                        </div>
                      </td>
                      <td className="px-6 py-4">
                        <span className="text-[#9ca3af] text-[13px]">{conv.duration ? `${conv.duration}s` : '—'}</span>
                      </td>
                      <td className="px-6 py-4">
                        <div className="flex flex-col">
                          <span className="text-white text-[13px]">{displayDate ? displayDate.toLocaleDateString() : '—'}</span>
                          <span className="text-[#3a3a3a] text-[11px]">{displayDate ? displayDate.toLocaleTimeString() : ''}</span>
                        </div>
                      </td>
                      <td className="px-6 py-4">
                        <button className="px-4 py-1.5 rounded-lg bg-white/5 hover:bg-brand/20 hover:text-brand transition-all text-[12px] font-semibold text-white/70">View History</button>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}

// ─── Conversation Detail View ───────────────────────────────────────────────
function ConversationDetailView({
  conversation,
  onBack
}: {
  conversation: any;
  onBack: () => void;
}) {
  return (
    <div className="absolute inset-0 overflow-y-auto p-6 md:p-10 custom-scrollbar bg-canvas z-10">
      <div className="max-w-4xl mx-auto pb-20">
        <header className="mb-10 flex items-center justify-between">
          <div className="flex items-center gap-4">
            <button onClick={onBack} className="p-2 rounded-xl bg-white/5 hover:bg-white/10 text-white/70 transition-all border border-white/5">
              <ChevronDownIcon className="rotate-90" />
            </button>
            <div>
              <h1 className="text-2xl font-bold text-white tracking-tight">Conversation with {conversation.bot_name || conversation.userName || "Agent"}</h1>
              <p className="text-[#6b7280] text-sm mt-1">{conversation.created_at ? new Date(conversation.created_at).toLocaleString() : "—"}{conversation.duration ? ` • ${conversation.duration}s` : ""}</p>
            </div>
          </div>
          <span className={`px-3 py-1 rounded-full text-[12px] font-bold uppercase border ${
            conversation.status?.toLowerCase() === "completed" ? "bg-green-500/10 text-green-500 border-green-500/20" :
            conversation.status?.toLowerCase() === "terminated" ? "bg-red-500/10 text-red-500 border-red-500/20" :
            "bg-yellow-500/10 text-yellow-500 border-yellow-500/20"
          }`}>
            {conversation.status || "Completed"}
          </span>
        </header>

        <div className="space-y-6">
          {Array.isArray(conversation.transcript) && conversation.transcript.length > 0 ? (
            conversation.transcript.map((msg: any, idx: number) => {
              const role = msg.role ?? (msg.isAgent ? "assistant" : "user");
              const isAgent = typeof msg.isAgent === "boolean" ? msg.isAgent : role !== "user";
              const label = role === "assistant" ? "Agent" : role === "tool" ? "Tool" : "User";
              const text = msg.text ?? msg.content ?? msg.output ?? msg.arguments ?? "";
              if (!text) return null;
              const ts = msg.timestamp ?? (msg.message_timestamp ? new Date(msg.message_timestamp * 1000).toISOString() : "");
              const timeLabel = ts ? new Date(ts).toLocaleTimeString() : "";

              return (
                <div key={idx} className={`flex ${isAgent ? "justify-start" : "justify-end"}`}>
                  <div className={`max-w-[80%] rounded-2xl p-4 ${isAgent ? "bg-white/5 border border-white/10 text-white" : "bg-brand/10 border border-brand/20 text-white"}`}>
                    <div className="flex items-center gap-2 mb-2">
                      <span className="text-[10px] font-bold uppercase tracking-wider opacity-50">{label}</span>
                      {timeLabel ? <span className="text-[10px] opacity-30">{timeLabel}</span> : null}
                    </div>
                    <p className="text-[15px] leading-relaxed">{text}</p>
                  </div>
                </div>
              );
            })
          ) : (
            <div className="text-center py-20 text-neutral-500">No transcript available for this session.</div>
          )}
        </div>
      </div>
    </div>
  );
}

// ─── Direct Call Dashboard ───────────────────────────────────────────────────
const AIGlowingOrb = () => {
  return (
    <div className="relative w-40 h-40 mb-10 flex items-center justify-center">
      {/* Massive subtle outer pulse */}
      <motion.div
        className="absolute w-full h-full rounded-full bg-brand/10 blur-[40px]"
        animate={{ scale: [1, 2, 1], opacity: [0.3, 0.6, 0.3] }}
        transition={{ duration: 4, repeat: Infinity, ease: "easeInOut" }}
      />
      {/* Secondary breathing ring */}
      <motion.div
        className="absolute w-32 h-32 rounded-full border border-brand/30"
        animate={{ scale: [1, 1.4, 1], opacity: [0.8, 0, 0.8] }}
        transition={{ duration: 3, repeat: Infinity, ease: "easeInOut", delay: 0.5 }}
      />
      {/* Rotational aura */}
      <motion.div
        className="absolute w-28 h-28 rounded-full bg-gradient-to-tr from-brand/40 to-transparent blur-xl"
        animate={{ rotate: [0, 360] }}
        transition={{ duration: 5, repeat: Infinity, ease: "linear" }}
      />
      {/* Inner pulsing core */}
      <motion.div
        className="absolute w-20 h-20 rounded-full bg-gradient-to-br from-white via-brand to-[#00A080] shadow-[0_0_50px_rgba(0,227,170,1)] flex items-center justify-center"
        animate={{ scale: [1, 1.15, 1] }}
        transition={{ duration: 1.5, repeat: Infinity, ease: "easeInOut" }}
      >
        <div className="w-full h-full rounded-full bg-white/20 blur-sm" />
        <div className="absolute w-12 h-12 rounded-full bg-white/50 blur-md mix-blend-overlay" />
      </motion.div>
    </div>
  );
};

function DirectCallDashboard({
  config,
  onStartCall,
  onBack,
  autoStart = false,
  isValidating = false,
}: {
  config: typeof DEFAULTS;
  onStartCall: () => void;
  onBack: () => void;
  autoStart?: boolean;
  isValidating?: boolean;
}) {
  const { state: agentState, audioTrack, videoTrack } = useVoiceAssistant();
  const [isConnecting, setIsConnecting] = useState(autoStart || false);
  const avatars = useAvatars();
  const selectedAvatar = avatars.find(a => a.id === config.avatarId) || avatars[0];

  const handleStartCall = async () => {
    setIsConnecting(true);
    try {
      await onStartCall();
    } catch (e) {
      const nowTimestamp = new Date().toISOString();
      console.error(`[${nowTimestamp}] ❌ handleStartCall error:`, e);
      setIsConnecting(false);
    }
  };

  useEffect(() => {
    if (autoStart) {
      handleStartCall();
    }
  }, []); // Run once on mount

  // Keep track of if we've successfully connected so we can detect a disconnection
  const hasConnectedRef = useRef(false);
  const isAgentInteractive = ["listening", "thinking", "speaking", "idle"].includes(agentState) && !!videoTrack;

  useEffect(() => {
    if (isAgentInteractive) {
      hasConnectedRef.current = true;
    } else if (agentState === "disconnected" && hasConnectedRef.current) {
      // If we were connected and then the agent disconnected, return to library
      onBack();
    }
  }, [agentState, isAgentInteractive, onBack]);

  const room = useRoomContext();
  useEffect(() => {
    const handleDisconnected = () => {
      if (isConnecting) setIsConnecting(false);
      onBack();
    };
    room.on(RoomEvent.Disconnected, handleDisconnected);
    return () => {
      room.off(RoomEvent.Disconnected, handleDisconnected);
    };
  }, [room, isConnecting, onBack]);

  // Transition to the active view once the agent is no longer pending
  if (isAgentInteractive) {
    return <ActiveVoiceAssistantView onConnectButtonClicked={onStartCall} />;
  }

  // Determine if we are in any state that should show the orb instead of the prep UI
  const isPending = isConnecting || isValidating || agentState === "connecting";

  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      className="flex flex-col items-center justify-center min-h-[80vh] w-full"
    >
      {isPending ? (
        <div className="flex flex-col items-center justify-center animate-in fade-in zoom-in duration-500">
          <AIGlowingOrb />
          <motion.h2 
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            className="text-2xl font-bold text-white mt-8 tracking-wide flex items-center"
          >
            {isValidating ? "Validating credits" : "Connecting to agent"}
            <span className="flex ml-1">
              <motion.span animate={{ opacity: [0, 1, 0] }} transition={{ duration: 1.5, repeat: Infinity }}>.</motion.span>
              <motion.span animate={{ opacity: [0, 1, 0] }} transition={{ duration: 1.5, repeat: Infinity, delay: 0.2 }}>.</motion.span>
              <motion.span animate={{ opacity: [0, 1, 0] }} transition={{ duration: 1.5, repeat: Infinity, delay: 0.4 }}>.</motion.span>
            </span>
          </motion.h2>
        </div>
      ) : (
        <div className="flex flex-col items-center space-y-12">
          <div className="w-56 h-56 rounded-full p-1.5 border-2 border-brand/30 shadow-[0_0_40px_rgba(0,227,170,0.15)] relative">
            <div className="w-full h-full rounded-full overflow-hidden relative">
              <Image 
                src={selectedAvatar.image} 
                alt={selectedAvatar.name} 
                fill 
                className="object-cover"
              />
            </div>
            <div className="absolute bottom-4 right-4 w-6 h-6 rounded-full bg-brand border-4 border-[#0A0A0A] shadow-lg animate-pulse" />
          </div>

          <button
            onClick={handleStartCall}
            className="group relative px-12 py-5 bg-brand hover:bg-[#00c994] text-black font-bold text-xl rounded-2xl transition-all duration-300 transform hover:scale-105 active:scale-95 shadow-[0_20px_40px_-12px_rgba(0,227,170,0.4)] flex items-center gap-3"
          >
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round">
              <polygon points="5 3 19 12 5 21 5 3"/>
            </svg>
            Start Call
          </button>
        </div>
      )}
    </motion.div>
  );
}

export default function Page() {
  return (
    <Suspense fallback={
      <div className="h-screen w-screen bg-surface-secondary flex items-center justify-center">
        <svg className="animate-spin text-brand" width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <circle cx="12" cy="12" r="10" opacity="0.25"/>
          <path d="M22 12a10 10 0 0 1-10 10" opacity="0.9"/>
        </svg>
      </div>
    }>
      <ClientPage />
    </Suspense>
  );
}
