<template>
  <view 
    v-if="hasActiveTasks"
    @click="emit('click')"
    class="task-indicator-fixed"
  >
    <view class="task-indicator-content active:scale-95 transition-transform">
      <text class="text-sm font-bold leading-none">任务处理中</text>
    </view>
  </view>
</template>

<script setup lang="ts">
defineProps<{
  'has-active-tasks': boolean
}>()

const emit = defineEmits(['click'])
</script>

<style scoped>
.task-indicator-fixed {
  position: fixed;
  bottom: 100rpx;
  left: 0;
  right: 0;
  display: flex;
  justify-content: center;
  z-index: 150;
  pointer-events: none;
}

.task-indicator-content {
  pointer-events: auto;
  background-color: #1c1917;
  color: white;
  padding: 24rpx 40rpx;
  border-radius: 100rpx;
  display: flex;
  align-items: center;
  box-shadow: 0 10rpx 30rpx rgba(0, 0, 0, 0.2);
}
</style>
