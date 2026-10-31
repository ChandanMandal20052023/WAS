import React, { useState, useEffect, useRef } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  Search,
  Send,
  MessageSquare,
  User,
  Check,
  CheckCheck,
  FileText,
  Lock,
  ShieldCheck,
  Loader2,
  RefreshCw,
} from 'lucide-react';
import { conversationService } from '../../services/conversationService';
import { ConversationStatus, Message, InternalNote } from '../../types/conversations';

export const InboxPage: React.FC = () => {
  const queryClient = useQueryClient();

  const [selectedConvId, setSelectedConvId] = useState<string | null>(null);
  const [statusFilter, setStatusFilter] = useState<ConversationStatus | 'all'>('all');
  const [search, setSearch] = useState('');
  const [outgoingText, setOutgoingText] = useState('');
  const [newNoteText, setNewNoteText] = useState('');
  const [activeRightTab, setActiveRightTab] = useState<'info' | 'notes'>('info');

  const messagesEndRef = useRef<HTMLDivElement>(null);

  // 1. Fetch Conversations List
  const { data: convData, isLoading: isLoadingConvs, refetch: refetchConvs } = useQuery({
    queryKey: ['conversations', statusFilter, search],
    queryFn: () =>
      conversationService.getConversations({
        status: statusFilter === 'all' ? undefined : statusFilter,
        search: search || undefined,
        limit: 50,
      }),
    refetchInterval: 5000, // Poll every 5s for new chats/messages
  });

  const conversations = convData?.conversations || [];

  // Auto-select first conversation if none selected
  useEffect(() => {
    if (!selectedConvId && conversations.length > 0) {
      setSelectedConvId(conversations[0].id);
    }
  }, [conversations, selectedConvId]);

  // Active selected conversation
  const activeConversation = conversations.find((c) => c.id === selectedConvId) || null;

  // 2. Fetch Messages for Active Conversation
  const { data: messages = [], isLoading: isLoadingMessages } = useQuery({
    queryKey: ['messages', selectedConvId],
    queryFn: () => (selectedConvId ? conversationService.getMessages(selectedConvId) : Promise.resolve([])),
    enabled: !!selectedConvId,
    refetchInterval: 3000, // Real-time poll every 3s
  });

  // 3. Fetch Internal Notes for Active Conversation
  const { data: notes = [] } = useQuery({
    queryKey: ['notes', selectedConvId],
    queryFn: () => (selectedConvId ? conversationService.getNotes(selectedConvId) : Promise.resolve([])),
    enabled: !!selectedConvId,
  });

  // Auto scroll to bottom when messages update
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  // Mark conversation as read when selected
  useEffect(() => {
    if (selectedConvId && activeConversation && activeConversation.unread_count > 0) {
      conversationService.markAsRead(selectedConvId).then(() => {
        queryClient.invalidateQueries({ queryKey: ['conversations'] });
      });
    }
  }, [selectedConvId]);

  // 4. Send Message Mutation
  const sendMessageMutation = useMutation({
    mutationFn: (text: string) => conversationService.sendMessage(selectedConvId!, text),
    onSuccess: (newMsg) => {
      setOutgoingText('');
      queryClient.setQueryData(['messages', selectedConvId], (old: Message[] = []) => [...old, newMsg]);
      queryClient.invalidateQueries({ queryKey: ['conversations'] });
    },
  });

  // 5. Create Internal Note Mutation
  const createNoteMutation = useMutation({
    mutationFn: (content: string) => conversationService.createNote(selectedConvId!, content),
    onSuccess: (newNote) => {
      setNewNoteText('');
      queryClient.setQueryData(['notes', selectedConvId], (old: InternalNote[] = []) => [...old, newNote]);
    },
  });

  // 6. Update Status Mutation
  const updateStatusMutation = useMutation({
    mutationFn: (status: ConversationStatus) =>
      conversationService.updateStatus(selectedConvId!, status),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['conversations'] });
    },
  });

  const handleSendMessage = (e: React.FormEvent) => {
    e.preventDefault();
    if (!outgoingText.trim() || !selectedConvId || sendMessageMutation.isPending) return;
    sendMessageMutation.mutate(outgoingText.trim());
  };

  const handleCreateNote = (e: React.FormEvent) => {
    e.preventDefault();
    if (!newNoteText.trim() || !selectedConvId || createNoteMutation.isPending) return;
    createNoteMutation.mutate(newNoteText.trim());
  };

  const formatMessageTime = (dateStr: string) => {
    try {
      const d = new Date(dateStr);
      return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    } catch {
      return '';
    }
  };

  return (
    <div className="h-[calc(100vh-6.5rem)] flex flex-col rounded-2xl border border-border bg-card overflow-hidden shadow-sm">
      <div className="flex-1 flex overflow-hidden">
        {/* ================================================================= */}
        {/* PANEL 1: LEFT PANEL - CONVERSATION THREADS LIST                   */}
        {/* ================================================================= */}
        <div className="w-80 sm:w-96 border-r border-border flex flex-col bg-background/50">
          {/* Header & Status Filter */}
          <div className="p-3.5 border-b border-border space-y-2.5">
            <div className="flex items-center justify-between">
              <h2 className="text-base font-bold text-foreground flex items-center gap-2">
                <MessageSquare className="h-4 w-4 text-emerald-500" />
                Inbox
              </h2>
              <button
                onClick={() => refetchConvs()}
                title="Refresh Conversations"
                className="p-1 rounded-lg hover:bg-muted text-muted-foreground transition"
              >
                <RefreshCw className="h-3.5 w-3.5" />
              </button>
            </div>

            {/* Search */}
            <div className="relative">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground" />
              <input
                type="text"
                placeholder="Search chats or phone..."
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                className="w-full rounded-lg border border-border bg-card pl-8 pr-3 py-1.5 text-xs text-foreground focus:border-emerald-500 focus:outline-none"
              />
            </div>

            {/* Filter Tabs */}
            <div className="flex gap-1 bg-muted p-0.5 rounded-lg text-xs font-medium">
              {(['all', 'open', 'pending', 'resolved'] as const).map((tab) => (
                <button
                  key={tab}
                  onClick={() => setStatusFilter(tab)}
                  className={`flex-1 capitalize py-1 rounded-md text-[11px] transition ${
                    statusFilter === tab
                      ? 'bg-card text-foreground shadow-sm font-semibold'
                      : 'text-muted-foreground hover:text-foreground'
                  }`}
                >
                  {tab}
                </button>
              ))}
            </div>
          </div>

          {/* Conversations Scrollable List */}
          <div className="flex-1 overflow-y-auto divide-y divide-border/50">
            {isLoadingConvs ? (
              <div className="flex items-center justify-center py-12 text-muted-foreground text-xs">
                <Loader2 className="h-4 w-4 animate-spin mr-2" />
                Loading conversations...
              </div>
            ) : conversations.length === 0 ? (
              <div className="text-center py-12 px-4 text-muted-foreground">
                <MessageSquare className="h-8 w-8 mx-auto opacity-30 mb-2" />
                <p className="text-xs font-medium">No conversations found</p>
                <p className="text-[11px] text-muted-foreground/70 mt-0.5">
                  Incoming WhatsApp messages will appear here automatically.
                </p>
              </div>
            ) : (
              conversations.map((conv) => {
                const isSelected = conv.id === selectedConvId;
                return (
                  <button
                    key={conv.id}
                    onClick={() => setSelectedConvId(conv.id)}
                    className={`w-full text-left p-3.5 transition flex gap-3 items-start relative ${
                      isSelected
                        ? 'bg-emerald-500/10 border-l-4 border-l-emerald-500'
                        : 'hover:bg-muted/40'
                    }`}
                  >
                    <div className="h-10 w-10 shrink-0 rounded-full bg-emerald-500/15 text-emerald-600 flex items-center justify-center font-bold text-sm">
                      {conv.contact_name?.charAt(0).toUpperCase() || 'C'}
                    </div>

                    <div className="flex-1 min-w-0">
                      <div className="flex items-center justify-between gap-1">
                        <span className="font-semibold text-xs text-foreground truncate">
                          {conv.contact_name || conv.contact_phone}
                        </span>
                        <span className="text-[10px] text-muted-foreground shrink-0">
                          {formatMessageTime(conv.last_message_time)}
                        </span>
                      </div>

                      <div className="text-[11px] text-muted-foreground font-mono truncate">
                        {conv.contact_phone}
                      </div>

                      <div className="mt-1 text-xs text-muted-foreground truncate">
                        {conv.last_message_preview || 'No messages yet'}
                      </div>

                      <div className="mt-1.5 flex items-center gap-1.5 flex-wrap">
                        <span
                          className={`text-[9px] px-1.5 py-0.5 rounded font-medium uppercase tracking-wider ${
                            conv.status === 'open'
                              ? 'bg-emerald-500/15 text-emerald-600'
                              : conv.status === 'pending'
                              ? 'bg-amber-500/15 text-amber-600'
                              : 'bg-muted text-muted-foreground'
                          }`}
                        >
                          {conv.status}
                        </span>

                        {conv.assigned_agent_name && (
                          <span className="text-[10px] text-muted-foreground flex items-center gap-0.5">
                            <User className="h-2.5 w-2.5" />
                            {conv.assigned_agent_name}
                          </span>
                        )}

                        {conv.unread_count > 0 && (
                          <span className="ml-auto bg-emerald-600 text-white text-[10px] font-bold px-1.5 py-0.2 rounded-full">
                            {conv.unread_count}
                          </span>
                        )}
                      </div>
                    </div>
                  </button>
                );
              })
            )}
          </div>
        </div>

        {/* ================================================================= */}
        {/* PANEL 2: MIDDLE PANEL - MESSAGES STREAM & COMPOSER               */}
        {/* ================================================================= */}
        {activeConversation ? (
          <div className="flex-1 flex flex-col bg-card/60">
            {/* Conversation Header */}
            <div className="px-5 py-3.5 border-b border-border flex items-center justify-between bg-card">
              <div className="flex items-center gap-3">
                <div className="h-9 w-9 rounded-full bg-emerald-500/20 text-emerald-600 flex items-center justify-center font-bold text-sm">
                  {activeConversation.contact_name?.charAt(0).toUpperCase() || 'C'}
                </div>
                <div>
                  <div className="flex items-center gap-2">
                    <span className="font-bold text-sm text-foreground">
                      {activeConversation.contact_name}
                    </span>
                    <span className="text-[11px] font-mono text-muted-foreground">
                      ({activeConversation.contact_phone})
                    </span>
                  </div>
                  <div className="text-[11px] text-muted-foreground flex items-center gap-1.5">
                    <span>Channel: {activeConversation.channel_name}</span>
                  </div>
                </div>
              </div>

              {/* Status Actions */}
              <div className="flex items-center gap-2">
                <select
                  value={activeConversation.status}
                  onChange={(e) => updateStatusMutation.mutate(e.target.value as ConversationStatus)}
                  className="rounded-lg border border-border bg-background px-2.5 py-1 text-xs font-medium text-foreground focus:outline-none"
                >
                  <option value="open">Open</option>
                  <option value="pending">Pending</option>
                  <option value="resolved">Resolved</option>
                </select>
              </div>
            </div>

            {/* Messages Chat Stream */}
            <div className="flex-1 overflow-y-auto p-5 space-y-3 bg-muted/15">
              {isLoadingMessages ? (
                <div className="flex items-center justify-center py-12 text-muted-foreground text-xs">
                  <Loader2 className="h-5 w-5 animate-spin mr-2" />
                  Loading message history...
                </div>
              ) : messages.length === 0 ? (
                <div className="text-center py-16 text-muted-foreground">
                  <MessageSquare className="h-10 w-10 mx-auto opacity-20 mb-2" />
                  <p className="text-sm font-medium">No messages in this conversation yet</p>
                  <p className="text-xs text-muted-foreground/70 mt-1">
                    Send a WhatsApp message below to initiate the chat.
                  </p>
                </div>
              ) : (
                messages.map((msg) => {
                  const isAgent = msg.sender_type === 'agent';
                  return (
                    <div
                      key={msg.id}
                      className={`flex flex-col ${isAgent ? 'items-end' : 'items-start'}`}
                    >
                      <div
                        className={`max-w-[70%] rounded-2xl px-4 py-2.5 text-xs shadow-sm ${
                          isAgent
                            ? 'bg-emerald-600 text-white rounded-br-none'
                            : 'bg-card text-foreground border border-border rounded-bl-none'
                        }`}
                      >
                        {!isAgent && (
                          <div className="text-[10px] font-semibold text-emerald-500 mb-0.5">
                            {activeConversation.contact_name}
                          </div>
                        )}
                        <p className="whitespace-pre-wrap leading-relaxed">{msg.body}</p>
                        <div
                          className={`mt-1 flex items-center justify-end gap-1 text-[9px] ${
                            isAgent ? 'text-emerald-100' : 'text-muted-foreground'
                          }`}
                        >
                          <span>{formatMessageTime(msg.sent_at)}</span>
                          {isAgent && (
                            <span>
                              {msg.delivery_status === 'sent' && <Check className="h-3 w-3 inline" />}
                              {(msg.delivery_status === 'delivered' || msg.delivery_status === 'read') && (
                                <CheckCheck className="h-3 w-3 inline" />
                              )}
                            </span>
                          )}
                        </div>
                      </div>
                    </div>
                  );
                })
              )}
              <div ref={messagesEndRef} />
            </div>

            {/* Outgoing Message Composer */}
            <form onSubmit={handleSendMessage} className="p-3.5 border-t border-border bg-card">
              <div className="flex items-center gap-2">
                <input
                  type="text"
                  placeholder="Type a WhatsApp message..."
                  value={outgoingText}
                  onChange={(e) => setOutgoingText(e.target.value)}
                  className="flex-1 rounded-xl border border-border bg-background px-4 py-2.5 text-xs text-foreground focus:border-emerald-500 focus:outline-none"
                />
                <button
                  type="submit"
                  disabled={!outgoingText.trim() || sendMessageMutation.isPending}
                  className="inline-flex items-center justify-center rounded-xl bg-emerald-600 p-2.5 text-white hover:bg-emerald-700 disabled:opacity-50 transition shadow-sm"
                >
                  {sendMessageMutation.isPending ? (
                    <Loader2 className="h-4 w-4 animate-spin" />
                  ) : (
                    <Send className="h-4 w-4" />
                  )}
                </button>
              </div>
            </form>
          </div>
        ) : (
          <div className="flex-1 flex flex-col items-center justify-center p-8 text-center text-muted-foreground">
            <MessageSquare className="h-12 w-12 opacity-20 mb-3" />
            <h3 className="text-sm font-semibold text-foreground">Select a conversation</h3>
            <p className="text-xs text-muted-foreground mt-1 max-w-sm">
              Choose a customer thread from the left panel to review message history and respond in real-time.
            </p>
          </div>
        )}

        {/* ================================================================= */}
        {/* PANEL 3: RIGHT PANEL - CONTACT INFO & INTERNAL PRIVATE NOTES     */}
        {/* ================================================================= */}
        {activeConversation && (
          <div className="w-72 sm:w-80 border-l border-border flex flex-col bg-background/50">
            {/* Tab switch */}
            <div className="flex border-b border-border bg-card">
              <button
                onClick={() => setActiveRightTab('info')}
                className={`flex-1 py-3 text-xs font-semibold border-b-2 transition flex items-center justify-center gap-1.5 ${
                  activeRightTab === 'info'
                    ? 'border-emerald-500 text-foreground'
                    : 'border-transparent text-muted-foreground hover:text-foreground'
                }`}
              >
                <User className="h-3.5 w-3.5" />
                Contact Info
              </button>
              <button
                onClick={() => setActiveRightTab('notes')}
                className={`flex-1 py-3 text-xs font-semibold border-b-2 transition flex items-center justify-center gap-1.5 ${
                  activeRightTab === 'notes'
                    ? 'border-emerald-500 text-foreground'
                    : 'border-transparent text-muted-foreground hover:text-foreground'
                }`}
              >
                <Lock className="h-3.5 w-3.5 text-amber-500" />
                Internal Notes
              </button>
            </div>

            {/* Tab Content: Info */}
            {activeRightTab === 'info' && (
              <div className="flex-1 overflow-y-auto p-4 space-y-5">
                {/* Contact Card */}
                <div className="text-center pb-4 border-b border-border">
                  <div className="h-16 w-16 rounded-full bg-emerald-500/20 text-emerald-600 flex items-center justify-center font-bold text-xl mx-auto mb-2">
                    {activeConversation.contact_name?.charAt(0).toUpperCase() || 'C'}
                  </div>
                  <h4 className="font-bold text-sm text-foreground">{activeConversation.contact_name}</h4>
                  <div className="text-xs text-muted-foreground font-mono mt-0.5">
                    {activeConversation.contact_phone}
                  </div>
                </div>

                {/* Details */}
                <div className="space-y-3 text-xs">
                  <div>
                    <span className="text-[11px] font-semibold uppercase text-muted-foreground block mb-1">
                      WhatsApp Channel
                    </span>
                    <span className="font-medium text-foreground">{activeConversation.channel_name}</span>
                  </div>

                  <div>
                    <span className="text-[11px] font-semibold uppercase text-muted-foreground block mb-1">
                      Assigned Agent
                    </span>
                    <span className="font-medium text-foreground">
                      {activeConversation.assigned_agent_name || 'Unassigned'}
                    </span>
                  </div>

                  <div>
                    <span className="text-[11px] font-semibold uppercase text-muted-foreground block mb-1">
                      Tags
                    </span>
                    <div className="flex flex-wrap gap-1 mt-1">
                      {activeConversation.tags?.length > 0 ? (
                        activeConversation.tags.map((t) => (
                          <span
                            key={t.id}
                            className="text-[10px] font-medium px-2 py-0.5 rounded-full text-white"
                            style={{ backgroundColor: t.color_hex }}
                          >
                            {t.name}
                          </span>
                        ))
                      ) : (
                        <span className="text-muted-foreground text-[11px] italic">No tags assigned</span>
                      )}
                    </div>
                  </div>
                </div>
              </div>
            )}

            {/* Tab Content: Internal Notes */}
            {activeRightTab === 'notes' && (
              <div className="flex-1 flex flex-col overflow-hidden">
                {/* Security Notice */}
                <div className="p-3 bg-amber-500/10 border-b border-amber-500/20 text-amber-600 text-[11px] flex items-start gap-2">
                  <ShieldCheck className="h-4 w-4 shrink-0 mt-0.5" />
                  <span>
                    Team notes are strictly confidential and <strong>never transmitted to WhatsApp</strong>.
                  </span>
                </div>

                {/* Notes List */}
                <div className="flex-1 overflow-y-auto p-4 space-y-3">
                  {notes.length === 0 ? (
                    <div className="text-center py-8 text-muted-foreground text-xs">
                      <FileText className="h-6 w-6 mx-auto opacity-30 mb-1" />
                      No private team notes yet
                    </div>
                  ) : (
                    notes.map((note) => (
                      <div
                        key={note.id}
                        className="p-3 rounded-xl border border-border bg-card text-xs space-y-1 shadow-xs"
                      >
                        <div className="flex items-center justify-between text-[10px] text-muted-foreground">
                          <span className="font-semibold text-foreground">{note.user_name}</span>
                          <span>{new Date(note.created_at).toLocaleDateString()}</span>
                        </div>
                        <p className="text-foreground/90 whitespace-pre-wrap">{note.content}</p>
                      </div>
                    ))
                  )}
                </div>

                {/* Add Note Form */}
                <form onSubmit={handleCreateNote} className="p-3 border-t border-border bg-card space-y-2">
                  <textarea
                    rows={2}
                    placeholder="Add an internal note..."
                    value={newNoteText}
                    onChange={(e) => setNewNoteText(e.target.value)}
                    className="w-full rounded-lg border border-border bg-background p-2 text-xs text-foreground focus:border-amber-500 focus:outline-none resize-none"
                  />
                  <button
                    type="submit"
                    disabled={!newNoteText.trim() || createNoteMutation.isPending}
                    className="w-full rounded-lg bg-amber-600 py-1.5 text-xs font-semibold text-white hover:bg-amber-700 disabled:opacity-50 transition"
                  >
                    {createNoteMutation.isPending ? 'Saving...' : 'Add Private Note'}
                  </button>
                </form>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
};
