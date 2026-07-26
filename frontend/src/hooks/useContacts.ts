/**
 * 联系人数据 Hook
 */

import { useState, useEffect, useCallback, useRef } from 'react';
import { contactsApi } from '../services/api';
import type { ContactStats, WordCount } from '../types';

export const useContacts = (enabled = true, refreshInterval = 0) => {
  const [contacts, setContacts] = useState<ContactStats[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  const fetchContacts = useCallback(async () => {
    try {
      setLoading(true);
      const data = await contactsApi.getStats();
      setContacts(data ?? []);
      setError(null);
    } catch (err) {
      setError(err as Error);
      console.error('Failed to fetch contacts:', err);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    if (!enabled) return;
    void fetchContacts();

    if (refreshInterval > 0) {
      const timer = setInterval(fetchContacts, refreshInterval);
      return () => clearInterval(timer);
    }
  }, [fetchContacts, enabled, refreshInterval]);

  return {
    contacts,
    loading,
    error,
    refresh: fetchContacts,
  };
};

export const useWordCloud = () => {
  const [data, setData] = useState<WordCount[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const requestRef = useRef<AbortController | null>(null);

  const fetchWordCloud = useCallback(async (username: string, includeMine = false) => {
    requestRef.current?.abort();
    const controller = new AbortController();
    requestRef.current = controller;
    try {
      setLoading(true);
      setData([]);
      const result = await contactsApi.getWordCloud(username, includeMine, controller.signal);
      if (controller.signal.aborted) return;
      setData(result || []);
      setError(null);
    } catch (err) {
      if (controller.signal.aborted) return;
      setError(err as Error);
      console.error('Failed to fetch word cloud:', err);
    } finally {
      if (!controller.signal.aborted && requestRef.current === controller) {
        setLoading(false);
      }
    }
  }, []);

  useEffect(() => () => requestRef.current?.abort(), []);

  return {
    data,
    loading,
    error,
    fetch: fetchWordCloud,
  };
};
