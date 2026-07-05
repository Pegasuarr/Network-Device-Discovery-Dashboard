import React, { createContext, useContext, useState, useEffect } from "react";
import { useAuth } from "./AuthContext";

interface SocketContextType {
  lastMessage: any;
  isConnected: boolean;
}

const SocketContext = createContext<SocketContextType | null>(null);

export const SocketProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { user } = useAuth();
  const [isConnected, setIsConnected] = useState(false);
  const [lastMessage, setLastMessage] = useState<any>(null);

  useEffect(() => {
    if (!user) {
      setIsConnected(false);
      setLastMessage(null);
      return;
    }

    const token = localStorage.getItem("accessToken");
    if (!token) return;

    const wsProtocol = window.location.protocol === "https:" ? "wss" : "ws";
    // Pointing directly to our backend server port
    const socket = new WebSocket(`${wsProtocol}://localhost:8080/ws?token=${token}`);

    socket.onopen = () => {
      setIsConnected(true);
    };

    socket.onmessage = (event) => {
      try {
        const parsed = JSON.parse(event.data);
        setLastMessage(parsed);
      } catch (err) {
        console.error("Failed to parse websocket message", err);
      }
    };

    socket.onclose = () => {
      setIsConnected(false);
      // Reconnect after 3 seconds
      setTimeout(() => {
        setIsConnected(false);
      }, 3000);
    };

    socket.onerror = () => {
      setIsConnected(false);
    };

    return () => {
      socket.close();
    };
  }, [user]);

  return (
    <SocketContext.Provider value={{ lastMessage, isConnected }}>
      {children}
    </SocketContext.Provider>
  );
};

export const useSocket = () => {
  const context = useContext(SocketContext);
  if (!context) {
    throw new Error("useSocket must be used within a SocketProvider");
  }
  return context;
};
