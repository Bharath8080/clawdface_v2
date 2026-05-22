"use client";
import React, { useState, useEffect, useRef } from 'react';
import Image from 'next/image';
import { useRouter, usePathname } from 'next/navigation';
import { getInitials } from '@/lib/auth';
import { useUser, useStackApp } from '@stackframe/stack';
import { getLicenseDetails } from '@/app/services/pricingPaymentService';
import { fetchAvatars, type AvatarItem } from '@/app/services/avatarService';
import { clearApiKey } from '@/app/services/apiKeyService';
import { type AgentBot } from '@/app/services/agentService';

// ---- Icons ----
const BotIcon = () => <svg width="19" height="19" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><rect width="18" height="10" x="3" y="11" rx="2"/><circle cx="12" cy="5" r="2"/><path d="M12 7v4"/><line x1="8" x2="8" y1="16" y2="16"/><line x1="16" x2="16" y1="16" y2="16"/></svg>;
const LibraryIcon = () => <svg width="19" height="19" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"/><path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"/></svg>;
const UserIcon = () => <svg width="19" height="19" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><path d="M19 21v-2a4 4 0 0 0-4-4H9a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>;
const HistoryIcon = () => <svg width="19" height="19" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><path d="M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8"/><path d="M3 3v5h5"/><path d="M12 7v5l4 2"/></svg>;
const MonitorIcon = () => <svg width="19" height="19" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><rect width="20" height="14" x="2" y="3" rx="2"/><line x1="8" x2="16" y1="21" y2="21"/><line x1="12" x2="12" y1="17" y2="21"/></svg>;
const ChevronDown = () => <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="m6 9 6 6 6-6"/></svg>;
const CollapseIcon = () => <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><rect width="18" height="18" x="3" y="3" rx="2"/><path d="M9 3v18"/><path d="m14 9 3 3-3 3"/></svg>;
const ExpandIcon = () => <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><rect width="18" height="18" x="3" y="3" rx="2"/><path d="M9 3v18"/><path d="m16 15-3-3 3-3"/></svg>;
const DotsIcon = () => <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><circle cx="12" cy="5" r="1"/><circle cx="12" cy="12" r="1"/><circle cx="12" cy="19" r="1"/></svg>;
const CrownIcon = () => <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="m2 4 3 12h14l3-12-6 7-4-7-4 7-6-7zm3 16h14"/></svg>;
const GearIcon = () => <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><path d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z"/><circle cx="12" cy="12" r="3"/></svg>;
const CardIcon = () => <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><rect width="20" height="14" x="2" y="5" rx="2"/><line x1="2" x2="22" y1="10" y2="10"/></svg>;
const SunIcon = () => <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><circle cx="12" cy="12" r="4"/><path d="M12 2v2M12 20v2M4.93 4.93l1.41 1.41M17.66 17.66l1.41 1.41M2 12h2M20 12h2M6.34 17.66l-1.41 1.41M19.07 4.93l-1.41 1.41"/></svg>;
const SignOutIcon = () => <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/><polyline points="16 17 21 12 16 7"/><line x1="21" x2="9" y1="12" y2="12"/></svg>;
const ActivityIcon = () => <svg width="19" height="19" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M22 12h-4l-3 9L9 3l-3 9H2"/></svg>;
const FileTextIcon = () => <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><path d="M14.5 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7.5L14.5 2z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/></svg>;

function Tooltip({ label }: { label: string }) {
  return (
    <div className="absolute left-full ml-4 z-[100] px-3 py-1.5 bg-[#0e0e0e]/95 border border-white/10 backdrop-blur-md text-white text-[11px] font-semibold tracking-wide uppercase rounded-lg shadow-2xl opacity-0 scale-95 translate-x-[-4px] group-hover:opacity-100 group-hover:scale-100 group-hover:translate-x-0 pointer-events-none whitespace-nowrap transition-all duration-200">
      {label}
    </div>
  );
}

interface NavItemProps {
  label: string;
  icon: React.ReactNode;
  isActive: boolean;
  onClick: () => void;
}

interface NavRowProps extends NavItemProps {
  badge?: string;
  badgeCls?: string;
  showChevron?: boolean;
}

function NavRow({ label, icon, isActive, onClick, badge, badgeCls, showChevron }: NavRowProps) {
  return (
    <button 
      onClick={onClick}
      className={`group/row w-full flex items-center gap-3.5 px-3.5 py-[11px] rounded-xl text-left transition-all duration-300 border border-transparent relative overflow-hidden
        ${isActive 
          ? 'bg-gradient-to-r from-[#00E3AA]/10 to-[#00b589]/5 text-[#00E3AA] border-[#00E3AA]/20 shadow-[0_4px_20px_-4px_rgba(0,227,170,0.15)] font-semibold' 
          : 'text-[#9ca3af] hover:bg-white/[0.03] hover:text-white hover:border-white/5'
        }`}
    >
      {/* Sleek active left glow line */}
      {isActive && (
        <span className="absolute left-0 top-[20%] bottom-[20%] w-[3px] bg-gradient-to-b from-[#00E3AA] to-[#00b589] rounded-r-full shadow-[0_0_10px_#00E3AA]" />
      )}
      <span className={`shrink-0 transition-transform duration-300 group-hover/row:scale-110 ${isActive ? 'text-[#00E3AA]' : 'text-[#9ca3af] group-hover/row:text-white'}`}>
        {icon}
      </span>
      <span className="flex-1 text-[14.5px] font-medium leading-none tracking-wide">{label}</span>
      {badge && (
        <span className={`text-[9px] font-bold uppercase px-2 py-0.5 rounded-full border tracking-wider ${badgeCls}`}>
          {badge}
        </span>
      )}
      {showChevron && <span className="text-[#5a5a5a] group-hover/row:text-white transition-colors"><ChevronDown /></span>}
    </button>
  );
}

function SubRow({ label, icon, badge, badgeCls, onClick, isActive }: { label: string; icon: React.ReactNode; badge?: string; badgeCls?: string; onClick?: () => void; isActive?: boolean }) {
  return (
    <div 
      onClick={onClick}
      className={`group/sub flex items-center gap-3 pl-[44px] pr-3.5 py-[9px] rounded-xl cursor-pointer transition-all duration-300 border border-transparent
        ${isActive 
          ? 'bg-gradient-to-r from-white/[0.04] to-transparent text-white font-medium border-white/5' 
          : 'text-[#9ca3af] hover:text-white hover:bg-white/[0.02]'
        }`}
    >
      <span className={`shrink-0 transition-transform duration-300 group-hover/sub:scale-105 ${isActive ? 'text-white' : 'text-[#7a7a7a] group-hover/sub:text-[#9ca3af]'}`}>
        {icon}
      </span>
      <span className="flex-1 text-[13.5px] font-medium leading-none tracking-wide">{label}</span>
      {badge && (
        <span className={`text-[9.5px] font-bold uppercase px-2 py-0.5 rounded-full border tracking-wide ${badgeCls}`}>
          {badge}
        </span>
      )}
    </div>
  );
}

function ColIconBtn({ label, icon, isActive, onClick }: NavItemProps) {
  return (
    <button 
      onClick={onClick}
      className={`group relative w-11 h-11 flex items-center justify-center rounded-xl transition-all duration-300 border
        ${isActive 
          ? 'bg-gradient-to-br from-[#00E3AA]/15 to-[#00b589]/5 text-[#00E3AA] border-[#00E3AA]/30 shadow-[0_0_15px_rgba(0,227,170,0.15)]' 
          : 'text-[#9ca3af] hover:bg-white/[0.04] hover:text-white border-transparent hover:border-white/5'
        }`}
    >
      {/* Left indicator line */}
      {isActive && (
        <span className="absolute left-[-4px] top-[25%] bottom-[25%] w-[3px] bg-gradient-to-b from-[#00E3AA] to-[#00b589] rounded-r-full shadow-[0_0_8px_#00E3AA]" />
      )}
      <span className="transition-transform duration-300 group-hover:scale-110">
        {icon}
      </span>
      <Tooltip label={label} />
    </button>
  );
}

// ---- Profile Dropdown ----
function ProfileDropdown({ user, initials, onClose, planLabel, onNavigate, className = "bottom-full left-0 mb-2 w-full" }: { user: any; initials: string; onClose: () => void; planLabel: string | null; onNavigate?: (action: () => void) => void; className?: string }) {
  const router = useRouter();
  const app = useStackApp();

  const handleLogout = async () => {
    clearApiKey();
    await app.signOut();
    onClose();
    router.push("/handler/sign-in");
  };

  const menuItem = (icon: React.ReactNode, label: string, onClick: () => void, additionalClasses = "") => (
    <button onClick={onClick}
      className={`w-full flex items-center gap-3 px-4 py-2.5 text-[14px] text-[#d1d5db] hover:bg-[#1c1c1c] hover:text-white transition-all duration-150 ${additionalClasses}`}>
      <span className="text-[#9ca3af] shrink-0">{icon}</span>
      {label}
    </button>
  );

  const navigate = (path: string) => {
    const go = () => {
      router.push(path);
      onClose();
    };
    if (onNavigate) {
      onNavigate(go);
    } else {
      go();
    }
  };

  const handleLogoutRequest = () => {
    const go = () => {
      handleLogout();
    };
    if (onNavigate) {
      onNavigate(go);
    } else {
      go();
    }
  };

  return (
    <div className={`absolute ${className} bg-[#161616] border border-[#242424] rounded-2xl shadow-[0_0_40px_rgba(0,0,0,0.5)] overflow-hidden z-[200]`}>
      {/* User header */}
      <div className="flex items-center gap-3 px-4 py-3.5 border-b border-[#242424]">
        {user.profileImageUrl ? (
          <Image src={user.profileImageUrl} alt={initials} width={36} height={36} className="rounded-xl shrink-0" />
        ) : (
          <div className="w-9 h-9 rounded-xl bg-gradient-to-br from-[#00E3AA] to-[#00b589] flex items-center justify-center text-black font-bold text-[13px] shrink-0">
            {initials}
          </div>
        )}
        <div className="flex flex-col min-w-0">
          <span className="text-white text-[13px] font-semibold truncate">{user.primaryEmail || user.displayName}</span>
          {planLabel !== null && <span className="text-[#6b7280] text-[11px]">{planLabel}</span>}
        </div>
      </div>

      {/* Menu items */}
      <div className="py-1">
        {planLabel === "Free Plan" && menuItem(<CrownIcon />, "Upgrade to Pro", () => navigate("/dashboard/settings/billing-and-subscription"), "text-yellow-400 hover:text-yellow-300")}
        <div className="border-[#1f1f1f] my-1" />
        {menuItem(<CardIcon />, "Plan & Billing", () => navigate("/dashboard/settings/billing-and-subscription"))}
        <div className="border-t border-[#1f1f1f] my-1" />
        {/* Light mode — dummy (coming soon) */}
        {menuItem(<SignOutIcon />, "Sign Out", handleLogoutRequest, "text-red-400")}
      </div>
    </div>
  );
}

// ---- Quick Call Dropdown ----
function QuickCallDropdown({ bots, avatars, onSelect, onClose, className = "bottom-full left-0 mb-2 w-full" }: { bots: AgentBot[]; avatars: AvatarItem[]; onSelect: (bot: AgentBot) => void; onClose: () => void; className?: string }) {
  return (
    <div className={`absolute ${className} bg-[#161616] border border-[#242424] rounded-2xl shadow-[0_0_40px_rgba(0,0,0,0.5)] overflow-hidden z-[200]`}>
      <div className="px-4 py-3 border-b border-[#242424]">
        <span className="text-white text-[12px] font-bold uppercase tracking-wider text-[#9ca3af]">Select a Agent</span>
      </div>
      <div className="py-1 max-h-[280px] overflow-y-auto custom-scrollbar">
        {bots?.length === 0 ? (
          <div className="px-4 py-6 text-center text-[#5a5a5a] text-[13px]">
            No agents found. Add an agent first.
          </div>
        ) : (
          bots?.map((bot) => {
            const botAvatarId = bot.avatars[0]?.avatar_key_id || bot.avatars[0]?.avatar_id || "";
            const avatar = avatars.find(a => a.id === botAvatarId);
            return (
              <button
                key={bot.id}
                onClick={() => {
                  onSelect(bot);
                  onClose();
                }}
                className="w-full flex items-center gap-3 px-3 py-2.5 text-[14px] text-[#d1d5db] hover:bg-[#1c1c1c] hover:text-white transition-all duration-150 group"
              >
                <div className="w-8 h-8 rounded-full overflow-hidden border border-white/10 shrink-0 group-hover:border-[#00E3AA]/40 transition-colors">
                  {avatar ? (
                    <img src={avatar.image} alt={bot.agent_name} className="w-full h-full object-cover object-top" />
                  ) : (
                    <div className="w-full h-full bg-[#1c2e28] flex items-center justify-center text-[10px] font-bold text-[#00E3AA]">
                      {bot.agent_name.charAt(0)}
                    </div>
                  )}
                </div>
                <div className="flex flex-col items-start min-w-0">
                  <span className="font-semibold truncate w-full text-left">{bot.agent_name}</span>
                  <span className="text-[10px] text-[#5a5a5a] uppercase tracking-tight">Saved Agent</span>
                </div>
              </button>
            );
          })
        )}
      </div>
    </div>
  );
}

// ─── Main Sidebar ─────────────────────────────────────────────────────────────
export function Sidebar({
  activeSession, setActiveSession, isMobileMenuOpen, setIsMobileMenuOpen, bots = [], onQuickCall = () => {}, avatars = [], gatewayError = false, onNavigate
}: {
  activeSession: string;
  setActiveSession: (s: string) => void;
  isMobileMenuOpen: boolean;
  setIsMobileMenuOpen: (o: boolean) => void;
  bots?: AgentBot[];
  onQuickCall?: (bot: AgentBot) => void | Promise<void>;
  avatars?: AvatarItem[];
  gatewayError?: boolean;
  onNavigate?: (action: () => void) => void;
}) {
  const [isCollapsed, setIsCollapsed] = useState(false);
  const [dropdownOpen, setDropdownOpen] = useState(false);
  const [settingsOpen, setSettingsOpen] = useState(false);
  const [planLabel, setPlanLabel] = useState<string | null>(null);
  const [quickCallOpen, setQuickCallOpen] = useState(false);

  const user = useUser();
  const router = useRouter();
  const pathname = usePathname();
  const dropdownRef = useRef<HTMLDivElement>(null);


  useEffect(() => {
    let cancelled = false;
    let attempts = 0;
    const maxAttempts = 10;
    const retryDelay = 1500;

    const tryFetch = () => {
      if (cancelled) return;
      const apiKey = localStorage.getItem("defaultApiKey");
      if (!apiKey) {
        if (attempts < maxAttempts) {
          attempts++;
          setTimeout(tryFetch, retryDelay);
        }
        return;
      }
      getLicenseDetails(apiKey).then(({ data }) => {
        if (cancelled || !data?.slug) return;
        const slug = data.slug;
        if (slug.includes("ente_ente")) setPlanLabel("Enterprise Plan");
        else if (slug.startsWith("pro")) setPlanLabel("Pro Plan");
        else setPlanLabel("Free Plan");
      });
    };

    tryFetch();
    return () => { cancelled = true; };
  }, [user?.id]);

  // Auto-open settings section when on a settings page
  useEffect(() => {
    if (pathname?.startsWith("/dashboard/settings")) {
      setSettingsOpen(true);
    }
  }, [pathname]);
  const quickCallRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(e.target as Node)) {
        setDropdownOpen(false);
      }
      if (quickCallRef.current && !quickCallRef.current.contains(e.target as Node)) {
        setQuickCallOpen(false);
      }
    };
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, []);

  const initials = getInitials(user?.primaryEmail, user?.displayName);

  const handleNav = (session: string) => { 
    setActiveSession(session); 
    setIsMobileMenuOpen(false); 
    setQuickCallOpen(false);
  };

  const handleRoute = (path: string) => {
    const go = () => {
      router.push(path);
      setIsMobileMenuOpen(false);
      setQuickCallOpen(false);
    };
    if (onNavigate) {
      onNavigate(go);
    } else {
      go();
    }
  };

  const UserAvatar = ({ size = "w-9 h-9" }: { size?: string }) => (
    user?.profileImageUrl ? (
      <div className={`relative ${size} shrink-0 group-hover:scale-105 transition-all duration-300`}>
        <Image 
          src={user.profileImageUrl} 
          alt={initials} 
          width={36} 
          height={36} 
          className="w-full h-full rounded-xl object-cover border border-white/10 group-hover:border-[#00E3AA]/40 transition-colors" 
        />
        <span className="absolute bottom-0 right-0 w-2.5 h-2.5 rounded-full bg-[#00E3AA] border-2 border-[#09090b] shadow-[0_0_8px_#00E3AA]" />
      </div>
    ) : (
      <div className={`relative ${size} shrink-0 group-hover:scale-105 transition-all duration-300`}>
        <div className="w-full h-full rounded-xl bg-gradient-to-br from-[#00E3AA] to-[#00b589] flex items-center justify-center text-black font-bold text-[13px] shadow-[0_2px_8px_rgba(0,227,170,0.2)]">
          {initials}
        </div>
        <span className="absolute bottom-0 right-0 w-2.5 h-2.5 rounded-full bg-[#00E3AA] border-2 border-[#09090b] shadow-[0_0_8px_#00E3AA]" />
      </div>
    )
  );

  // ---- Expanded Layout ----
  const expanded = (
    <>
      {/* Header */}
      <div className="flex items-center justify-between px-4 pt-5 pb-4 shrink-0">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-[#00E3AA]/20 to-[#00b589]/5 border border-[#00E3AA]/20 flex items-center justify-center shrink-0 overflow-hidden shadow-[0_0_15px_rgba(0,227,170,0.1)]">
            <Image src="/clawdface-logo.svg" alt="ClawdFace" width={50} height={50} className="object-contain transform hover:scale-110 transition-transform duration-300" />
          </div>
          <div className="flex flex-col">
            <span className="text-white font-bold text-[17px] leading-tight tracking-[-0.015em]">ClawdFace</span>
            <span className="text-[#00E3AA] text-[11px] font-bold uppercase tracking-wider leading-none mt-0.5">Beta</span>
          </div>
        </div>
        <button 
          onClick={() => setIsCollapsed(true)} 
          className="text-[#7a7a7a] hover:text-white transition-all duration-300 p-1.5 rounded-lg hover:bg-white/5 border border-transparent hover:border-white/5"
          title="Collapse Sidebar"
        >
          <CollapseIcon />
        </button>
      </div>

      {/* Nav */}
      <nav className="flex-1 overflow-y-auto px-3 pb-2 flex flex-col gap-0.5 custom-scrollbar">
        <NavRow label="Agent Library" icon={<LibraryIcon />} isActive={activeSession === "Library" || activeSession === "DirectCall"} onClick={() => handleNav("Library")} />
        <NavRow label="Add Agent" icon={<BotIcon />} isActive={activeSession === "AddBot"} onClick={() => handleNav("AddBot")} />
        <NavRow 
          label="Gateway Doctor" 
          icon={<ActivityIcon />} 
          onClick={() => handleNav("Doctor")} 
          isActive={activeSession === "Doctor"} 
          badge={bots.length === 0 ? "Inactive" : (gatewayError ? "Offline" : "Active")} 
          badgeCls={bots.length === 0 ? "border-neutral-500/30 text-neutral-400 bg-neutral-500/5" : (gatewayError ? "border-red-500/40 text-red-400 bg-red-500/10" : "border-[#00E3AA]/20 text-[#00E3AA]")} 
        />
        <NavRow label="Stock Avatars" icon={<UserIcon />}     isActive={activeSession === "Avatars"}   onClick={() => handleNav("Avatars")} />

        <NavRow label="Conversations" icon={<HistoryIcon />} onClick={() => handleNav("Conversations")} isActive={activeSession === "Conversations"} />
        <button 
          onClick={() => setSettingsOpen(!settingsOpen)}
          className={`group/row w-full flex items-center gap-3.5 px-3.5 py-[11px] rounded-xl text-left transition-all duration-300 border border-transparent relative overflow-hidden
            ${pathname?.startsWith("/dashboard/settings")
              ? 'bg-gradient-to-r from-[#00E3AA]/10 to-[#00b589]/5 text-[#00E3AA] border-[#00E3AA]/20 shadow-[0_4px_20px_-4px_rgba(0,227,170,0.15)] font-semibold'
              : 'text-[#9ca3af] hover:bg-white/[0.03] hover:text-white hover:border-white/5'
            }`}
        >
          {pathname?.startsWith("/dashboard/settings") && (
            <span className="absolute left-0 top-[20%] bottom-[20%] w-[3px] bg-gradient-to-b from-[#00E3AA] to-[#00b589] rounded-r-full shadow-[0_0_10px_#00E3AA]" />
          )}
          <span className={`shrink-0 transition-transform duration-300 group-hover/row:scale-110 ${pathname?.startsWith("/dashboard/settings") ? 'text-[#00E3AA]' : 'text-[#9ca3af] group-hover/row:text-white'}`}>
            <CardIcon />
          </span>
          <span className="flex-1 text-[14.5px] font-medium leading-none tracking-wide">Billing</span>
          <span className={`transition-transform duration-200 ${settingsOpen ? '' : '-rotate-90'} ${pathname?.startsWith("/dashboard/settings") ? 'text-[#00E3AA]' : 'text-[#5a5a5a] group-hover/row:text-white'}`}>
            <ChevronDown />
          </span>
        </button>
        {settingsOpen && <div className="flex flex-col gap-0.5">
          <SubRow
            label="Plan & Subscription"
            icon={<CardIcon />}
            onClick={() => handleRoute("/dashboard/settings/billing-and-subscription")}
            isActive={pathname === "/dashboard/settings/billing-and-subscription"}
          />
          <SubRow
            label="Invoice History"
            icon={<FileTextIcon />}
            onClick={() => handleRoute("/dashboard/settings/invoices")}
            isActive={pathname === "/dashboard/settings/invoices"}
          />
        </div>}
      </nav>

      {/* Footer / Configuration */}
      <div className="px-3 pb-4 flex flex-col gap-0.5 shrink-0">
        <div className="bg-gradient-to-r from-transparent via-white/[0.08] to-transparent h-[1px] w-full my-3" />
        
        <div ref={quickCallRef} className="relative">
          {quickCallOpen && (
            <QuickCallDropdown bots={bots} avatars={avatars} onSelect={onQuickCall} onClose={() => setQuickCallOpen(false)} />
          )}
          <NavRow 
            label="Quick Call" 
            icon={<MonitorIcon />} 
            isActive={activeSession === "My Bot"} 
            onClick={() => setQuickCallOpen(!quickCallOpen)} 
            showChevron={true}
          />
        </div>

        <div className="bg-gradient-to-r from-transparent via-white/[0.08] to-transparent h-[1px] w-full my-3" />

        {/* Profile + dropdown */}
        <div ref={dropdownRef} className="relative">
          {dropdownOpen && user && (
            <ProfileDropdown user={user} initials={initials} onClose={() => setDropdownOpen(false)} planLabel={planLabel} onNavigate={onNavigate} />
          )}
          <div 
            onClick={() => setDropdownOpen(!dropdownOpen)}
            className="group flex items-center gap-3 px-2 py-2.5 rounded-xl hover:bg-white/[0.03] border border-transparent hover:border-white/5 cursor-pointer transition-all duration-300"
          >
            <UserAvatar />
            <div className="flex flex-col flex-1 min-w-0">
              <span className="text-white text-[14.5px] font-semibold leading-tight truncate tracking-wide">
                {user ? (user.displayName || user.primaryEmail) : "Loading..."}
              </span>
              {planLabel !== null && <span className="text-[#7a7a7a] text-[11px] leading-tight font-medium uppercase tracking-wider mt-0.5">{planLabel}</span>}
            </div>
            <button className="text-[#5a5a5a] group-hover:text-[#9ca3af] transition-colors p-1 shrink-0">
              <DotsIcon />
            </button>
          </div>
        </div>
      </div>
    </>
  );

  // ---- Collapsed Layout ----
  const collapsed = (
    <>
      <div className="flex flex-col items-center pt-5 pb-2 shrink-0 gap-3">
        <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-[#00E3AA]/20 to-[#00b589]/5 border border-[#00E3AA]/20 flex items-center justify-center overflow-hidden shadow-[0_0_15px_rgba(0,227,170,0.1)]">
          <Image src="/clawdface-logo.svg" alt="ClawdFace" width={26} height={26} className="object-contain transform hover:scale-110 transition-transform duration-300" />
        </div>
        <button 
          onClick={() => setIsCollapsed(false)} 
          className="text-[#7a7a7a] hover:text-[#00E3AA] transition-all duration-300 p-1.5 rounded-xl hover:bg-[#00E3AA]/10 border border-transparent hover:border-[#00E3AA]/20"
          title="Expand Sidebar"
        >
          <ExpandIcon />
        </button>
      </div>
      <div className="bg-gradient-to-r from-transparent via-white/[0.08] to-transparent h-[1px] w-full my-3" />
      <nav className="flex-1 overflow-y-auto flex flex-col items-center gap-2.5 px-2 custom-scrollbar">
        <ColIconBtn label="Agent Library" icon={<LibraryIcon />} isActive={activeSession === "Library" || activeSession === "DirectCall"} onClick={() => handleNav("Library")} />
        <ColIconBtn label="Add Agent" icon={<BotIcon />} isActive={activeSession === "AddBot"} onClick={() => handleNav("AddBot")} />
        <ColIconBtn label="Doctor" icon={<ActivityIcon />} isActive={activeSession === "Doctor"} onClick={() => handleNav("Doctor")} />
        <ColIconBtn label="Stock Avatars" icon={<UserIcon />}     isActive={activeSession === "Avatars"}   onClick={() => handleNav("Avatars")} />
        <ColIconBtn label="Conversations" icon={<HistoryIcon />} isActive={activeSession === "Conversations"} onClick={() => handleNav("Conversations")} />
        <ColIconBtn label="Billing" icon={<CardIcon />} isActive={!!pathname?.startsWith("/dashboard/settings")} onClick={() => handleRoute("/dashboard/settings/billing-and-subscription")} />
      </nav>
      <div className="flex flex-col items-center px-2 pb-4 shrink-0 gap-3">
        <div className="bg-gradient-to-r from-transparent via-white/[0.08] to-transparent h-[1px] w-full my-3" />
        
        <div ref={quickCallRef} className="relative w-full flex justify-center group">
          {quickCallOpen && (
            <QuickCallDropdown
              bots={bots}
              avatars={avatars}
              onSelect={onQuickCall}
              onClose={() => setQuickCallOpen(false)}
              className="bottom-0 left-full ml-4 w-[260px]"
            />
          )}
          <button 
            onClick={() => setQuickCallOpen(!quickCallOpen)} 
            className={`group relative w-11 h-11 flex items-center justify-center rounded-xl transition-all duration-300 border
              ${activeSession === 'My Bot' 
                ? 'bg-gradient-to-br from-[#00E3AA]/15 to-[#00b589]/5 text-[#00E3AA] border-[#00E3AA]/30 shadow-[0_0_15px_rgba(0,227,170,0.15)]' 
                : 'text-[#9ca3af] hover:bg-white/[0.04] hover:text-white border-transparent hover:border-white/5'
              }`}
          >
             <MonitorIcon />
             <Tooltip label="Quick Call" />
          </button>
        </div>

        <div className="bg-gradient-to-r from-transparent via-white/[0.08] to-transparent h-[1px] w-full my-3" />
        <div ref={dropdownRef} className="relative group w-full flex justify-center">
          {dropdownOpen && user && (
            <ProfileDropdown
              user={user}
              initials={initials}
              onClose={() => setDropdownOpen(false)}
              planLabel={planLabel}
              onNavigate={onNavigate}
              className="bottom-0 left-full ml-4 w-[260px]"
            />
          )}
          <button onClick={() => setDropdownOpen(!dropdownOpen)} className="relative p-1 shrink-0">
            <UserAvatar />
            <Tooltip label={user ? (user.primaryEmail || user.displayName || "Profile") : "Profile"} />
          </button>
        </div>
      </div>
    </>
  );

  return (
    <>
      {/* Mobile backdrop */}
      <div
        className={`md:hidden fixed inset-0 bg-black/60 z-40 backdrop-blur-sm transition-opacity duration-300 ${isMobileMenuOpen ? 'opacity-100 pointer-events-auto' : 'opacity-0 pointer-events-none'}`}
        onClick={() => setIsMobileMenuOpen(false)}
      />
      <div className={`
        h-screen bg-[#09090b] bg-[radial-gradient(circle_at_top,_var(--tw-gradient-stops))] from-[#00E3AA]/[0.02] via-transparent to-transparent border-r border-white/[0.04] flex flex-col shrink-0 z-50 overflow-hidden
        fixed md:relative inset-y-0 left-0
        transition-[width,transform] duration-300 ease-in-out
        ${isMobileMenuOpen ? 'translate-x-0' : '-translate-x-full md:translate-x-0'}
        ${isCollapsed ? 'w-[68px]' : 'w-[268px]'}
      `}>
        {isCollapsed ? collapsed : expanded}
      </div>
    </>
  );
}

