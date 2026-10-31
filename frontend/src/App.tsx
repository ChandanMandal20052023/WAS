import React from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { DashboardLayout } from './layouts/DashboardLayout';
import { DashboardPage } from './pages/dashboard/DashboardPage';
import { LoginPage } from './pages/auth/LoginPage';
import { RegisterPage } from './pages/auth/RegisterPage';
import { ProtectedRoute } from './components/common/ProtectedRoute';

import { WhatsAppChannelsPage } from './pages/channels/WhatsAppChannelsPage';
import { ContactsPage } from './pages/contacts/ContactsPage';
import { InboxPage } from './pages/inbox/InboxPage';
import { TeamPage } from './pages/team/TeamPage';

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
        {/* Public Authentication Routes */}
        <Route path="/login" element={<LoginPage />} />
        <Route path="/register" element={<RegisterPage />} />

        {/* Protected Dashboard Routes */}
        <Route
          path="/"
          element={
            <ProtectedRoute>
              <DashboardLayout />
            </ProtectedRoute>
          }
        >
          <Route index element={<DashboardPage />} />
          <Route path="inbox" element={<InboxPage />} />
          <Route path="contacts" element={<ContactsPage />} />
          <Route path="channels" element={<WhatsAppChannelsPage />} />
          <Route path="team" element={<TeamPage />} />
          <Route path="analytics" element={<AnalyticsPlaceholder />} />
          <Route path="settings" element={<SettingsPlaceholder />} />
        </Route>

        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </BrowserRouter>
  );
};

export default App;
