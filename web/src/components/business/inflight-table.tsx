"use client";

import { Fragment, useState } from "react";
import { useTranslations } from "next-intl";

import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { CopyableText } from "@/components/business/copyable-text";
import { EntityLabel } from "@/components/business/entity-label";
import { AttemptDots } from "@/components/business/attempt-dots";
import { FallbackChainInline } from "@/components/business/attempt-state";
import type { InflightSnapshot } from "@/lib/types";

export interface InflightRow extends InflightSnapshot {
  agent_id?: number;
  agent_name?: string;
}

interface InflightTableProps {
  rows: InflightRow[];
  showAgent?: boolean;
  onInterrupt?: (row: InflightRow) => void;
  onSelectRow?: (row: InflightRow) => void;
  emptyText: string;
}

export function InflightTable({ rows, showAgent, onInterrupt, onSelectRow, emptyText }: InflightTableProps) {
  const t = useTranslations("agents");
  const [target, setTarget] = useState<InflightRow | null>(null);
  const [expanded, setExpanded] = useState<Set<string>>(new Set());
  const toggleExpand = (k: string) =>
    setExpanded((prev) => {
      const next = new Set(prev);
      if (next.has(k)) next.delete(k);
      else next.add(k);
      return next;
    });
  const colCount = 8 + (showAgent ? 1 : 0) + (onInterrupt ? 1 : 0);

  if (rows.length === 0) {
    return <p className="text-xs text-muted-foreground px-4 py-3">{emptyText}</p>;
  }

  return (
    <>
      <Table>
        <TableHeader>
          <TableRow>
            {showAgent && <TableHead className="h-8">{t("inflightColNode")}</TableHead>}
            <TableHead className="h-8">{t("inflightColModel")}</TableHead>
            <TableHead className="h-8">{t("inflightColChannel")}</TableHead>
            <TableHead className="h-8">{t("inflightColStage")}</TableHead>
            <TableHead className="h-8">{t("inflightColRetry")}</TableHead>
            <TableHead className="h-8">{t("inflightColQueue")}</TableHead>
            <TableHead className="h-8">{t("inflightColElapsed")}</TableHead>
            <TableHead className="h-8">{t("inflightColStream")}</TableHead>
            <TableHead className="h-8">{t("inflightColReqId")}</TableHead>
            {onInterrupt && <TableHead className="h-8">{t("inflightColActions")}</TableHead>}
          </TableRow>
        </TableHeader>
        <TableBody>
          {rows.map((row) => {
            const elapsedSec = (row.elapsed_ms / 1000).toFixed(1);
            const isSlow = row.elapsed_ms > 60000;
            const rowKey = `${row.agent_id ?? 0}-${row.id}`;
            const isExpanded = expanded.has(rowKey);
            const hasChain =
              (row.view.fallback_chain?.length ?? 0) > 0 || !!row.current_attempt;
            return (
              <Fragment key={rowKey}>
              <TableRow
                className={onSelectRow ? "cursor-pointer hover:bg-muted/30" : undefined}
                onClick={onSelectRow ? () => onSelectRow(row) : undefined}
              >
                {showAgent && <TableCell className="py-1.5">{row.agent_name}</TableCell>}
                <TableCell className="py-1.5">{row.view.model_name}</TableCell>
                <TableCell className="py-1.5">
                  <EntityLabel entity="channel" id={row.view.channel_id} />
                </TableCell>
                <TableCell className="py-1.5">{row.stage}</TableCell>
                <TableCell className="py-1.5" onClick={(e) => e.stopPropagation()}>
                  <button
                    type="button"
                    className="flex items-center gap-1 cursor-pointer disabled:cursor-default"
                    disabled={!hasChain}
                    aria-expanded={isExpanded}
                    onClick={() => toggleExpand(rowKey)}
                  >
                    <AttemptDots chain={row.view.fallback_chain} pending={row.current_attempt} />
                  </button>
                </TableCell>
                <TableCell className="py-1.5">
                  {row.queued_ms > 0 ? (
                    <Badge variant="outline" className="text-xs px-1.5 py-0 border-amber-500/60 text-amber-600">
                      {t("inflightQueuedFor", { s: (row.queued_ms / 1000).toFixed(1) })}
                      {row.queued_reason ? ` · ${row.queued_reason}` : ""}
                    </Badge>
                  ) : (
                    <span className="text-muted-foreground">—</span>
                  )}
                </TableCell>
                <TableCell className="py-1.5">
                  {isSlow ? (
                    <span className="text-destructive font-medium">
                      {elapsedSec}s ({t("inflightSlowHint")})
                    </span>
                  ) : (
                    <span>{elapsedSec}s</span>
                  )}
                </TableCell>
                <TableCell className="py-1.5">
                  {row.view.is_stream ? (
                    <Badge variant="secondary" className="text-xs px-1.5 py-0">
                      {t("inflightStreamYes")}
                    </Badge>
                  ) : (
                    <span className="text-muted-foreground">{t("inflightStreamNo")}</span>
                  )}
                </TableCell>
                <TableCell className="py-1.5 font-mono" onClick={(e) => e.stopPropagation()}>
                  <CopyableText text={row.req_id} />
                </TableCell>
                {onInterrupt && (
                  <TableCell className="py-1.5" onClick={(e) => e.stopPropagation()}>
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-7 text-xs text-destructive"
                      onClick={() => setTarget(row)}
                    >
                      {t("inflightInterrupt")}
                    </Button>
                  </TableCell>
                )}
              </TableRow>
              {isExpanded && hasChain && (
                <TableRow className="bg-muted/20 hover:bg-muted/20">
                  <TableCell colSpan={colCount} className="py-2 pl-10">
                    <FallbackChainInline
                      chain={row.view.fallback_chain}
                      pending={row.current_attempt}
                    />
                  </TableCell>
                </TableRow>
              )}
              </Fragment>
            );
          })}
        </TableBody>
      </Table>

      <AlertDialog open={target !== null} onOpenChange={(o) => { if (!o) setTarget(null); }}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{t("inflightInterruptConfirmTitle")}</AlertDialogTitle>
            <AlertDialogDescription>{t("inflightInterruptConfirmDesc")}</AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>{t("inflightInterruptCancel")}</AlertDialogCancel>
            <AlertDialogAction
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
              onClick={() => { if (target) onInterrupt?.(target); setTarget(null); }}
            >
              {t("inflightInterrupt")}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}
