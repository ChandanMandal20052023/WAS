import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  Users,
  UserPlus,
  Shield,
  Mail,
  Loader2,
  X,
  AlertCircle,
  CheckCircle2,
} from 'lucide-react';
import { teamService } from '../../services/teamService';
import { Role } from '../../types/auth';
import { useAuthStore } from '../../stores/authStore';

export const TeamPage: React.FC = () => {
  const queryClient = useQueryClient();
  const currentUserRole = useAuthStore((state) => state.role);
  const currentUserId = useAuthStore((state) => state.user?.id);

  const canManage = currentUserRole === 'Owner' || currentUserRole === 'Admin';

  const [isInviteModalOpen, setIsInviteModalOpen] = useState(false);
  const [inviteEmail, setInviteEmail] = useState('');
  const [inviteRole, setInviteRole] = useState<Role>('Agent');
  const [inviteSuccessMsg, setInviteSuccessMsg] = useState<string | null>(null);
  const [inviteErrorMsg, setInviteErrorMsg] = useState<string | null>(null);

  const { data: members = [], isLoading } = useQuery({
    queryKey: ['team-members'],
    queryFn: () => teamService.getMembers(),
  });

  const updateRoleMutation = useMutation({
    mutationFn: ({ id, role }: { id: string; role: Role }) =>
      teamService.updateMemberRole(id, role),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['team-members'] });
    },
  });

  const updateStatusMutation = useMutation({
    mutationFn: ({ id, isActive }: { id: string; isActive: boolean }) =>
      teamService.updateMemberStatus(id, isActive),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['team-members'] });
    },
  });

  const inviteMutation = useMutation({
    mutationFn: () => teamService.inviteMember(inviteEmail, inviteRole),
    onSuccess: (data) => {
      setInviteSuccessMsg(`Invitation dispatched! Shareable invite link: ${data?.invite_url || 'Token generated'}`);
      setInviteErrorMsg(null);
      setInviteEmail('');
      queryClient.invalidateQueries({ queryKey: ['team-members'] });
    },
    onError: (err: any) => {
      setInviteErrorMsg(err.response?.data?.error || 'Failed to send invite');
      setInviteSuccessMsg(null);
    },
  });

  const handleSendInvite = (e: React.FormEvent) => {
    e.preventDefault();
    setInviteErrorMsg(null);
    setInviteSuccessMsg(null);
    inviteMutation.mutate();
  };

  return (
    <div className="p-6 max-w-7xl mx-auto space-y-6">
      {/* Header */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-foreground">Team Management</h1>
          <p className="text-sm text-muted-foreground">
            Manage agents, assign role-based access control, and monitor live conversation workloads.
          </p>
        </div>
        {canManage && (
          <button
            onClick={() => {
              setInviteEmail('');
              setInviteSuccessMsg(null);
              setInviteErrorMsg(null);
              setIsInviteModalOpen(true);
            }}
            className="inline-flex items-center gap-1.5 rounded-lg bg-emerald-600 px-4 py-2 text-sm font-medium text-white hover:bg-emerald-700 transition shadow-sm"
          >
            <UserPlus className="h-4 w-4" />
            Invite Member
          </button>
        )}
      </div>

      {/* Team Members List */}
      <div className="bg-card rounded-xl border border-border overflow-hidden">
        {isLoading ? (
          <div className="flex items-center justify-center py-16 text-muted-foreground text-sm">
            <Loader2 className="h-5 w-5 animate-spin mr-2" />
            Loading team members...
          </div>
        ) : members.length === 0 ? (
          <div className="text-center py-16 px-4 text-muted-foreground">
            <Users className="mx-auto h-12 w-12 opacity-30 mb-2" />
            <h3 className="text-base font-semibold text-foreground">No team members found</h3>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-sm">
              <thead className="bg-muted/50 text-xs uppercase text-muted-foreground border-b border-border">
                <tr>
                  <th className="px-6 py-3.5">Agent</th>
                  <th className="px-6 py-3.5">Role</th>
                  <th className="px-6 py-3.5">Workload (Open / Pending / Resolved)</th>
                  <th className="px-6 py-3.5">Status</th>
                  {canManage && <th className="px-6 py-3.5 text-right">Actions</th>}
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {members.map((member) => {
                  const isSelf = member.user_id === currentUserId;
                  return (
                    <tr key={member.id} className="hover:bg-muted/30 transition">
                      <td className="px-6 py-4">
                        <div className="flex items-center gap-3">
                          <div className="h-9 w-9 rounded-full bg-emerald-500/10 text-emerald-600 flex items-center justify-center font-bold text-sm">
                            {member.user?.full_name?.charAt(0).toUpperCase() || 'U'}
                          </div>
                          <div>
                            <div className="font-semibold text-foreground flex items-center gap-1.5">
                              {member.user?.full_name}
                              {isSelf && (
                                <span className="text-[10px] bg-muted px-1.5 py-0.2 rounded font-mono text-muted-foreground">
                                  You
                                </span>
                              )}
                            </div>
                            <div className="text-xs text-muted-foreground flex items-center gap-1">
                              <Mail className="h-3 w-3" />
                              {member.user?.email}
                            </div>
                          </div>
                        </div>
                      </td>

                      <td className="px-6 py-4">
                        {canManage && !isSelf ? (
                          <select
                            value={member.role}
                            onChange={(e) =>
                              updateRoleMutation.mutate({
                                id: member.id,
                                role: e.target.value as Role,
                              })
                            }
                            className="rounded-lg border border-border bg-background px-2.5 py-1 text-xs font-semibold capitalize text-foreground focus:outline-none"
                          >
                            <option value="Owner">Owner</option>
                            <option value="Admin">Admin</option>
                            <option value="Manager">Manager</option>
                            <option value="Agent">Agent</option>
                          </select>
                        ) : (
                          <span
                            className={`inline-flex items-center gap-1 text-xs font-semibold px-2.5 py-0.5 rounded-full capitalize ${
                              member.role === 'Owner'
                                ? 'bg-purple-500/15 text-purple-600'
                                : member.role === 'Admin'
                                ? 'bg-blue-500/15 text-blue-600'
                                : member.role === 'Manager'
                                ? 'bg-amber-500/15 text-amber-600'
                                : 'bg-emerald-500/15 text-emerald-600'
                            }`}
                          >
                            <Shield className="h-3 w-3" />
                            {member.role}
                          </span>
                        )}
                      </td>

                      <td className="px-6 py-4">
                        <div className="flex items-center gap-3 text-xs">
                          <span className="inline-flex items-center gap-1 font-semibold text-emerald-600 bg-emerald-500/10 px-2 py-0.5 rounded">
                            {member.open_conversations} Open
                          </span>
                          <span className="inline-flex items-center gap-1 font-semibold text-amber-600 bg-amber-500/10 px-2 py-0.5 rounded">
                            {member.pending_conversations} Pending
                          </span>
                          <span className="inline-flex items-center gap-1 font-semibold text-muted-foreground bg-muted px-2 py-0.5 rounded">
                            {member.resolved_conversations} Resolved
                          </span>
                        </div>
                      </td>

                      <td className="px-6 py-4">
                        <span
                          className={`inline-flex items-center gap-1.5 text-xs font-medium ${
                            member.is_active ? 'text-emerald-600' : 'text-muted-foreground'
                          }`}
                        >
                          <span
                            className={`h-2 w-2 rounded-full ${
                              member.is_active ? 'bg-emerald-500' : 'bg-muted-foreground'
                            }`}
                          />
                          {member.is_active ? 'Active' : 'Inactive'}
                        </span>
                      </td>

                      {canManage && (
                        <td className="px-6 py-4 text-right">
                          {!isSelf && (
                            <button
                              onClick={() =>
                                updateStatusMutation.mutate({
                                  id: member.id,
                                  isActive: !member.is_active,
                                })
                              }
                              className="text-xs font-medium text-muted-foreground hover:text-foreground border border-border px-2.5 py-1 rounded-lg hover:bg-muted transition"
                            >
                              {member.is_active ? 'Deactivate' : 'Activate'}
                            </button>
                          )}
                        </td>
                      )}
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Invite Member Modal */}
      {isInviteModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4 backdrop-blur-sm">
          <div className="w-full max-w-md rounded-2xl border border-border bg-card p-6 shadow-2xl">
            <div className="flex items-center justify-between pb-3 border-b border-border">
              <h3 className="text-lg font-bold text-foreground">Invite Team Member</h3>
              <button
                onClick={() => setIsInviteModalOpen(false)}
                className="text-muted-foreground hover:text-foreground"
              >
                <X className="h-5 w-5" />
              </button>
            </div>

            {inviteErrorMsg && (
              <div className="mt-4 flex items-center gap-2 rounded-lg bg-rose-500/10 p-3 text-sm text-rose-500 border border-rose-500/20">
                <AlertCircle className="h-4 w-4 shrink-0" />
                <span>{inviteErrorMsg}</span>
              </div>
            )}

            {inviteSuccessMsg && (
              <div className="mt-4 flex items-start gap-2 rounded-lg bg-emerald-500/10 p-3 text-xs text-emerald-600 border border-emerald-500/20">
                <CheckCircle2 className="h-4 w-4 shrink-0 mt-0.5" />
                <span className="break-all">{inviteSuccessMsg}</span>
              </div>
            )}

            <form onSubmit={handleSendInvite} className="mt-4 space-y-4">
              <div>
                <label className="block text-xs font-semibold uppercase text-muted-foreground">
                  Email Address
                </label>
                <input
                  type="email"
                  required
                  placeholder="agent@company.com"
                  value={inviteEmail}
                  onChange={(e) => setInviteEmail(e.target.value)}
                  className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2 text-sm text-foreground focus:border-emerald-500 focus:outline-none"
                />
              </div>

              <div>
                <label className="block text-xs font-semibold uppercase text-muted-foreground">
                  Assign Organization Role
                </label>
                <select
                  value={inviteRole}
                  onChange={(e) => setInviteRole(e.target.value as Role)}
                  className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2 text-sm text-foreground focus:border-emerald-500 focus:outline-none capitalize"
                >
                  <option value="Agent">Agent (Handles customer chats and notes)</option>
                  <option value="Manager">Manager (Can assign chats and oversee teams)</option>
                  <option value="Admin">Admin (Full access to channels and settings)</option>
                </select>
              </div>

              <div className="flex items-center justify-end gap-2 pt-3 border-t border-border">
                <button
                  type="button"
                  onClick={() => setIsInviteModalOpen(false)}
                  className="rounded-lg border border-border px-4 py-2 text-sm font-medium text-foreground hover:bg-muted"
                >
                  Close
                </button>
                <button
                  type="submit"
                  disabled={!inviteEmail.trim() || inviteMutation.isPending}
                  className="inline-flex items-center gap-1.5 rounded-lg bg-emerald-600 px-4 py-2 text-sm font-medium text-white hover:bg-emerald-700 disabled:opacity-50"
                >
                  {inviteMutation.isPending && <Loader2 className="h-4 w-4 animate-spin" />}
                  Send Invitation
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};
