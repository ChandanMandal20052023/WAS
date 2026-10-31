import React from 'react';
import {
  Users,
  MessageSquare,
  Clock,
  CheckCircle2,
  Radio,
  ArrowUpRight,
} from 'lucide-react';
import { Link } from 'react-router-dom';

export const DashboardPage: React.FC = () => {
  return (
    <div className="p-6 max-w-7xl mx-auto space-y-6">
      {/* Welcome Banner */}
      <div className="flex flex-col md:flex-row items-start md:items-center justify-between gap-4 bg-gradient-to-r from-whatsapp/15 via-primary/10 to-transparent p-6 rounded-2xl border border-whatsapp/20">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">Welcome to WAS WhatsApp CRM</h2>
          <p className="text-sm text-muted-foreground mt-1">
            Real-time multi-tenant dashboard for managing WhatsApp channels, conversations, and agent performance.
          </p>
        </div>
        <Link
          to="/channels"
          className="inline-flex items-center gap-2 bg-whatsapp hover:bg-whatsapp-dark text-white px-4 py-2.5 rounded-xl font-medium text-sm transition-all shadow-sm hover:shadow"
        >
          <Radio className="w-4 h-4" />
          <span>Connect WhatsApp</span>
        </Link>
      </div>

      {/* Metrics Grid */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="bg-card border border-border p-5 rounded-xl shadow-sm">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground">Total Contacts</span>
            <div className="p-2 bg-primary/10 rounded-lg text-primary">
              <Users className="w-4 h-4" />
            </div>
          </div>
          <div className="mt-3">
            <span className="text-2xl font-bold tracking-tight">1,248</span>
            <span className="text-xs text-emerald-600 font-medium ml-2 inline-flex items-center">
              +12% <ArrowUpRight className="w-3 h-3" />
            </span>
          </div>
        </div>

        <div className="bg-card border border-border p-5 rounded-xl shadow-sm">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground">Open Conversations</span>
            <div className="p-2 bg-whatsapp/10 rounded-lg text-whatsapp">
              <MessageSquare className="w-4 h-4" />
            </div>
          </div>
          <div className="mt-3">
            <span className="text-2xl font-bold tracking-tight text-whatsapp">42</span>
            <span className="text-xs text-muted-foreground ml-2">Active now</span>
          </div>
        </div>

        <div className="bg-card border border-border p-5 rounded-xl shadow-sm">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground">Pending Follow-ups</span>
            <div className="p-2 bg-amber-500/10 rounded-lg text-amber-500">
              <Clock className="w-4 h-4" />
            </div>
          </div>
          <div className="mt-3">
            <span className="text-2xl font-bold tracking-tight text-amber-500">9</span>
            <span className="text-xs text-muted-foreground ml-2">Awaiting agent</span>
          </div>
        </div>

        <div className="bg-card border border-border p-5 rounded-xl shadow-sm">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground">Resolved Today</span>
            <div className="p-2 bg-blue-500/10 rounded-lg text-blue-500">
              <CheckCircle2 className="w-4 h-4" />
            </div>
          </div>
          <div className="mt-3">
            <span className="text-2xl font-bold tracking-tight">85</span>
            <span className="text-xs text-emerald-600 font-medium ml-2">98% satisfaction</span>
          </div>
        </div>
      </div>

      {/* Main Content Panels */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Active WhatsApp Channels Overview */}
        <div className="lg:col-span-2 bg-card border border-border rounded-xl p-6 shadow-sm space-y-4">
          <div className="flex items-center justify-between">
            <h3 className="font-semibold text-base">Connected WhatsApp Channels</h3>
            <Link to="/channels" className="text-xs font-medium text-primary hover:underline">
              View all
            </Link>
          </div>
          <div className="divide-y divide-border">
            <div className="py-3 flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div className="w-9 h-9 rounded-full bg-whatsapp/15 flex items-center justify-center text-whatsapp font-bold">
                  <Radio className="w-4 h-4" />
                </div>
                <div>
                  <h4 className="text-sm font-medium">Support Line (Sales)</h4>
                  <p className="text-xs text-muted-foreground">+1 (555) 234-5678</p>
                </div>
              </div>
              <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold bg-emerald-500/10 text-emerald-600">
                <span className="w-1.5 h-1.5 rounded-full bg-emerald-500"></span>
                Connected
              </span>
            </div>

            <div className="py-3 flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div className="w-9 h-9 rounded-full bg-muted flex items-center justify-center text-muted-foreground font-bold">
                  <Radio className="w-4 h-4" />
                </div>
                <div>
                  <h4 className="text-sm font-medium">Customer Helpdesk</h4>
                  <p className="text-xs text-muted-foreground">+1 (555) 987-6543</p>
                </div>
              </div>
              <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold bg-amber-500/10 text-amber-600">
                <span className="w-1.5 h-1.5 rounded-full bg-amber-500"></span>
                Waiting for QR Scan
              </span>
            </div>
          </div>
        </div>

        {/* Live System Status */}
        <div className="bg-card border border-border rounded-xl p-6 shadow-sm space-y-4">
          <h3 className="font-semibold text-base">Platform Infrastructure</h3>
          <div className="space-y-3">
            <div className="flex items-center justify-between text-sm p-2.5 rounded-lg bg-muted/40">
              <span className="text-muted-foreground font-medium">PostgreSQL Database</span>
              <span className="text-xs font-bold text-emerald-600">Operational</span>
            </div>
            <div className="flex items-center justify-between text-sm p-2.5 rounded-lg bg-muted/40">
              <span className="text-muted-foreground font-medium">Redis Event Hub</span>
              <span className="text-xs font-bold text-emerald-600">Active</span>
            </div>
            <div className="flex items-center justify-between text-sm p-2.5 rounded-lg bg-muted/40">
              <span className="text-muted-foreground font-medium">Evolution Go Gateway</span>
              <span className="text-xs font-bold text-emerald-600">Ready</span>
            </div>
            <div className="flex items-center justify-between text-sm p-2.5 rounded-lg bg-muted/40">
              <span className="text-muted-foreground font-medium">WebSocket Stream</span>
              <span className="text-xs font-bold text-emerald-600">Listening</span>
            </div>
          </div>
          <div className="pt-2 border-t border-border">
            <p className="text-xs text-muted-foreground text-center">
              WAS v1.0.0 — Multi-Tenant Architecture
            </p>
          </div>
        </div>
      </div>
    </div>
  );
};
