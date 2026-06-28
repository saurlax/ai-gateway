"use client";

import { useTranslations } from "next-intl";
import { Plus, Trash2 } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { DateTimePicker } from "@/components/business/date-picker/date-time-picker";
import { UNIT_QUOTA_SCALE } from "@/lib/utils/format";
import { ChannelLimit } from "./utils";

type Rule = NonNullable<ChannelLimit["rules"]>[number];

const METRICS: Array<Rule["metric"]> = ["calls", "cost"];
const WINDOWS: Array<Rule["window"]> = ["lifetime", "daily", "weekly", "monthly", "rolling_days"];

export interface LimitRulesEditorProps {
  limit: ChannelLimit;
  onChange: (next: ChannelLimit) => void;
}

export function LimitRulesEditor({ limit, onChange }: LimitRulesEditorProps) {
  const t = useTranslations("channels");
  const rules: Rule[] = limit.rules ?? [];

  const updateRule = (idx: number, patch: Partial<Rule>) => {
    onChange({ ...limit, rules: rules.map((r, i) => (i === idx ? { ...r, ...patch } : r)) });
  };
  const addRule = () => {
    const blank: Rule = { metric: "cost", window: "monthly", threshold: 0, cost_basis: "raw" };
    onChange({ ...limit, rules: [...rules, blank] });
  };
  const removeRule = (idx: number) => {
    onChange({ ...limit, rules: rules.filter((_, i) => i !== idx) });
  };
  const thresholdDisplay = (r: Rule): string => {
    if (r.metric === "cost") return r.threshold ? String(r.threshold / UNIT_QUOTA_SCALE) : "";
    return r.threshold ? String(r.threshold) : "";
  };
  const onThresholdChange = (idx: number, r: Rule, raw: string) => {
    const n = Number(raw);
    const v = Number.isNaN(n) ? 0 : n;
    const threshold = r.metric === "cost" ? Math.round(v * UNIT_QUOTA_SCALE) : Math.round(v);
    updateRule(idx, { threshold });
  };

  return (
    <div className="space-y-4">
      <p className="text-xs text-muted-foreground">{t("usageLimitHint")}</p>

      <div className="space-y-2">
        <Label>{t("limitCutoff")}</Label>
        <DateTimePicker
          value={limit.disable_at || null}
          onChange={(v) => onChange({ ...limit, disable_at: v ?? 0 })}
          placeholder={t("limitCutoff")}
        />
        <p className="text-xs text-muted-foreground">{t("limitCutoffHint")}</p>
      </div>

      <div className="space-y-2">
        <Label>{t("limitRules")}</Label>
        {rules.map((r, idx) => (
          <div key={idx} className="flex flex-wrap items-end gap-2 rounded-md border p-2">
            <div className="space-y-1">
              <Label className="text-xs">{t("limitMetric")}</Label>
              <Select value={r.metric} onValueChange={(v) => updateRule(idx, { metric: v as Rule["metric"] })}>
                <SelectTrigger className="w-28"><SelectValue /></SelectTrigger>
                <SelectContent>
                  {METRICS.map((m) => (
                    <SelectItem key={m} value={m}>{t(m === "calls" ? "metricCalls" : "metricCost")}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            {r.metric === "cost" && (
              <div className="space-y-1">
                <Label className="text-xs">{t("costBasis")}</Label>
                <Select
                  value={r.cost_basis ?? "raw"}
                  onValueChange={(v) => updateRule(idx, { cost_basis: v as Rule["cost_basis"] })}
                >
                  <SelectTrigger className="w-36"><SelectValue /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="raw">{t("costBasisRaw")}</SelectItem>
                    <SelectItem value="billed">{t("costBasisBilled")}</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            )}
            <div className="space-y-1">
              <Label className="text-xs">{t("limitWindow")}</Label>
              <Select value={r.window} onValueChange={(v) => updateRule(idx, { window: v as Rule["window"] })}>
                <SelectTrigger className="w-28"><SelectValue /></SelectTrigger>
                <SelectContent>
                  {WINDOWS.map((w) => (
                    <SelectItem key={w} value={w}>
                      {t(
                        w === "lifetime" ? "windowLifetime"
                        : w === "daily" ? "windowDaily"
                        : w === "weekly" ? "windowWeekly"
                        : w === "monthly" ? "windowMonthly"
                        : "windowRollingDays"
                      )}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-1">
              <Label className="text-xs">{t("limitThreshold")}</Label>
              <Input
                type="number"
                min={0}
                className="w-32"
                value={thresholdDisplay(r)}
                onChange={(e) => onThresholdChange(idx, r, e.target.value)}
              />
            </div>
            {r.window === "rolling_days" && (
              <div className="space-y-1">
                <Label className="text-xs">{t("limitDays")}</Label>
                <Input
                  type="number"
                  min={1}
                  className="w-20"
                  value={r.days ?? ""}
                  onChange={(e) => updateRule(idx, { days: Math.max(1, Number(e.target.value) || 1) })}
                />
              </div>
            )}
            <Button type="button" variant="ghost" size="icon" onClick={() => removeRule(idx)} aria-label={t("limitDelete")}>
              <Trash2 className="size-4" />
            </Button>
          </div>
        ))}
        <Button type="button" variant="outline" size="sm" onClick={addRule}>
          <Plus className="size-4" /> {t("limitAddRule")}
        </Button>
      </div>
    </div>
  );
}
