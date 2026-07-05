import React from "react";
import { NavLink } from "react-router-dom";
import { useAuth } from "../context/AuthContext";
import {
  LayoutDashboard,
  HardDrive,
  Activity,
  AlertTriangle,
  Users,
  Settings,
  ShieldCheck,
} from "lucide-react";

export const Sidebar: React.FC = () => {
  const { user } = useAuth();

  const links = [
    { to: "/", label: "Dashboard", icon: LayoutDashboard },
    { to: "/devices", label: "Devices", icon: HardDrive },
    { to: "/monitoring", label: "Discovery Scanner", icon: Activity },
    { to: "/alerts", label: "Alerts & Rules", icon: AlertTriangle },
  ];

  // Admin-only paths
  if (user?.role?.name === "Admin") {
    links.push(
      { to: "/users", label: "Users", icon: Users },
      { to: "/settings", label: "Settings", icon: Settings }
    );
  }

  return (
    <aside className="w-64 bg-white border-r border-slate-200 dark:bg-darkCard dark:border-darkBorder flex flex-col h-[calc(100vh-4rem)]">
      <nav className="flex-1 px-4 py-6 space-y-2 overflow-y-auto">
        {links.map((link) => {
          const Icon = link.icon;
          return (
            <NavLink
              key={link.to}
              to={link.to}
              className={({ isActive }) =>
                `flex items-center px-4 py-3 rounded-lg text-sm font-semibold transition-all duration-200 ${
                  isActive
                    ? "bg-indigo-50 text-indigo-600 dark:bg-indigo-950/40 dark:text-indigo-400 border-l-4 border-indigo-600 dark:border-indigo-400"
                    : "text-slate-600 hover:bg-slate-100 hover:text-slate-900 dark:text-slate-400 dark:hover:bg-slate-800/60 dark:hover:text-slate-200"
                }`
              }
            >
              <Icon className="h-5 w-5 mr-3" />
              {link.label}
            </NavLink>
          );
        })}
      </nav>

      {/* Footer Info */}
      <div className="p-4 border-t border-slate-200 dark:border-darkBorder bg-slate-50/50 dark:bg-slate-900/10">
        <div className="flex items-center space-x-2 text-xs text-slate-500 dark:text-slate-400">
          <ShieldCheck className="h-4 w-4 text-emerald-500" />
          <span>Tenant: {user?.role?.id === 1 ? "System NOC Tenant" : "Global Enterprise"}</span>
        </div>
      </div>
    </aside>
  );
};
