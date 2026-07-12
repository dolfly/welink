import React from 'react';
import { AlertCircle, RefreshCw } from 'lucide-react';

interface LoadingStateProps {
  title?: string;
  description?: string;
  rows?: number;
}

export const LoadingState: React.FC<LoadingStateProps> = ({
  title = '正在加载数据',
  description = '请稍候，内容准备好后会自动显示。',
  rows = 5,
}) => (
  <div className="ui-card overflow-hidden" role="status" aria-live="polite" aria-busy="true">
    <div className="border-b dk-border px-5 py-4">
      <p className="text-sm font-bold dk-text">{title}</p>
      <p className="ui-caption mt-1">{description}</p>
    </div>
    <div className="divide-y dk-divide">
      {Array.from({ length: rows }, (_, index) => (
        <div key={index} className="flex items-center gap-3 px-5 py-4" aria-hidden="true">
          <div className="ui-skeleton h-10 w-10 shrink-0 rounded-xl" />
          <div className="min-w-0 flex-1 space-y-2">
            <div className="ui-skeleton h-3.5 rounded-full" style={{ width: `${42 + (index % 3) * 12}%` }} />
            <div className="ui-skeleton h-3 rounded-full w-1/3" />
          </div>
          <div className="ui-skeleton h-7 w-16 rounded-full" />
        </div>
      ))}
    </div>
  </div>
);

interface ErrorStateProps {
  title?: string;
  description?: string;
  onRetry?: () => void;
}

export const ErrorState: React.FC<ErrorStateProps> = ({
  title = '数据加载失败',
  description = '可能是后端暂时忙碌，请稍后重试。',
  onRetry,
}) => (
  <div className="ui-card flex flex-col items-center justify-center px-6 py-14 text-center" role="alert">
    <div className="mb-4 flex h-12 w-12 items-center justify-center rounded-2xl bg-red-50 text-red-500 dark:bg-red-500/10">
      <AlertCircle size={22} aria-hidden="true" />
    </div>
    <h3 className="text-base font-bold dk-text">{title}</h3>
    <p className="mt-1 max-w-md text-sm leading-relaxed text-gray-500 dark:text-gray-400">{description}</p>
    {onRetry && (
      <button
        type="button"
        onClick={onRetry}
        className="ui-control mt-5 inline-flex items-center gap-2 bg-[#07c160] px-4 text-sm font-bold text-white transition-colors hover:bg-[#06ad56]"
      >
        <RefreshCw size={15} aria-hidden="true" />
        重新加载
      </button>
    )}
  </div>
);
