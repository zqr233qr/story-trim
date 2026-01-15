<template>
  <view v-if="modelValue" class="fixed inset-0 z-[200] flex flex-col justify-end" @touchmove.stop.prevent>
    <!-- Backdrop -->
    <view 
      class="absolute inset-0 bg-black/40 backdrop-blur-sm transition-opacity duration-300"
      @click="close"
    ></view>

    <!-- Sheet Content -->
    <view 
      class="bg-white rounded-t-3xl relative z-10 w-full overflow-hidden pb-safe transition-transform duration-300 transform translate-y-0"
      :class="{ 'translate-y-full': !animateShow }"
    >
      <!-- Drag Handle -->
      <view class="flex justify-center pt-3 pb-2" @click="close">
        <view class="w-10 h-1 rounded-full bg-stone-200"></view>
      </view>

      <!-- Book Info -->
      <view class="px-6 py-4 border-b border-stone-100">
        <text class="text-lg font-bold text-stone-800 line-clamp-1">{{ title }}</text>
        <text class="text-xs text-stone-400 mt-1">请选择操作</text>
      </view>

      <!-- Actions -->
      <view class="p-4 space-y-2">
        <!-- Sync Action -->
        <view 
          v-if="showSync"
          @click="handleAction('sync')"
          class="flex items-center gap-4 p-4 rounded-2xl active:bg-stone-50 transition-colors"
        >
          <view class="w-10 h-10 rounded-full bg-blue-50 flex items-center justify-center">
            <image src="/static/icons/cloud.svg" class="w-5 h-5 opacity-80" />
          </view>
          <view class="flex-1">
            <text class="text-base font-bold text-stone-700">同步至云端</text>
            <text class="text-xs text-stone-400 block mt-0.5">备份书籍进度与精简记录</text>
          </view>
        </view>

        <!-- Delete Action -->
        <view 
          @click="handleAction('delete')"
          class="flex items-center gap-4 p-4 rounded-2xl active:bg-red-50 transition-colors"
        >
          <view class="w-10 h-10 rounded-full bg-red-50 flex items-center justify-center">
            <image src="/static/icons/trash.svg" class="w-5 h-5 opacity-80" />
          </view>
          <view class="flex-1">
            <text class="text-base font-bold text-red-500">删除书籍</text>
            <text class="text-xs text-red-300 block mt-0.5">删除后不可恢复</text>
          </view>
        </view>
      </view>
      
      <!-- Cancel Button -->
      <view class="px-4 pb-6 pt-2">
        <view 
          @click="close"
          class="w-full py-3.5 bg-stone-100 rounded-2xl flex items-center justify-center active:bg-stone-200 transition-colors"
        >
          <text class="text-sm font-bold text-stone-500">取消</text>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue';

const props = withDefaults(defineProps<{
  modelValue: boolean;
  title: string;
  showSync?: boolean;
}>(), {
  showSync: true
});

const emit = defineEmits(['update:modelValue', 'action']);

const animateShow = ref(false);

watch(() => props.modelValue, (val) => {
  if (val) {
    // Small delay to allow v-if to render before animating transform
    setTimeout(() => {
      animateShow.value = true;
    }, 10);
  } else {
    animateShow.value = false;
  }
});

const close = () => {
  animateShow.value = false;
  setTimeout(() => {
    emit('update:modelValue', false);
  }, 300); // Wait for animation
};

const handleAction = (action: 'sync' | 'delete') => {
  emit('action', action);
  close();
};
</script>

<style scoped>
.pb-safe {
  padding-bottom: env(safe-area-inset-bottom);
}
</style>