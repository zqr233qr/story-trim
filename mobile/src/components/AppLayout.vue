<script setup lang="ts">
import GlobalToast from "@/components/GlobalToast.vue";
import GlobalLoading from "@/components/GlobalLoading.vue";
</script>

<template>
  <view class="relative min-h-screen">
    <slot />
    <GlobalToast />
    <GlobalLoading />
  </view>
</template>
