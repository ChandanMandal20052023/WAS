import React from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { DashboardLayout } from './layouts/DashboardLayout';
import { DashboardPage } from './pages/dashboard/DashboardPage';

// Placeholder views for subsequent phases
const InboxPlaceholder: React.FC = () => (
  <div className="p-8 text-center text-muted-foreground">
    <h2 className="text-xl font-bold text-foreground">Centralized WhatsApp Inbox</h2>
    <p className="mt-2 text-sm">3-Panel Inbox with real-time customer messaging scheduled for Phase 7.</p>
  </div>
);

const ContactsPlaceholder: React.FC = () => (
  <div className="p-8 text-center text-muted-foreground">
    <h2 className="text-xl font-bold text-foreground">Contact Intelligence & Management</h2>
    <p className="mt-2 text-sm">E.164 normalized customer profiles scheduled for Phase 6.</p>
  </div>
);

const ChannelsPlaceholder: React.FC = () => (
  <div className="p-8 text-center text-muted-foreground">
    <h2 className="text-xl font-bold text-foreground">WhatsApp Channels (Evolution Go)</h2>
    <p className="mt-2 text-sm">QR Code pairing and device link manager scheduled for Phase 5.</p>
  </div>
);

const TeamPlaceholder: React.FC = () => (
  <div className="p-8 text-center text-muted-foreground">
    <h2 className="text-xl font-bold text-foreground">Team Management & Workloads</h2>
    <p className="mt-2 text-sm">Role-based member assignments scheduled for Phase 9.</p>
  </div>
);

const AnalyticsPlaceholder: React.FC = () => (
  <div className="p-8 text-center text-muted-foreground">
    <h2 className="text-xl font-bold text-foreground">Analytics & Response Metrics</h2>
    <p className="mt-2 text-sm">Conversation performance analytics scheduled for Phase 9.</p>
  </div>
);

const SettingsPlaceholder: React.FC = () => (
  <div className="p-8 text-center text-muted-foreground">
    <h2 className="text-xl font-bold text-foreground">Organization & Security Settings</h2>
    <p className="mt-2 text-sm">Tenant customization and webhook keys.</p>
  </div>
);

export const App: React.FC = () => {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<DashboardLayout />}>
          <Route index element={<DashboardPage />} />
          <Route path="inbox" element={<InboxPlaceholder />} />
          <Route path="contacts" element={<ContactsPlaceholder />} />
          <Route path="channels" element={<ChannelsPlaceholder />} />
          <Route path="team" element={<TeamPlaceholder />} />
          <Route path="analytics" element={<AnalyticsPlaceholder />} />
          <Route path="settings" element={<SettingsPlaceholder />} />
        </Route>
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </BrowserRouter>
  );
};

export default App;

