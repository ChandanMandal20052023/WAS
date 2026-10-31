import React, { useState } from 'react';
import { Link, useLocation, Outlet, useNavigate } from 'react-router-dom';
import {
  LayoutDashboard,
  MessageSquare,
  Users,
  Radio,
  BarChart3,
  Settings,
  ShieldCheck,
  Menu,
  X,
  Sun,
  Moon,
  LogOut,
  Building2,
} from 'lucide-react';

import { useAuthStore } from '../stores/authStore';
import { useWebSocket } from '../hooks/useWebSocket';

const navItems = [
  { name: 'Dashboard', path: '/', icon: LayoutDashboard },
  { name: 'Inbox', path: '/inbox', icon: MessageSquare, badge: 'Live' },
  { name: 'Contacts', path: '/contacts', icon: Users },
  { name: 'WhatsApp Channels', path: '/channels', icon: Radio },
  { name: 'Team', path: '/team', icon: ShieldCheck },
  { name: 'Analytics', path: '/analytics', icon: BarChart3 },
  { name: 'Settings', path: '/settings', icon: Settings },
];

export const DashboardLayout: React.FC = () => {
  const location = useLocation();
  const navigate = useNavigate();
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const [isDarkMode, setIsDarkMode] = useState(false);

  const { user, organization, role, logout } = useAuthStore();
  const { isConnected: isWSConnected } = useWebSocket();

  const toggleDarkMode = () => {
    setIsDarkMode(!isDarkMode);
    document.documentElement.classList.toggle('dark');
  };

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  const userInitials = user?.full_name
    ? user.full_name
        .split(' ')
        .map((n) => n[0])
        .join('')
        .toUpperCase()
        .slice(0, 2)
    : 'U';

  return (
    <div className="flex h-screen w-full bg-background overflow-hidden">
      {/* Sidebar Navigation */}
      <aside
        className={`fixed inset-y-0 left-0 z-50 w-64 bg-card border-r border-border flex flex-col transition-transform duration-200 md:static md:translate-x-0 ${
          mobileMenuOpen ? 'translate-x-0' : '-translate-x-full'
        }`}
      >
        {/* Brand Header */}
        <div className="flex items-center justify-between h-16 px-6 border-b border-border">
          <div className="flex items-center gap-3">
            <div className="w-8 h-8 rounded-lg bg-whatsapp flex items-center justify-center text-white font-bold text-lg shadow-sm">
              W
            </div>
            <div>
              <span className="font-bold text-lg tracking-tight text-foreground">WAS</span>
              <span className="block text-xs text-muted-foreground font-medium">WhatsApp CRM</span>
            </div>
          </div>
          <button
            onClick={() => setMobileMenuOpen(false)}
            className="md:hidden text-muted-foreground hover:text-foreground"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Current Organization Badge */}
        <div className="px-4 py-3 border-b border-border bg-muted/40">
          <div className="flex items-center gap-2">
            <Building2 className="w-4 h-4 text-primary flex-shrink-0" />
            <span className="text-xs font-semibold truncate">
              {organization?.name || 'Default Organization'}
            </span>
            <span className="ml-auto text-[10px] bg-primary/10 text-primary font-bold px-1.5 py-0.5 rounded">
              {role || 'Agent'}
            </span>
          </div>
        </div>

        {/* Nav Links */}
        <nav className="flex-1 px-3 py-4 space-y-1 overflow-y-auto">
          {navItems.map((item) => {
            const isActive = location.pathname === item.path;
            const Icon = item.icon;
            return (
              <Link
                key={item.path}
                to={item.path}
                onClick={() => setMobileMenuOpen(false)}
                className={`flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-colors ${
                  isActive
                    ? 'bg-primary text-primary-foreground shadow-sm'
                    : 'text-muted-foreground hover:bg-muted hover:text-foreground'
                }`}
              >
                <Icon className="w-4 h-4" />
                <span>{item.name}</span>
                {item.badge && (
                  <span
                    className={`ml-auto text-[10px] font-bold px-1.5 py-0.5 rounded ${
                      isActive ? 'bg-white/20 text-white' : 'bg-whatsapp/10 text-whatsapp'
                    }`}
                  >
                    {item.badge}
                  </span>
                )}
              </Link>
            );
          })}
        </nav>

        {/* Footer Actions */}
        <div className="p-4 border-t border-border flex items-center justify-between">
          <button
            onClick={toggleDarkMode}
            className="p-2 rounded-lg text-muted-foreground hover:bg-muted hover:text-foreground transition-colors"
            title="Toggle Theme"
          >
            {isDarkMode ? <Sun className="w-4 h-4" /> : <Moon className="w-4 h-4" />}
          </button>
          <div className="flex items-center gap-2">
            <div
              className="w-8 h-8 rounded-full bg-primary/20 text-primary font-bold flex items-center justify-center text-xs"
              title={user?.email || 'User'}
            >
              {userInitials}
            </div>
            <button
              onClick={handleLogout}
              className="text-muted-foreground hover:text-destructive p-1.5 rounded transition-colors"
              title="Logout"
            >
              <LogOut className="w-4 h-4" />
            </button>
          </div>
        </div>
      </aside>

      {/* Main Content Area */}
      <div className="flex-1 flex flex-col h-screen overflow-hidden">
        {/* Top Navbar */}
        <header className="h-16 border-b border-border bg-card/60 backdrop-blur px-6 flex items-center justify-between flex-shrink-0">
          <div className="flex items-center gap-3">
            <button
              onClick={() => setMobileMenuOpen(true)}
              className="md:hidden text-muted-foreground hover:text-foreground"
            >
              <Menu className="w-6 h-6" />
            </button>
            <h1 className="text-base font-semibold text-foreground capitalize">
              {navItems.find((i) => i.path === location.pathname)?.name || 'Dashboard'}
            </h1>
          </div>
          <div className="flex items-center gap-3">
            <div className="flex items-center gap-1.5 text-xs text-muted-foreground bg-muted px-2.5 py-1 rounded-full">
              <span
                className={`w-2 h-2 rounded-full ${
                  isWSConnected ? 'bg-emerald-500 animate-pulse' : 'bg-amber-500'
                }`}
              ></span>
              <span className="font-mono text-[11px]">
                {isWSConnected ? 'Live Socket: Online' : 'Connecting...'}
              </span>
            </div>
            <div className="flex items-center gap-2 text-xs text-muted-foreground bg-muted px-2.5 py-1 rounded-full">
              <span className="w-2 h-2 rounded-full bg-emerald-500"></span>
              <span>Tenant: {organization?.slug || 'was'}</span>
            </div>
          </div>
        </header>

        {/* Page Content Viewport */}
        <main className="flex-1 overflow-y-auto">
          <Outlet />
        </main>
      </div>
    </div>
  );
};
