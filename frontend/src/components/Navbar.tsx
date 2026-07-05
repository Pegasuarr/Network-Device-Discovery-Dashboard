import React, { useEffect, useState } from "react";
import { useAuth } from "../context/AuthContext";
import { useSocket } from "../context/SocketContext";
import { Sun, Moon, LogOut, Activity } from "lucide-react";

export const Navbar: React.FC = () => {
  const { user, logout } = useAuth();
  const { isConnected } = useSocket();
  const [dark, setDark] = useState(true);

  useEffect(() => {
    // Force dark theme as default for NOC premium experience
    const bodyClass = window.document.body.classList;
    if (dark) {
      bodyClass.add("dark");
    } else {
      bodyClass.remove("dark");
    }
  }, [dark]);

  return (
    <header className="flex items-center justify-between px-6 py-4 bg-white border-b border-slate-200 dark:bg-darkCard dark:border-darkBorder shadow-sm h-16">
      <div className="flex items-center space-x-3">
        <Activity className="h-6 w-6 text-indigo-500 animate-pulse" />
        <span className="text-lg font-bold text-slate-800 dark:text-slate-100 tracking-wider">
          AETHER NOC
        </span>
        <span className="hidden sm:inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-slate-100 dark:bg-slate-800 text-slate-400 dark:text-slate-300">
          v1.0.0
        </span>
      </div>

      <div className="flex items-center space-x-6">
        {/* WS Stream Status indicator */}
        <div className="flex items-center space-x-2">
          <div
            className={`h-2.5 w-2.5 rounded-full ${
              isConnected ? "bg-green-500 animate-ping" : "bg-red-500"
            }`}
          />
          <span className="text-xs text-slate-500 dark:text-slate-400 font-medium">
            {isConnected ? "Live Stream" : "Connecting..."}
          </span>
        </div>

        {/* Theme toggle */}
        <button
          onClick={() => setDark(!dark)}
          className="p-1 rounded-lg text-slate-500 hover:bg-slate-100 dark:text-slate-400 dark:hover:bg-slate-800 transition-colors"
        >
          {dark ? <Sun className="h-5 w-5" /> : <Moon className="h-5 w-5" />}
        </button>

        {/* Profile Card */}
        {user && (
          <div className="flex items-center space-x-3 border-l border-slate-200 dark:border-darkBorder pl-4">
            <div className="flex flex-col text-right">
              <span className="text-sm font-semibold text-slate-700 dark:text-slate-200">
                {user.username}
              </span>
              <span className="text-xs text-indigo-500 dark:text-indigo-400 font-mono">
                {user.role?.name || "Viewer"}
              </span>
            </div>
            <button
              onClick={logout}
              className="p-1.5 rounded-lg text-red-500 hover:bg-red-50/50 dark:hover:bg-red-950/20 transition-colors"
              title="Logout"
            >
              <LogOut className="h-5 w-5" />
            </button>
          </div>
        )}
      </div>
    </header>
  );
};
