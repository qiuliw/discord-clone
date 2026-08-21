"use client";

import {
  createContext,
  useCallback,
  useContext,
  useMemo,
  useState,
} from "react";

import { ApiError, getMe, type User } from "@/lib/api";

type AuthContextValue = {
  user: User | null;
  loading: boolean;
  setUser: (user: User | null) => void;
  refresh: () => Promise<void>;
};

const AuthContext = createContext<AuthContextValue | null>(null);

// 全局认证提供者组件，负责管理用户认证状态和提供相关功能
export function AuthProvider({
  children,
  initialUser,
}: {
  children: React.ReactNode;
  initialUser: User | null;
}) {
  const [user, setUser] = useState<User | null>(initialUser);
  const [loading, setLoading] = useState(false);

  // 函数与组件生命周期弱相关，使用 useCallback 来创建一个刷新用户信息的函数，避免组件在每次渲染时都重新创建该函数，保证其引用的稳定性
  // 其只是缓存到 react 的全局内存中，假如 react 全局刷新了，会重新创建这个函数
  const refresh = useCallback(async () => {
    setLoading(true);
    try {
      const data = await getMe();
      setUser(data.user);
    } catch (error) {
      if (error instanceof ApiError && error.status === 401) {
        setUser(null);
        return;
      }
      throw error;
    } finally {
      setLoading(false);
    }
  }, []);

  const value = useMemo(
    () => ({ user, loading, setUser, refresh }),
    [user, loading, refresh],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

// 提供给被包裹的组件使用的 Hook，用于访问认证上下文
export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error("useAuth must be used within AuthProvider");
  }
  return ctx;
}
