import React, { useEffect, useState } from "react";
import api from "../services/api";
import { Mail, MessageSquare, Save, Loader2, KeyRound } from "lucide-react";

export const Settings: React.FC = () => {
  const [settings, setSettings] = useState<Record<string, string>>({
    smtp_host: "",
    smtp_port: "",
    smtp_username: "",
    smtp_password: "",
    smtp_to: "",
    slack_webhook: "",
    telegram_token: "",
    telegram_chat_id: "",
    discord_webhook: "",
    api_key: "",
  });

  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    fetchSettings();
  }, []);

  const fetchSettings = async () => {
    try {
      const res = await api.get("/settings");
      setSettings((prev) => ({
        ...prev,
        ...res.data,
      }));
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    setSaving(true);
    try {
      await api.put("/settings", settings);
      alert("Integrations updated successfully!");
    } catch (err) {
      console.error(err);
      alert("Failed to save configurations.");
    } finally {
      setSaving(false);
    }
  };

  const handleGenerateApiKey = () => {
    const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";
    let key = "aether_";
    for (let i = 0; i < 32; i++) {
      key += chars.charAt(Math.floor(Math.random() * chars.length));
    }
    setSettings((prev) => ({ ...prev, api_key: key }));
  };

  if (loading) {
    return <p className="text-center p-12 text-slate-500">Querying config registry...</p>;
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold tracking-tight text-slate-800 dark:text-slate-100">
          NOC Core & Notifications Settings
        </h1>
      </div>

      <form onSubmit={handleSave} className="space-y-6 max-w-4xl">
        {/* Email Alerts */}
        <div className="p-6 bg-white dark:bg-darkCard border border-slate-200 dark:border-darkBorder rounded-xl shadow-sm space-y-4">
          <div className="flex items-center space-x-2 border-b border-slate-100 dark:border-darkBorder pb-3">
            <Mail className="h-5 w-5 text-indigo-500" />
            <h3 className="text-sm font-bold text-slate-800 dark:text-slate-250">SMTP Email Alerts</h3>
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
            <div>
              <label className="block text-xs font-bold text-slate-400 mb-1">SMTP Host</label>
              <input
                type="text"
                value={settings.smtp_host}
                onChange={(e) => setSettings({ ...settings, smtp_host: e.target.value })}
                placeholder="smtp.mailtrap.io"
                className="w-full px-3 py-2 bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-darkBorder rounded-lg text-sm"
              />
            </div>
            <div>
              <label className="block text-xs font-bold text-slate-400 mb-1">SMTP Port</label>
              <input
                type="text"
                value={settings.smtp_port}
                onChange={(e) => setSettings({ ...settings, smtp_port: e.target.value })}
                placeholder="2525"
                className="w-full px-3 py-2 bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-darkBorder rounded-lg text-sm"
              />
            </div>
            <div>
              <label className="block text-xs font-bold text-slate-400 mb-1">Alert Destination Email</label>
              <input
                type="email"
                value={settings.smtp_to}
                onChange={(e) => setSettings({ ...settings, smtp_to: e.target.value })}
                placeholder="alerts@enterprise.local"
                className="w-full px-3 py-2 bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-darkBorder rounded-lg text-sm"
              />
            </div>
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label className="block text-xs font-bold text-slate-400 mb-1">SMTP Username</label>
              <input
                type="text"
                value={settings.smtp_username}
                onChange={(e) => setSettings({ ...settings, smtp_username: e.target.value })}
                className="w-full px-3 py-2 bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-darkBorder rounded-lg text-sm"
              />
            </div>
            <div>
              <label className="block text-xs font-bold text-slate-400 mb-1">SMTP Password</label>
              <input
                type="password"
                value={settings.smtp_password}
                onChange={(e) => setSettings({ ...settings, smtp_password: e.target.value })}
                className="w-full px-3 py-2 bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-darkBorder rounded-lg text-sm"
              />
            </div>
          </div>
        </div>

        {/* Messaging Hooks */}
        <div className="p-6 bg-white dark:bg-darkCard border border-slate-200 dark:border-darkBorder rounded-xl shadow-sm space-y-4">
          <div className="flex items-center space-x-2 border-b border-slate-100 dark:border-darkBorder pb-3">
            <MessageSquare className="h-5 w-5 text-emerald-500" />
            <h3 className="text-sm font-bold text-slate-800 dark:text-slate-250">Chat & Channel Webhooks</h3>
          </div>
          <div>
            <label className="block text-xs font-bold text-slate-400 mb-1">Slack Webhook URL</label>
            <input
              type="text"
              value={settings.slack_webhook}
              onChange={(e) => setSettings({ ...settings, slack_webhook: e.target.value })}
              placeholder="https://hooks.slack.com/services/..."
              className="w-full px-3 py-2 bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-darkBorder rounded-lg text-sm font-mono"
            />
          </div>
          <div>
            <label className="block text-xs font-bold text-slate-400 mb-1">Discord Webhook URL</label>
            <input
              type="text"
              value={settings.discord_webhook}
              onChange={(e) => setSettings({ ...settings, discord_webhook: e.target.value })}
              placeholder="https://discord.com/api/webhooks/..."
              className="w-full px-3 py-2 bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-darkBorder rounded-lg text-sm font-mono"
            />
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label className="block text-xs font-bold text-slate-400 mb-1">Telegram Bot Token</label>
              <input
                type="text"
                value={settings.telegram_token}
                onChange={(e) => setSettings({ ...settings, telegram_token: e.target.value })}
                className="w-full px-3 py-2 bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-darkBorder rounded-lg text-sm font-mono"
              />
            </div>
            <div>
              <label className="block text-xs font-bold text-slate-400 mb-1">Telegram Chat ID</label>
              <input
                type="text"
                value={settings.telegram_chat_id}
                onChange={(e) => setSettings({ ...settings, telegram_chat_id: e.target.value })}
                className="w-full px-3 py-2 bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-darkBorder rounded-lg text-sm font-mono"
              />
            </div>
          </div>
        </div>

        {/* API Credentials */}
        <div className="p-6 bg-white dark:bg-darkCard border border-slate-200 dark:border-darkBorder rounded-xl shadow-sm space-y-4">
          <div className="flex items-center space-x-2 border-b border-slate-100 dark:border-darkBorder pb-3">
            <KeyRound className="h-5 w-5 text-amber-500" />
            <h3 className="text-sm font-bold text-slate-800 dark:text-slate-250">API Key Credentials</h3>
          </div>
          <p className="text-xs text-slate-400">
            Generate a personal access token for REST API commands or Prometheus scraper configurations. Use header <code className="bg-slate-100 dark:bg-slate-800 px-1 py-0.5 rounded font-mono font-bold">X-API-Key</code> on requests.
          </p>
          <div className="flex items-center space-x-3">
            <input
              type="text"
              readOnly
              value={settings.api_key}
              placeholder="No active API key generated"
              className="flex-1 px-3 py-2 bg-slate-100 dark:bg-slate-900 border border-slate-200 dark:border-darkBorder rounded-lg text-sm font-mono text-slate-800 dark:text-slate-200"
            />
            <button
              type="button"
              onClick={handleGenerateApiKey}
              className="px-4 py-2 bg-slate-150 hover:bg-slate-200 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-300 font-bold rounded-lg text-sm border border-slate-200 dark:border-darkBorder"
            >
              Generate Token
            </button>
          </div>
        </div>

        {/* Save button */}
        <div className="flex justify-end">
          <button
            type="submit"
            disabled={saving}
            className="flex items-center px-6 py-3 bg-indigo-600 hover:bg-indigo-500 active:bg-indigo-700 text-white font-bold rounded-xl shadow-lg transition-all"
          >
            {saving ? (
              <>
                <Loader2 className="h-5 w-5 mr-2 animate-spin" />
                Saving Changes...
              </>
            ) : (
              <>
                <Save className="h-5 w-5 mr-2" />
                Save Configurations
              </>
            )}
          </button>
        </div>
      </form>
    </div>
  );
};
