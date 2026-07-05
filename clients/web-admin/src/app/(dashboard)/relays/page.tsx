"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { ArrowRightLeft, RefreshCw } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  listRelays,
  getRelayStats,
  forceUnregisterRelay,
  bulkMigrateSessions,
  RelayInfo,
  RelayStats,
  RelayListResponse,
} from "@/lib/api/admin";
import { RelayStatsCards } from "./relay-stats-cards";
import { RelayListCard } from "./relay-list-card";

export default function RelaysPage() {
  const router = useRouter();
  const [selectedSource, setSelectedSource] = useState<string>("");
  const [selectedTarget, setSelectedTarget] = useState<string>("");
  const [isMigrating, setIsMigrating] = useState(false);

  const [relaysData, setRelaysData] = useState<RelayListResponse | null>(null);
  const [stats, setStats] = useState<RelayStats | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [refetchKey, setRefetchKey] = useState(0);

  useEffect(() => {
    let cancelled = false;
    listRelays()
      .then((result) => {
        if (cancelled) return;
        setRelaysData(result);
        setIsLoading(false);
      })
      .catch(() => {
        if (cancelled) return;
        setIsLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [refetchKey]);

  useEffect(() => {
    let cancelled = false;
    getRelayStats()
      .then((result) => {
        if (!cancelled) setStats(result);
      })
      .catch(() => {
        // Non-critical
      });
    return () => {
      cancelled = true;
    };
  }, [refetchKey]);

  useEffect(() => {
    const interval = setInterval(() => setRefetchKey((k) => k + 1), 10000);
    return () => clearInterval(interval);
  }, []);

  const handleUnregister = async (relay: RelayInfo, migrate: boolean) => {
    const msg = migrate
      ? `注销中继 "${relay.id}" 并将所有会话迁移到其他中继？`
      : `注销中继 "${relay.id}"？${relay.connections} 个活跃连接将受影响。`;
    if (!confirm(msg)) return;
    try {
      const data = await forceUnregisterRelay(relay.id, migrate);
      toast.success(`中继已注销，影响 ${data.affected_sessions} 个会话。`);
      setRefetchKey((k) => k + 1);
    } catch (err: unknown) {
      toast.error((err as { error?: string })?.error || "注销中继失败");
    }
  };

  const handleBulkMigrate = async () => {
    if (!selectedSource || !selectedTarget) {
      toast.error("请选择源中继和目标中继");
      return;
    }
    if (selectedSource === selectedTarget) {
      toast.error("源中继和目标中继不能相同");
      return;
    }
    if (!confirm(`将所有会话从 "${selectedSource}" 迁移到 "${selectedTarget}"？`)) return;
    setIsMigrating(true);
    try {
      const data = await bulkMigrateSessions(selectedSource, selectedTarget);
      toast.success(`迁移完成：${data.migrated}/${data.total} 个会话已迁移`);
      setSelectedSource("");
      setSelectedTarget("");
      setRefetchKey((k) => k + 1);
    } catch (err: unknown) {
      toast.error((err as { error?: string })?.error || "迁移会话失败");
    } finally {
      setIsMigrating(false);
    }
  };

  const healthyRelays = relaysData?.data.filter((r) => r.healthy) || [];

  return (
    <div className="space-y-4">
      <RelayStatsCards stats={stats} />

      {healthyRelays.length >= 2 && (
        <Card>
          <CardHeader>
            <CardTitle className="text-base">批量会话迁移</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex flex-col gap-4 sm:flex-row sm:items-center">
              <div className="flex-1">
                <Select value={selectedSource} onValueChange={setSelectedSource}>
                  <SelectTrigger>
                    <SelectValue placeholder="源中继" displayValue={selectedSource || undefined} />
                  </SelectTrigger>
                  <SelectContent>
                    {relaysData?.data.map((relay) => (
                      <SelectItem key={relay.id} value={relay.id}>
                        {relay.id} ({relay.connections} 个连接)
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <ArrowRightLeft className="h-4 w-4 text-muted-foreground" />
              <div className="flex-1">
                <Select value={selectedTarget} onValueChange={setSelectedTarget}>
                  <SelectTrigger>
                    <SelectValue placeholder="目标中继" displayValue={selectedTarget || undefined} />
                  </SelectTrigger>
                  <SelectContent>
                    {healthyRelays
                      .filter((r) => r.id !== selectedSource)
                      .map((relay) => (
                        <SelectItem key={relay.id} value={relay.id}>
                          {relay.id} ({relay.region})
                        </SelectItem>
                      ))}
                  </SelectContent>
                </Select>
              </div>
              <Button
                onClick={handleBulkMigrate}
                disabled={!selectedSource || !selectedTarget || isMigrating}
              >
                {isMigrating ? (
                  <RefreshCw className="mr-2 h-4 w-4 animate-spin" />
                ) : (
                  <ArrowRightLeft className="mr-2 h-4 w-4" />
                )}
                迁移
              </Button>
            </div>
          </CardContent>
        </Card>
      )}

      <RelayListCard
        relaysData={relaysData}
        isLoading={isLoading}
        onRelayClick={(id) => router.push(`/relays/${encodeURIComponent(id)}`)}
        onUnregister={handleUnregister}
      />
    </div>
  );
}
