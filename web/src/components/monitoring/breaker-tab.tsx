"use client";

import { useTranslations } from "next-intl";

import { LiveTabHeader } from "@/components/monitoring/live-tab-header";
import { BreakerBoard } from "@/components/observability/breaker-board";

export function BreakerTab() {
  const t = useTranslations("monitoring");
  return (
    <div className="space-y-4">
      <LiveTabHeader title={t("tab.breaker")} />
      <BreakerBoard />
    </div>
  );
}
