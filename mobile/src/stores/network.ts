import { ref } from "vue";
import { defineStore } from "pinia";
import { api } from "@/api";

// 全局网络状态管理。
export const useNetworkStore = defineStore("network", () => {
  const serverReachable = ref(true);
  const lastPingAt = ref(0);
  let pingTimer: ReturnType<typeof setInterval> | null = null;

  // 执行一次健康探测。
  const pingServer = async () => {
    try {
      const res = await api.ping();
      serverReachable.value = res.code === 0;
    } catch (e) {
      console.warn("[Network] ping failed", e);
      serverReachable.value = false;
    } finally {
      lastPingAt.value = Date.now();
    }
  };

  // 启动定时探测。
  const startPing = () => {
    if (pingTimer) return;
    pingServer();
    pingTimer = setInterval(pingServer, 5000);
  };

  // 停止定时探测。
  const stopPing = () => {
    if (!pingTimer) return;
    clearInterval(pingTimer);
    pingTimer = null;
  };

  return {
    serverReachable,
    lastPingAt,
    startPing,
    stopPing,
    pingServer,
  };
});
