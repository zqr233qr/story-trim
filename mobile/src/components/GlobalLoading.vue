<script setup lang="ts">
import { useLoadingStore } from "@/stores/loading";

const loadingStore = useLoadingStore();
</script>

<template>
  <view
    v-if="loadingStore.visible"
    class="fixed inset-0 z-[250] flex items-center justify-center pointer-events-none"
  >
    <view class="px-6 py-4 rounded-2xl bg-white/90 shadow-[0_10px_30px_rgba(0,0,0,0.12)] backdrop-blur">
      <view class="loading-spinner"></view>
    </view>
  </view>
</template>

<style scoped>
.loading-spinner {
  width: 28px;
  height: 28px;
  border-radius: 9999px;
  border: 2px solid rgba(148, 163, 184, 0.2);
  border-top-color: rgba(51, 65, 85, 0.8);
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
