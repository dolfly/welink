import React from 'react';
import {
  Area,
  AreaChart,
  Bar,
  BarChart,
  Cell,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts';

interface MonthlyPoint {
  month: string;
  label: string;
  count: number;
}

interface TimePoint {
  label: string;
  value: number;
  isLateNight?: boolean;
}

export const GroupMonthlyTrendChart: React.FC<{ data: MonthlyPoint[] }> = ({ data }) => (
  <ResponsiveContainer width="100%" height={120}>
    <AreaChart data={data} margin={{ top: 4, right: 4, bottom: 0, left: -30 }}>
      <defs>
        <linearGradient id="groupTrendGrad" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stopColor="#07c160" stopOpacity={0.3} />
          <stop offset="100%" stopColor="#07c160" stopOpacity={0.02} />
        </linearGradient>
      </defs>
      <XAxis
        dataKey="label"
        tick={{ fontSize: 9, fill: '#bbb' }}
        tickLine={false}
        interval={Math.max(0, Math.floor(data.length / 8) - 1)}
      />
      <YAxis tick={false} axisLine={false} tickLine={false} />
      <Tooltip
        contentStyle={{ borderRadius: 8, fontSize: 12 }}
        formatter={(value: number) => [`${value.toLocaleString()} 条`, '消息数']}
        labelFormatter={(label: string) => `20${label.replace('/', ' 年 ')} 月`}
      />
      <Area
        type="monotone"
        dataKey="count"
        stroke="#07c160"
        strokeWidth={2}
        fill="url(#groupTrendGrad)"
        dot={false}
        activeDot={{ r: 4, fill: '#07c160' }}
      />
    </AreaChart>
  </ResponsiveContainer>
);

export const GroupTimeDistributionCharts: React.FC<{
  hourlyData: TimePoint[];
  weeklyData: TimePoint[];
}> = ({ hourlyData, weeklyData }) => (
  <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
    <div>
      <div className="text-xs text-gray-500 font-semibold mb-1.5">24 小时</div>
      <ResponsiveContainer width="100%" height={90}>
        <BarChart data={hourlyData} margin={{ top: 0, right: 0, bottom: 0, left: -30 }}>
          <XAxis dataKey="label" tick={{ fontSize: 9, fill: '#bbb' }} tickLine={false} interval={3} />
          <YAxis tick={false} axisLine={false} tickLine={false} />
          <Tooltip
            contentStyle={{ borderRadius: 8, fontSize: 12 }}
            formatter={(value: number) => [`${value} 条`, '']}
            labelFormatter={(label: string) => `${label}:00`}
          />
          <Bar dataKey="value" radius={[3, 3, 0, 0]} maxBarSize={14}>
            {hourlyData.map((entry, index) => (
              <Cell key={index} fill={entry.isLateNight ? '#576b95' : '#10aeff'} opacity={0.8} />
            ))}
          </Bar>
        </BarChart>
      </ResponsiveContainer>
    </div>
    <div>
      <div className="text-xs text-gray-500 font-semibold mb-1.5">每周</div>
      <ResponsiveContainer width="100%" height={90}>
        <BarChart data={weeklyData} margin={{ top: 0, right: 0, bottom: 0, left: -30 }}>
          <XAxis dataKey="label" tick={{ fontSize: 10, fill: '#999' }} tickLine={false} />
          <YAxis tick={false} axisLine={false} tickLine={false} />
          <Tooltip contentStyle={{ borderRadius: 8, fontSize: 12 }} formatter={(value: number) => [`${value} 条`, '']} />
          <Bar dataKey="value" fill="#07c160" radius={[4, 4, 0, 0]} maxBarSize={28} opacity={0.8} />
        </BarChart>
      </ResponsiveContainer>
    </div>
  </div>
);
