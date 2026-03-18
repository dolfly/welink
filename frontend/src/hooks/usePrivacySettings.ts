/**
 * 隐私屏蔽设置 Hook
 * 屏蔽名单持久化在 localStorage，支持按昵称/备注/微信ID匹配联系人和群聊
 */

import { useState, useCallback } from 'react';

const STORAGE_KEY_USERS = 'welink_blocked_users';
const STORAGE_KEY_GROUPS = 'welink_blocked_groups';

function loadList(key: string): string[] {
  try {
    const raw = localStorage.getItem(key);
    return raw ? JSON.parse(raw) : [];
  } catch {
    return [];
  }
}

function saveList(key: string, list: string[]) {
  localStorage.setItem(key, JSON.stringify(list));
}

export function usePrivacySettings() {
  const [blockedUsers, setBlockedUsers] = useState<string[]>(() => loadList(STORAGE_KEY_USERS));
  const [blockedGroups, setBlockedGroups] = useState<string[]>(() => loadList(STORAGE_KEY_GROUPS));

  const addBlockedUser = useCallback((value: string) => {
    const v = value.trim();
    if (!v) return;
    setBlockedUsers((prev) => {
      if (prev.includes(v)) return prev;
      const next = [...prev, v];
      saveList(STORAGE_KEY_USERS, next);
      return next;
    });
  }, []);

  const removeBlockedUser = useCallback((value: string) => {
    setBlockedUsers((prev) => {
      const next = prev.filter((x) => x !== value);
      saveList(STORAGE_KEY_USERS, next);
      return next;
    });
  }, []);

  const addBlockedGroup = useCallback((value: string) => {
    const v = value.trim();
    if (!v) return;
    setBlockedGroups((prev) => {
      if (prev.includes(v)) return prev;
      const next = [...prev, v];
      saveList(STORAGE_KEY_GROUPS, next);
      return next;
    });
  }, []);

  const removeBlockedGroup = useCallback((value: string) => {
    setBlockedGroups((prev) => {
      const next = prev.filter((x) => x !== value);
      saveList(STORAGE_KEY_GROUPS, next);
      return next;
    });
  }, []);

  return {
    blockedUsers,
    blockedGroups,
    addBlockedUser,
    removeBlockedUser,
    addBlockedGroup,
    removeBlockedGroup,
  };
}
