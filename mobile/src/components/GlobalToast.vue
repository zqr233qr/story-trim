<script setup lang="ts">
import { computed } from "vue";
import { useToastStore } from "@/stores/toast";

const toastStore = useToastStore();

const typeClasses = computed(() => {
  switch (toastStore.type) {
    case "success":
      return "bg-emerald-50/80 text-emerald-700";
    case "error":
      return "bg-rose-50/80 text-rose-700";
    default:
      return "bg-stone-50/80 text-stone-700";
  }
});
</script>

<template>
  <view
    v-if="toastStore.visible"
    class="fixed bottom-6 inset-x-0 z-[300] flex items-center justify-center px-6"
  >
    <view
      :class="typeClasses"
      class="px-4 py-1.5 rounded-full shadow-[0_6px_16px_rgba(0,0,0,0.08)] text-xs font-medium transition-opacity duration-200 animate-fade backdrop-blur"
    >
      {{ toastStore.message }}
    </view>
  </view>
</template>

<style scoped>
@keyframes toast-fade {
  0% {
    opacity: 0;
    transform: translateY(6px);
  }
  100% {
    opacity: 1;
    transform: translateY(0);
  }
}
.animate-fade {
  animation: toast-fade 0.2s ease-out;
}
</style>
