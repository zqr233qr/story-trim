<script setup lang="ts">
import { computed, ref } from "vue";
import { onShow } from "@dcloudio/uni-app";
import { pointsApi } from "@/api/points";
import { useUserStore } from "@/stores/user";
import { useToastStore } from "@/stores/toast";

// 积分流水展示项。
interface LedgerItem {
  id: number;
  title: string;
  subtitle?: string;
  delta: number;
  time: string;
}

const userStore = useUserStore();
const toastStore = useToastStore();
const pointsBalance = ref(0);
const records = ref<LedgerItem[]>([]);
const isLoading = ref(false);

// 加载积分余额。
const loadPointsBalance = async () => {
  if (!userStore.isLoggedIn()) {
    pointsBalance.value = 0;
    return;
  }
  try {
    const res = await pointsApi.getBalance();
    if (res.code === 0) {
      pointsBalance.value = res.data.balance || 0;
      return;
    }
    toastStore.show({ message: res.msg || "获取积分失败", type: "error" });
  } catch (error) {
    console.warn("[Points] load balance failed", error);
  }
};

// 映射积分流水标题。
const mapLedgerTitle = (reason: string) => {
  if (reason === "register_bonus") return "注册赠送";
  if (reason === "trim_use") return "精简消耗";
  if (reason === "trim_refund") return "精简退款";
  if (reason === "recharge") return "积分充值";
  if (reason === "manual_adjust") return "积分调整";
  return "积分变动";
};

// 拼接流水附加信息。
const buildLedgerSubtitle = (extra?: Record<string, string>) => {
  if (!extra) return "";
  const parts: string[] = [];
  if (extra.book_title) {
    parts.push(`《${extra.book_title}》`);
  }
  if (extra.chapter_title) {
    parts.push(extra.chapter_title);
  }
  if (extra.prompt_name) {
    parts.push(extra.prompt_name);
  }
  return parts.join(" · ");
};

// 加载积分流水。
const loadPointsLedger = async () => {
  if (!userStore.isLoggedIn()) {
    records.value = [];
    return;
  }
  isLoading.value = true;
  try {
    const res = await pointsApi.getLedger(1, 30);
    if (res.code === 0) {
      records.value = (res.data.items || []).map((item) => ({
        id: item.id,
        title: mapLedgerTitle(item.reason),
        subtitle: buildLedgerSubtitle(item.extra),
        delta: item.change,
        time: item.created_at,
      }));
      return;
    }
    toastStore.show({ message: res.msg || "获取积分流水失败", type: "error" });
  } catch (error) {
    console.warn("[Points] load ledger failed", error);
  } finally {
    isLoading.value = false;
  }
};

onShow(() => {
  loadPointsBalance();
  loadPointsLedger();
});

// 格式化积分显示。
const formatDelta = (delta: number) => {
  return delta > 0 ? `+${delta}` : `${delta}`;
};

// 区分积分颜色。
const deltaClass = (delta: number) => {
  return delta > 0 ? "text-emerald-600" : "text-rose-500";
};

const recordList = computed(() => records.value);
</script>

<template>
  <view class="min-h-screen bg-stone-50 p-6">
    <view class="bg-white rounded-3xl p-6 shadow-[0_8px_30px_rgba(0,0,0,0.04)]">
      <text class="text-xs text-stone-400 font-semibold tracking-[0.2em]">积分余额</text>
      <text class="block text-3xl font-black text-stone-900 mt-2">{{ pointsBalance }}</text>
    </view>

    <view class="mt-6 bg-white rounded-3xl p-6 shadow-[0_8px_30px_rgba(0,0,0,0.04)]">
      <text class="text-xs text-stone-400 font-semibold tracking-[0.2em]">最近积分记录</text>
      <view v-if="isLoading" class="mt-6 text-center text-xs text-stone-400">正在加载积分记录...</view>
      <view v-else-if="recordList.length === 0" class="mt-6 text-center text-xs text-stone-400">暂无积分记录</view>
      <view v-else class="mt-4 flex flex-col gap-4">
        <view
          v-for="record in recordList"
          :key="record.id"
          class="flex items-center justify-between border-b border-stone-100 pb-3 last:border-0 last:pb-0"
        >
          <view class="flex flex-col">
            <text class="text-sm font-semibold text-stone-800">{{ record.title }}</text>
            <text v-if="record.subtitle" class="text-xs text-stone-400 mt-1">{{ record.subtitle }}</text>
            <text class="text-xs text-stone-400 mt-1">{{ record.time }}</text>
          </view>
          <text :class="`text-sm font-bold ${deltaClass(record.delta)}`">{{ formatDelta(record.delta) }}</text>
        </view>
      </view>
    </view>
  </view>
</template>
