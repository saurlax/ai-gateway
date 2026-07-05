"use client";

import type { UsageLog, AttemptInProgress } from "@/lib/types";
import { attemptDotClass } from "@/components/business/attempt-state";

interface AttemptDotsProps {
  chain?: UsageLog["fallback_chain"];
  pending?: AttemptInProgress | null;
}

// AttemptDots 压缩呈现某在途请求的 fallback 进展:已结束尝试各一个按结果着色的点,
// 进行中候选一个脉冲蓝点,>1 次时标 ×N。详情由 inflight-table 点击就地展开(不再用原生 title)。
export function AttemptDots({ chain, pending }: AttemptDotsProps) {
  const settled = chain ?? [];
  const total = settled.length + (pending ? 1 : 0);
  if (total === 0) return <span className="text-muted-foreground">—</span>;
  return (
    <div className="flex items-center gap-1">
      {settled.map((e) => (
        <span key={e.seq} className={`inline-block h-2 w-2 rounded-full ${attemptDotClass(e)}`} />
      ))}
      {pending && <span className="inline-block h-2 w-2 rounded-full bg-blue-500 animate-pulse" />}
      {total > 1 && <span className="ml-0.5 text-xs text-muted-foreground">×{total}</span>}
    </div>
  );
}
