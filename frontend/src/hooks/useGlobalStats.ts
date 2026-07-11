/**
 * 全局统计数据 Hook
 */

import { useState, useEffect, useCallback } from 'react';
import { globalApi } from '../services/api';
import type { GlobalStats } from '../types';

export const useGlobalStats = (enabled = true, refreshInterval = 0) => {
  const [stats, setStats] = useState<GlobalStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  const fetchStats = useCallback(async () => {
    try {
      setLoading(true);
      const data = await globalApi.getStats();
      setStats(data);
      setError(null);
    } catch (err) {
      setError(err as Error);
      console.error('Failed to fetch global stats:', err);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    if (!enabled) return;
    void fetchStats();

    if (refreshInterval > 0) {
      const timer = setInterval(fetchStats, refreshInterval);
      return () => clearInterval(timer);
    }
  }, [fetchStats, enabled, refreshInterval]);

  return {
    stats,
    loading,
    error,
    refresh: fetchStats,
  };
};
