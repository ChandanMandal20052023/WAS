import React, { useState, useEffect, useRef } from 'react';
import { WhatsAppChannel } from '../../types/channels';
import { channelService } from '../../services/channelService';
import {
  X,
  QrCode,
  Smartphone,
  RefreshCw,
  CheckCircle2,
  AlertCircle,
  Loader2,
  Copy,
  Check,
} from 'lucide-react';

interface ConnectChannelModalProps {
  channel: WhatsAppChannel | null;
  isOpen: boolean;
  onClose: () => void;
  onSuccess: () => void;
}

export const ConnectChannelModal: React.FC<ConnectChannelModalProps> = ({
  channel,
  isOpen,
  onClose,
  onSuccess,
}) => {
  const [activeTab, setActiveTab] = useState<'qr' | 'code'>('qr');
  const [qrCode, setQrCode] = useState<string | null>(null);
  const [isLoadingQR, setIsLoadingQR] = useState(false);
  const [countdown, setCountdown] = useState(60);
  const [pairingPhone, setPairingPhone] = useState('');
  const [pairingCode, setPairingCode] = useState<string | null>(null);
  const [isPairingLoading, setIsPairingLoading] = useState(false);
  const [isConnected, setIsConnected] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [copiedCode, setCopiedCode] = useState(false);

  const pollIntervalRef = useRef<NodeJS.Timeout | null>(null);

  // Load QR code when modal opens
  const fetchQRCode = async () => {
    if (!channel) return;
    try {
      setIsLoadingQR(true);
      setErrorMessage(null);
      const res = await channelService.getQRCode(channel.id);
      if (res.status === 'Connected') {
        setIsConnected(true);
        setTimeout(onSuccess, 1500);
      } else {
        setQrCode(res.qrcode);
        setCountdown(res.expires_in || 60);
      }
    } catch (err: unknown) {
      if (err instanceof Error) {
        setErrorMessage(err.message);
      } else {
        setErrorMessage('Failed to load QR code. Please try again.');
      }
    } finally {
      setIsLoadingQR(false);
    }
  };

  useEffect(() => {
    if (isOpen && channel) {
      setIsConnected(channel.status === 'Connected');
      setPairingCode(null);
      fetchQRCode();

      // Start status polling
      pollIntervalRef.current = setInterval(async () => {
        try {
          const statusRes = await channelService.syncStatus(channel.id);
          if (statusRes.connected || statusRes.status === 'Connected') {
            setIsConnected(true);
            if (pollIntervalRef.current) clearInterval(pollIntervalRef.current);
            setTimeout(() => {
              onSuccess();
              onClose();
            }, 1200);
          }
        } catch {
          // Ignore transient polling errors
        }
      }, 3000);
    }

    return () => {
      if (pollIntervalRef.current) {
        clearInterval(pollIntervalRef.current);
      }
    };
  }, [isOpen, channel]);

  // Countdown timer for QR expiration
  useEffect(() => {
    if (!isOpen || isConnected) return;
    const timer = setInterval(() => {
      setCountdown((prev) => {
        if (prev <= 1) {
          fetchQRCode();
          return 60;
        }
        return prev - 1;
      });
    }, 1000);

    return () => clearInterval(timer);
  }, [isOpen, isConnected, channel]);

  const handlePairPhone = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!channel || !pairingPhone) return;
    try {
      setIsPairingLoading(true);
      setErrorMessage(null);
      const res = await channelService.pairPhone(channel.id, pairingPhone);
      setPairingCode(res.pairing_code);
    } catch (err: unknown) {
      if (err instanceof Error) {
        setErrorMessage(err.message);
      } else {
        setErrorMessage('Failed to request pairing code');
      }
    } finally {
      setIsPairingLoading(false);
    }
  };

  const handleCopyCode = () => {
    if (pairingCode) {
      navigator.clipboard.writeText(pairingCode);
      setCopiedCode(true);
      setTimeout(() => setCopiedCode(false), 2000);
    }
  };

  if (!isOpen || !channel) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4">
      <div className="w-full max-w-lg bg-card border border-border rounded-2xl shadow-2xl overflow-hidden flex flex-col">
        {/* Modal Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-border">
          <div>
            <h2 className="text-lg font-bold text-foreground">Connect WhatsApp Account</h2>
            <p className="text-xs text-muted-foreground">{channel.name}</p>
          </div>
          <button
            onClick={onClose}
            className="text-muted-foreground hover:text-foreground p-1 rounded-lg transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Success Banner */}
        {isConnected ? (
          <div className="p-10 flex flex-col items-center justify-center space-y-3 text-center">
            <div className="w-16 h-16 rounded-full bg-emerald-500/10 text-emerald-500 flex items-center justify-center">
              <CheckCircle2 className="w-10 h-10 animate-bounce" />
            </div>
            <h3 className="text-xl font-bold text-foreground">WhatsApp Linked Successfully!</h3>
            <p className="text-sm text-muted-foreground max-w-sm">
              Your WhatsApp channel is now connected and ready to receive customer conversations.
            </p>
          </div>
        ) : (
          <div className="p-6 space-y-5">
            {/* Error Message */}
            {errorMessage && (
              <div className="flex items-center gap-2 p-3 text-xs text-destructive bg-destructive/10 border border-destructive/20 rounded-xl">
                <AlertCircle className="w-4 h-4 flex-shrink-0" />
                <span>{errorMessage}</span>
              </div>
            )}

            {/* Mode Switch Tabs */}
            <div className="flex bg-muted p-1 rounded-xl">
              <button
                onClick={() => setActiveTab('qr')}
                className={`flex-1 flex items-center justify-center gap-2 py-2 text-xs font-semibold rounded-lg transition-all ${
                  activeTab === 'qr'
                    ? 'bg-card text-foreground shadow-sm'
                    : 'text-muted-foreground hover:text-foreground'
                }`}
              >
                <QrCode className="w-4 h-4" />
                <span>Scan QR Code</span>
              </button>
              <button
                onClick={() => setActiveTab('code')}
                className={`flex-1 flex items-center justify-center gap-2 py-2 text-xs font-semibold rounded-lg transition-all ${
                  activeTab === 'code'
                    ? 'bg-card text-foreground shadow-sm'
                    : 'text-muted-foreground hover:text-foreground'
                }`}
              >
                <Smartphone className="w-4 h-4" />
                <span>Link with Phone Code</span>
              </button>
            </div>

            {/* QR Code Tab View */}
            {activeTab === 'qr' && (
              <div className="space-y-4">
                <div className="flex flex-col items-center justify-center p-6 bg-muted/30 border border-border rounded-xl">
                  {isLoadingQR ? (
                    <div className="h-56 w-56 flex flex-col items-center justify-center gap-2">
                      <Loader2 className="w-8 h-8 text-whatsapp animate-spin" />
                      <span className="text-xs text-muted-foreground font-medium">Generating QR...</span>
                    </div>
                  ) : qrCode ? (
                    <div className="relative group">
                      <img
                        src={qrCode}
                        alt="WhatsApp QR Code"
                        className="w-56 h-56 rounded-xl shadow-sm border border-border bg-white p-2"
                      />
                      <div className="mt-2 text-center">
                        <span className="text-[11px] text-muted-foreground">
                          Auto-refreshing in <strong className="text-foreground">{countdown}s</strong>
                        </span>
                      </div>
                    </div>
                  ) : (
                    <div className="h-56 w-56 flex flex-col items-center justify-center gap-3 text-center">
                      <AlertCircle className="w-8 h-8 text-muted-foreground" />
                      <p className="text-xs text-muted-foreground">Could not load QR code</p>
                      <button
                        onClick={fetchQRCode}
                        className="px-3 py-1.5 rounded-lg text-xs bg-whatsapp text-white font-medium"
                      >
                        Try Again
                      </button>
                    </div>
                  )}
                </div>

                {/* Instructions */}
                <div className="text-xs text-muted-foreground space-y-1.5 bg-muted/20 p-3 rounded-xl border border-border">
                  <p className="font-semibold text-foreground">How to link:</p>
                  <ol className="list-decimal list-inside space-y-1">
                    <li>Open WhatsApp on your mobile phone</li>
                    <li>Tap <strong>Settings</strong> or <strong>Menu (⋮)</strong> ➔ <strong>Linked Devices</strong></li>
                    <li>Tap <strong>Link a Device</strong> and point your camera at this QR code</li>
                  </ol>
                </div>

                <div className="flex items-center justify-between pt-1">
                  <span className="text-xs text-muted-foreground flex items-center gap-1.5">
                    <span className="w-2 h-2 rounded-full bg-whatsapp animate-pulse"></span>
                    Listening for scan...
                  </span>
                  <button
                    onClick={fetchQRCode}
                    disabled={isLoadingQR}
                    className="inline-flex items-center gap-1 text-xs text-primary font-medium hover:underline"
                  >
                    <RefreshCw className={`w-3.5 h-3.5 ${isLoadingQR ? 'animate-spin' : ''}`} />
                    <span>Refresh QR</span>
                  </button>
                </div>
              </div>
            )}

            {/* Phone Code Tab View */}
            {activeTab === 'code' && (
              <div className="space-y-4">
                <form onSubmit={handlePairPhone} className="space-y-3">
                  <div className="space-y-1">
                    <label className="text-xs font-semibold text-foreground">WhatsApp Phone Number</label>
                    <input
                      type="tel"
                      placeholder="+1 (555) 000-0000"
                      value={pairingPhone}
                      onChange={(e) => setPairingPhone(e.target.value)}
                      className="w-full bg-background border border-input px-3 py-2 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary/50"
                      required
                    />
                    <p className="text-[11px] text-muted-foreground">
                      Include country code (e.g., +1 for US, +44 for UK, +91 for India).
                    </p>
                  </div>

                  <button
                    type="submit"
                    disabled={isPairingLoading}
                    className="w-full bg-whatsapp hover:bg-whatsapp-dark text-white font-medium py-2 rounded-xl text-xs flex items-center justify-center gap-1.5 transition-all shadow-sm disabled:opacity-50"
                  >
                    {isPairingLoading ? (
                      <Loader2 className="w-4 h-4 animate-spin" />
                    ) : (
                      <span>Request 8-Digit Code</span>
                    )}
                  </button>
                </form>

                {pairingCode && (
                  <div className="p-4 bg-whatsapp/10 border border-whatsapp/20 rounded-xl text-center space-y-2">
                    <span className="text-xs text-muted-foreground">Enter this code on your phone:</span>
                    <div className="flex items-center justify-center gap-2">
                      <span className="font-mono text-xl font-bold tracking-widest text-whatsapp">
                        {pairingCode}
                      </span>
                      <button
                        onClick={handleCopyCode}
                        className="p-1.5 rounded-lg hover:bg-whatsapp/20 text-whatsapp transition-colors"
                        title="Copy Code"
                      >
                        {copiedCode ? <Check className="w-4 h-4" /> : <Copy className="w-4 h-4" />}
                      </button>
                    </div>
                  </div>
                )}
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
};
