<script setup lang="ts">
import { computed, ref } from "vue";
import { onShow } from "@dcloudio/uni-app";
import { useUserStore } from "@/stores/user";
import { useBookStore } from "@/stores/book";
import { pointsApi } from "@/api/points";

const userStore = useUserStore();
const bookStore = useBookStore();
const isLoggedIn = computed(() => userStore.isLoggedIn());
const pointsBalance = ref(0);

// 加载积分余额。
const loadPointsBalance = async () => {
  if (!isLoggedIn.value) {
    pointsBalance.value = 0;
    return;
  }
  try {
    const res = await pointsApi.getBalance();
    if (res.code === 0) {
      pointsBalance.value = res.data.balance || 0;
      return;
    }
    uni.showToast({ title: res.msg || "获取积分失败", icon: "none" });
  } catch (error) {
    console.warn("[Profile] load points failed", error);
  }
};

onShow(() => {
  loadPointsBalance();
});

// 打开积分明细页面。
const openPoints = () => {
  uni.navigateTo({ url: "/pages/user/points" });
};

// 退出登录并返回书架。
const handleLogout = async () => {
  if (!isLoggedIn.value) return;
  const res = await new Promise<UniApp.ShowModalRes>((resolve) => {
    uni.showModal({
      title: "退出登录",
      content: "确定要退出当前账号吗？本地数据将保留，但无法同步云端进度。",
      success: resolve,
      fail: () => resolve({ confirm: false, cancel: true } as UniApp.ShowModalRes),
    });
  });
  if (!res.confirm) return;
  userStore.logout();
  await bookStore.fetchBooks();
  uni.navigateBack();
};

// 跳转登录页。
const handleLogin = () => {
  uni.navigateTo({ url: "/pages/login/login" });
};
</script>

<template>
  <view class="min-h-screen bg-stone-50 p-6">
    <view class="bg-white rounded-3xl p-6 shadow-[0_8px_30px_rgba(0,0,0,0.04)]">
      <text class="text-xs text-stone-400 font-semibold tracking-[0.2em]">账户信息</text>
      <view class="mt-4">
        <text class="text-xl font-bold text-stone-900">{{ userStore.username || "未登录" }}</text>
      </view>

      <view class="mt-6 bg-stone-50 rounded-2xl p-5">
        <text class="text-xs text-stone-400 font-semibold tracking-[0.2em]">当前积分</text>
        <text class="block text-3xl font-black text-stone-900 mt-2">{{ pointsBalance }}</text>
      </view>

      <view class="mt-6 flex gap-3">
        <button
          class="flex-1 bg-stone-900 text-white text-sm rounded-xl py-2"
          @click="openPoints"
        >
          查看积分明细
        </button>
        <button
          v-if="isLoggedIn"
          class="flex-1 bg-stone-100 text-stone-700 text-sm rounded-xl py-2"
          @click="handleLogout"
        >
          退出登录
        </button>
        <button
          v-else
          class="flex-1 bg-stone-100 text-stone-700 text-sm rounded-xl py-2"
          @click="handleLogin"
        >
          立即登录
        </button>
      </view>
    </view>
  </view>
</template>
