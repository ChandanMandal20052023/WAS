import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import {
  Users,
  UserPlus,
  Search,
  Tag as TagIcon,
  Phone,
  Mail,
  MessageSquare,
  Edit2,
  Trash2,
  Loader2,
  X,
  AlertCircle,
} from 'lucide-react';
import { contactService } from '../../services/contactService';
import { ContactWithTags, CreateContactInput } from '../../types/contacts';

export const ContactsPage: React.FC = () => {
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const [search, setSearch] = useState('');
  const [selectedTagId, setSelectedTagId] = useState<string | undefined>(undefined);
  const [isAddModalOpen, setIsAddModalOpen] = useState(false);
  const [editingContact, setEditingContact] = useState<ContactWithTags | null>(null);

  // Form states for create/edit
  const [phone, setPhone] = useState('');
  const [fullName, setFullName] = useState('');
  const [email, setEmail] = useState('');
  const [selectedTagIds, setSelectedTagIds] = useState<string[]>([]);
  const [formError, setFormError] = useState<string | null>(null);

  // New tag state
  const [isNewTagModalOpen, setIsNewTagModalOpen] = useState(false);
  const [newTagName, setNewTagName] = useState('');
  const [newTagColor, setNewTagColor] = useState('#128C7E');

  const { data: contactsData, isLoading: isLoadingContacts } = useQuery({
    queryKey: ['contacts', search, selectedTagId],
    queryFn: () =>
      contactService.getContacts({
        search: search || undefined,
        tag_id: selectedTagId,
        limit: 50,
      }),
  });

  const { data: tagsData = [] } = useQuery({
    queryKey: ['tags'],
    queryFn: () => contactService.getTags(),
  });

  const createContactMutation = useMutation({
    mutationFn: (input: CreateContactInput) => contactService.createContact(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['contacts'] });
      setIsAddModalOpen(false);
      resetForm();
    },
    onError: (err: any) => {
      setFormError(err.response?.data?.error || 'Failed to save contact');
    },
  });

  const updateContactMutation = useMutation({
    mutationFn: ({ id, input }: { id: string; input: any }) =>
      contactService.updateContact(id, input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['contacts'] });
      setEditingContact(null);
      resetForm();
    },
    onError: (err: any) => {
      setFormError(err.response?.data?.error || 'Failed to update contact');
    },
  });

  const deleteContactMutation = useMutation({
    mutationFn: (id: string) => contactService.deleteContact(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['contacts'] });
    },
  });

  const createTagMutation = useMutation({
    mutationFn: (input: { name: string; color_hex?: string }) => contactService.createTag(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tags'] });
      setIsNewTagModalOpen(false);
      setNewTagName('');
    },
  });

  const resetForm = () => {
    setPhone('');
    setFullName('');
    setEmail('');
    setSelectedTagIds([]);
    setFormError(null);
  };

  const openAddModal = () => {
    resetForm();
    setIsAddModalOpen(true);
  };

  const openEditModal = (c: ContactWithTags) => {
    setEditingContact(c);
    setPhone(c.phone_number);
    setFullName(c.full_name);
    setEmail(c.email || '');
    setSelectedTagIds(c.tags.map((t) => t.id));
    setFormError(null);
  };

  const handleSaveContact = (e: React.FormEvent) => {
    e.preventDefault();
    setFormError(null);

    if (editingContact) {
      updateContactMutation.mutate({
        id: editingContact.id,
        input: {
          full_name: fullName,
          email: email || undefined,
          tag_ids: selectedTagIds,
        },
      });
    } else {
      createContactMutation.mutate({
        phone_number: phone,
        full_name: fullName,
        email: email || undefined,
        tag_ids: selectedTagIds,
      });
    }
  };

  const toggleTagSelection = (tagId: string) => {
    setSelectedTagIds((prev) =>
      prev.includes(tagId) ? prev.filter((id) => id !== tagId) : [...prev, tagId]
    );
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-foreground">Contact Directory</h1>
          <p className="text-sm text-muted-foreground">
            Manage your organization's customer profiles, normalized E.164 phone numbers, and tags.
          </p>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={() => setIsNewTagModalOpen(true)}
            className="inline-flex items-center gap-1.5 rounded-lg border border-border bg-card px-3.5 py-2 text-sm font-medium text-foreground hover:bg-accent transition"
          >
            <TagIcon className="h-4 w-4" />
            New Tag
          </button>
          <button
            onClick={openAddModal}
            className="inline-flex items-center gap-1.5 rounded-lg bg-emerald-600 px-4 py-2 text-sm font-medium text-white hover:bg-emerald-700 transition"
          >
            <UserPlus className="h-4 w-4" />
            Add Contact
          </button>
        </div>
      </div>

      {/* Filters & Search */}
      <div className="flex flex-col sm:flex-row gap-3 items-center justify-between bg-card p-4 rounded-xl border border-border">
        <div className="relative w-full sm:w-96">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <input
            type="text"
            placeholder="Search by name, phone or email..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full rounded-lg border border-border bg-background pl-9 pr-4 py-2 text-sm text-foreground focus:border-emerald-500 focus:outline-none focus:ring-1 focus:ring-emerald-500"
          />
        </div>

        {/* Tags filter pills */}
        <div className="flex flex-wrap gap-1.5 items-center w-full sm:w-auto">
          <button
            onClick={() => setSelectedTagId(undefined)}
            className={`px-3 py-1 rounded-full text-xs font-medium transition ${
              selectedTagId === undefined
                ? 'bg-emerald-600 text-white'
                : 'bg-muted text-muted-foreground hover:bg-accent'
            }`}
          >
            All Tags
          </button>
          {tagsData.map((t) => (
            <button
              key={t.id}
              onClick={() => setSelectedTagId(selectedTagId === t.id ? undefined : t.id)}
              className={`px-3 py-1 rounded-full text-xs font-medium transition ${
                selectedTagId === t.id
                  ? 'ring-2 ring-emerald-500 text-white'
                  : 'hover:opacity-80'
              }`}
              style={{
                backgroundColor: t.color_hex,
                color: '#fff',
              }}
            >
              {t.name}
            </button>
          ))}
        </div>
      </div>

      {/* Contacts List */}
      <div className="bg-card rounded-xl border border-border overflow-hidden">
        {isLoadingContacts ? (
          <div className="flex items-center justify-center py-16 text-muted-foreground">
            <Loader2 className="h-6 w-6 animate-spin mr-2" />
            Loading contacts...
          </div>
        ) : contactsData?.contacts.length === 0 ? (
          <div className="text-center py-16 px-4">
            <Users className="mx-auto h-12 w-12 text-muted-foreground/50 mb-3" />
            <h3 className="text-base font-semibold text-foreground">No contacts found</h3>
            <p className="text-sm text-muted-foreground mt-1 max-w-sm mx-auto">
              Create your first customer contact or wait for incoming WhatsApp messages to auto-populate.
            </p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-sm">
              <thead className="bg-muted/50 text-xs uppercase text-muted-foreground border-b border-border">
                <tr>
                  <th className="px-6 py-3.5">Customer</th>
                  <th className="px-6 py-3.5">Phone (E.164)</th>
                  <th className="px-6 py-3.5">Email</th>
                  <th className="px-6 py-3.5">Tags</th>
                  <th className="px-6 py-3.5 text-right">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {contactsData?.contacts.map((contact) => (
                  <tr key={contact.id} className="hover:bg-muted/30 transition">
                    <td className="px-6 py-4">
                      <div className="flex items-center gap-3">
                        <div className="h-9 w-9 rounded-full bg-emerald-500/10 text-emerald-600 flex items-center justify-center font-semibold text-sm">
                          {contact.full_name?.charAt(0).toUpperCase() || 'C'}
                        </div>
                        <div>
                          <div className="font-medium text-foreground">{contact.full_name}</div>
                          <div className="text-xs text-muted-foreground">
                            Added {new Date(contact.created_at).toLocaleDateString()}
                          </div>
                        </div>
                      </div>
                    </td>
                    <td className="px-6 py-4 font-mono text-xs text-foreground">
                      <div className="flex items-center gap-1.5">
                        <Phone className="h-3.5 w-3.5 text-emerald-500" />
                        {contact.phone_number}
                      </div>
                    </td>
                    <td className="px-6 py-4 text-muted-foreground">
                      {contact.email ? (
                        <div className="flex items-center gap-1.5">
                          <Mail className="h-3.5 w-3.5" />
                          {contact.email}
                        </div>
                      ) : (
                        <span className="text-xs italic">—</span>
                      )}
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex flex-wrap gap-1">
                        {contact.tags.map((t) => (
                          <span
                            key={t.id}
                            className="text-[11px] font-medium px-2 py-0.5 rounded-full text-white"
                            style={{ backgroundColor: t.color_hex }}
                          >
                            {t.name}
                          </span>
                        ))}
                      </div>
                    </td>
                    <td className="px-6 py-4 text-right">
                      <div className="flex items-center justify-end gap-1.5">
                        <button
                          onClick={() => navigate('/inbox')}
                          title="Open WhatsApp Chat"
                          className="p-1.5 text-muted-foreground hover:text-emerald-500 hover:bg-muted rounded-lg transition"
                        >
                          <MessageSquare className="h-4 w-4" />
                        </button>
                        <button
                          onClick={() => openEditModal(contact)}
                          title="Edit Profile"
                          className="p-1.5 text-muted-foreground hover:text-blue-500 hover:bg-muted rounded-lg transition"
                        >
                          <Edit2 className="h-4 w-4" />
                        </button>
                        <button
                          onClick={() => {
                            if (confirm(`Delete contact ${contact.full_name}?`)) {
                              deleteContactMutation.mutate(contact.id);
                            }
                          }}
                          title="Delete Contact"
                          className="p-1.5 text-muted-foreground hover:text-rose-500 hover:bg-muted rounded-lg transition"
                        >
                          <Trash2 className="h-4 w-4" />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Create / Edit Contact Modal */}
      {(isAddModalOpen || editingContact) && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4 backdrop-blur-sm">
          <div className="w-full max-w-md rounded-2xl border border-border bg-card p-6 shadow-2xl">
            <div className="flex items-center justify-between pb-4 border-b border-border">
              <h3 className="text-lg font-bold text-foreground">
                {editingContact ? 'Edit Contact Profile' : 'Add New Customer Contact'}
              </h3>
              <button
                onClick={() => {
                  setIsAddModalOpen(false);
                  setEditingContact(null);
                }}
                className="text-muted-foreground hover:text-foreground"
              >
                <X className="h-5 w-5" />
              </button>
            </div>

            {formError && (
              <div className="mt-4 flex items-center gap-2 rounded-lg bg-rose-500/10 p-3 text-sm text-rose-500 border border-rose-500/20">
                <AlertCircle className="h-4 w-4 shrink-0" />
                <span>{formError}</span>
              </div>
            )}

            <form onSubmit={handleSaveContact} className="mt-4 space-y-4">
              <div>
                <label className="block text-xs font-semibold uppercase text-muted-foreground">
                  Phone Number (E.164)
                </label>
                <input
                  type="text"
                  required
                  disabled={!!editingContact}
                  placeholder="+14155552671"
                  value={phone}
                  onChange={(e) => setPhone(e.target.value)}
                  className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2 text-sm text-foreground focus:border-emerald-500 focus:outline-none disabled:opacity-60"
                />
                <span className="text-[11px] text-muted-foreground">
                  Format with country code (e.g. +1..., +91..., +44...)
                </span>
              </div>

              <div>
                <label className="block text-xs font-semibold uppercase text-muted-foreground">
                  Full Name
                </label>
                <input
                  type="text"
                  required
                  placeholder="John Doe"
                  value={fullName}
                  onChange={(e) => setFullName(e.target.value)}
                  className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2 text-sm text-foreground focus:border-emerald-500 focus:outline-none"
                />
              </div>

              <div>
                <label className="block text-xs font-semibold uppercase text-muted-foreground">
                  Email Address (Optional)
                </label>
                <input
                  type="email"
                  placeholder="john@example.com"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2 text-sm text-foreground focus:border-emerald-500 focus:outline-none"
                />
              </div>

              <div>
                <label className="block text-xs font-semibold uppercase text-muted-foreground mb-2">
                  Assign Tags
                </label>
                <div className="flex flex-wrap gap-1.5">
                  {tagsData.map((t) => {
                    const isSelected = selectedTagIds.includes(t.id);
                    return (
                      <button
                        type="button"
                        key={t.id}
                        onClick={() => toggleTagSelection(t.id)}
                        className={`px-2.5 py-1 rounded-full text-xs font-medium transition ${
                          isSelected ? 'ring-2 ring-emerald-500' : 'opacity-60 hover:opacity-100'
                        }`}
                        style={{ backgroundColor: t.color_hex, color: '#fff' }}
                      >
                        {t.name}
                      </button>
                    );
                  })}
                </div>
              </div>

              <div className="flex items-center justify-end gap-3 pt-4 border-t border-border">
                <button
                  type="button"
                  onClick={() => {
                    setIsAddModalOpen(false);
                    setEditingContact(null);
                  }}
                  className="rounded-lg border border-border px-4 py-2 text-sm font-medium text-foreground hover:bg-muted"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={createContactMutation.isPending || updateContactMutation.isPending}
                  className="inline-flex items-center gap-1.5 rounded-lg bg-emerald-600 px-4 py-2 text-sm font-medium text-white hover:bg-emerald-700 disabled:opacity-50"
                >
                  {(createContactMutation.isPending || updateContactMutation.isPending) && (
                    <Loader2 className="h-4 w-4 animate-spin" />
                  )}
                  {editingContact ? 'Save Changes' : 'Create Contact'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Create New Tag Modal */}
      {isNewTagModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4 backdrop-blur-sm">
          <div className="w-full max-w-sm rounded-2xl border border-border bg-card p-6 shadow-2xl">
            <div className="flex items-center justify-between pb-3 border-b border-border">
              <h3 className="text-base font-bold text-foreground">Create Customer Tag</h3>
              <button
                onClick={() => setIsNewTagModalOpen(false)}
                className="text-muted-foreground hover:text-foreground"
              >
                <X className="h-5 w-5" />
              </button>
            </div>

            <div className="mt-4 space-y-4">
              <div>
                <label className="block text-xs font-semibold uppercase text-muted-foreground">
                  Tag Name
                </label>
                <input
                  type="text"
                  placeholder="VIP, Lead, Enterprise..."
                  value={newTagName}
                  onChange={(e) => setNewTagName(e.target.value)}
                  className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2 text-sm text-foreground focus:border-emerald-500 focus:outline-none"
                />
              </div>

              <div>
                <label className="block text-xs font-semibold uppercase text-muted-foreground">
                  Color
                </label>
                <div className="mt-2 flex items-center gap-2">
                  <input
                    type="color"
                    value={newTagColor}
                    onChange={(e) => setNewTagColor(e.target.value)}
                    className="h-9 w-12 cursor-pointer rounded border border-border bg-background p-1"
                  />
                  <span className="font-mono text-xs text-foreground uppercase">{newTagColor}</span>
                </div>
              </div>

              <div className="flex items-center justify-end gap-2 pt-3 border-t border-border">
                <button
                  type="button"
                  onClick={() => setIsNewTagModalOpen(false)}
                  className="rounded-lg border border-border px-3.5 py-1.5 text-xs font-medium text-foreground hover:bg-muted"
                >
                  Cancel
                </button>
                <button
                  type="button"
                  disabled={!newTagName.trim() || createTagMutation.isPending}
                  onClick={() =>
                    createTagMutation.mutate({ name: newTagName, color_hex: newTagColor })
                  }
                  className="inline-flex items-center gap-1 rounded-lg bg-emerald-600 px-3.5 py-1.5 text-xs font-medium text-white hover:bg-emerald-700 disabled:opacity-50"
                >
                  {createTagMutation.isPending && <Loader2 className="h-3.5 w-3.5 animate-spin" />}
                  Save Tag
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
