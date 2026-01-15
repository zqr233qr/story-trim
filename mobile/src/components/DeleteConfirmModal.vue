<template>
  <view v-if="visible" class="fixed inset-0 z-[300] flex items-center justify-center bg-black/50" @click.self="handleCancel">
    <view class="bg-white rounded-2xl w-72 overflow-hidden">
      <view class="p-4 border-b border-stone-100">
        <text class="text-lg font-bold text-stone-800">{{ title }}</text>
      </view>
      <view class="p-4">
        <text class="text-sm text-stone-500 leading-relaxed">{{ content }}</text>
      </view>
      <view class="flex border-t border-stone-100">
        <view 
          class="flex-1 py-3 text-center active:bg-stone-50" 
          @click="handleCancel"
        >
          <text class="text-stone-500 font-medium">取消</text>
        </view>
        <view class="w-px bg-stone-100"></view>
        <view 
          class="flex-1 py-3 text-center active:bg-red-50" 
          @click="handleConfirm"
        >
          <text class="text-red-500 font-medium">{{ confirmText }}</text>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue';

const props = withDefaults(defineProps<{
  visible: boolean;
  title?: string;
  content?: string;
  confirmText?: string;
}>(), {
  title: '确认操作',
  content: '确定要执行此操作吗？',
  confirmText: '确定'
});

const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void;
  (e: 'confirm'): void;
  (e: 'cancel'): void;
}>();

const handleCancel = () => {
  emit('cancel');
  emit('update:visible', false);
};

const handleConfirm = () => {
  emit('confirm');
  emit('update:visible', false);
};
</script>
