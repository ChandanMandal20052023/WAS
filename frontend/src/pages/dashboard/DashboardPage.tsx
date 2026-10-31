import React from 'react';
import { useQuery } from '@tanstack/react-query';
import {
  Users,
  MessageSquare,
  Clock,
  CheckCircle2,
  Radio,
  Loader2,
} from 'lucide-react';
import { Link } from 'react-router-dom';
import { apiClient } from '../../services/apiClient';

interface DashboardMetrics {
  total_contacts: number;
  total_conversations: number;
  open_conversations: number;
  pending_conversations: number;
  resolved_conversations: number;
  unread_messages: number;
  connected_channels: number;
  active_agents: number;
}

export const DashboardPage: React.FC = () => {
  const { data: metrics, isLoading } = useQuery<DashboardMetrics>({
    queryKey: ['dashboard-metrics'],
    queryFn: async () => {
      const { data } = await apiClient.get<DashboardMetrics>('/dashboard/metrics');
      return data;
    },
    refetchInterval: 10000,
  });

  return (
    <div className="p-6 max-w-7xl mx-auto space-y-6">
      {/* Welcome Banner */}
      <div className="flex flex-col md:flex-row items-start md:items-center justify-between gap-4 bg-gradient-to-r from-emerald-500/15 via-primary/10 to-transparent p-6 rounded-2xl border border-emerald-500/20">
        <div>
          <h2 className="text-2xl font-bold tracking-tight text-foreground">Welcome to WAS WhatsApp CRM</h2>
          <p className="text-sm text-muted-foreground mt-1">
            Real-time multi-tenant dashboard for managing WhatsApp channels, customer conversations, and agent workflows.
          </p>
        </div>
        <Link
          to="/channels"
          className="inline-flex items-center gap-2 bg-emerald-600 hover:bg-emerald-700 text-white px-4 py-2.5 rounded-xl font-medium text-sm transition-all shadow-sm hover:shadow"
        >
          <Radio className="w-4 h-4" />
          <span>Connect WhatsApp</span>
        </Link>
      </div>

      {/* Metrics Grid */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {/* Total Contacts */}
        <div className="bg-card border border-border p-5 rounded-xl shadow-sm">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground">Total Contacts</span>
            <div className="p-2 bg-primary/10 rounded-lg text-primary">
              <Users className="w-4 h-4" />
            </div>
          </div>
          <div className="mt-3">
            <span className="text-2xl font-bold tracking-tight text-foreground">
              {isLoading ? <Loader2 className="w-5 h-5 animate-spin" /> : metrics?.total_contacts ?? 0}
            </span>
            <span className="text-xs text-muted-foreground ml-2">Normalized E.164</span>
          </div>
        </div>

        {/* Open Conversations */}
        <div className="bg-card border border-border p-5 rounded-xl shadow-sm">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground">Open Conversations</span>
            <div className="p-2 bg-emerald-500/10 rounded-lg text-emerald-600">
              <MessageSquare className="w-4 h-4" />
            </div>
          </div>
          <div className="mt-3">
            <span className="text-2xl font-bold tracking-tight text-emerald-600">
              {isLoading ? <Loader2 className="w-5 h-5 animate-spin" /> : metrics?.open_conversations ?? 0}
            </span>
            <span className="text-xs text-muted-foreground ml-2">Active now</span>
          </div>
        </div>

        {/* Pending Follow-ups */}
        <div className="bg-card border border-border p-5 rounded-xl shadow-sm">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground">Pending Follow-ups</span>
            <div className="p-2 bg-amber-500/10 rounded-lg text-amber-500">
              <Clock className="w-4 h-4" />
            </div>
          </div>
          <div className="mt-3">
            <span className="text-2xl font-bold tracking-tight text-amber-500">
              {isLoading ? <Loader2 className="w-5 h-5 animate-spin" /> : metrics?.pending_conversations ?? 0}
            </span>
            <span className="text-xs text-muted-foreground ml-2">Awaiting action</span>
          </div>
        </div>

        {/* Resolved Conversations */}
        <div className="bg-card border border-border p-5 rounded-xl shadow-sm">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground">Resolved Conversations</span>
            <div className="p-2 bg-blue-500/10 rounded-lg text-blue-500">
              <CheckCircle2 className="w-4 h-4" />
            </div>
          </div>
          <div className="mt-3">
            <span className="text-2xl font-bold tracking-tight text-foreground">
              {isLoading ? <Loader2 className="w-5 h-5 animate-spin" /> : metrics?.resolved_conversations ?? 0}
            </span>
            <span className="text-xs text-muted-foreground ml-2">Customer satisfied</span>
          </div>
        </div>
      </div>

      {/* Main Content Panels */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Quick Links & Channel Status */}
        <div className="lg:col-span-2 bg-card border border-border rounded-xl p-6 shadow-sm space-y-4">
          <div className="flex items-center justify-between">
            <h3 className="font-semibold text-base text-foreground">WhatsApp Channels Overview</h3>
            <Link to="/channels" className="text-xs font-medium text-emerald-600 hover:underline">
              Manage channels
            </Link>
          </div>
          <div className="space-y-3">
            <div className="p-4 rounded-xl border border-border bg-background flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div className="h-10 w-10 rounded-full bg-emerald-500/10 text-emerald-600 flex items-center justify-center font-bold">
                  <Radio className="h-5 w-5" />
                </div>
                <div>
                  <h4 className="text-sm font-semibold text-foreground">Evolution Go Device Gateway</h4>
                  <p className="text-xs text-muted-foreground">
                    Connected Instances: {isLoading ? '...' : metrics?.connected_channels ?? 0}
                  </p>
                </div>
              </div>
              <Link
                to="/channels"
                className="text-xs font-medium bg-emerald-600 text-white px-3 py-1.5 rounded-lg hover:bg-emerald-700 transition"
              >
                Open Channels
              </Link>
            </div>

            <div className="p-4 rounded-xl border border-border bg-background flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div className="h-10 w-10 rounded-full bg-blue-500/10 text-blue-600 flex items-center justify-center font-bold">
                  <Users className="h-5 w-5" />
                </div>
                <div>
                  <h4 className="text-sm font-semibold text-foreground">Centralized Contact Book</h4>
                  <p className="text-xs text-muted-foreground">
                    Active Contacts: {isLoading ? '...' : metrics?.total_contacts ?? 0}
                  </p>
                </div>
              </div>
              <Link
                to="/contacts"
                className="text-xs font-medium border border-border px-3 py-1.5 rounded-lg hover:bg-muted transition text-foreground"
              >
                View Contacts
              </Link>
            </div>
          </div>
        </div>

        {/* Live System Status */}
        <div className="bg-card border border-border rounded-xl p-6 shadow-sm space-y-4">
          <h3 className="font-semibold text-base text-foreground">Platform Infrastructure</h3>
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
              <span className="text-muted-foreground font-medium">Multi-Tenant Isolation</span>
              <span className="text-xs font-bold text-emerald-600">Enforced</span>
            </div>
          </div>
          <div className="pt-2 border-t border-border">
            <p className="text-xs text-muted-foreground text-center">
              WAS v1.0.0 — Production-grade WhatsApp CRM
            </p>
          </div>
        </div>
      </div>
    </div>
  );
};
