<script setup lang="ts">
const props = defineProps<{
  visible: boolean
  title?: string
  content?: string
  confirmText?: string
}>()

const emit = defineEmits(['update:visible', 'confirm'])

const handleClose = () => {
  emit('update:visible', false)
}

const handleConfirm = () => {
  emit('update:visible', false)
  emit('confirm')
}
</script>

<template>
  <view 
    v-if="visible" 
    class="fixed inset-0 z-[200] flex items-center justify-center p-8"
  >
    <!-- Backdrop -->
    <view 
      class="absolute inset-0 bg-black/40 backdrop-blur-sm transition-opacity" 
      @click="handleClose"
    ></view>

    <!-- Modal Card -->
    <view 
      class="relative w-full max-w-[320px] bg-white rounded-[2rem] p-8 shadow-2xl flex flex-col items-center text-center animate-in"
    >
      <!-- Icon -->
      <view class="w-14 h-14 bg-stone-50 rounded-2xl flex items-center justify-center mb-5 rotate-3 shadow-sm border border-stone-100">
        <image src="/static/icons/info.svg" class="w-7 h-7 text-stone-900 opacity-80" />
      </view>

      <!-- Content -->
      <text class="text-lg font-bold text-stone-900 tracking-tight mb-2">{{ title || '提示' }}</text>
      <text class="text-sm text-stone-500 leading-relaxed mb-8">
        {{ content }}
      </text>

      <!-- Actions -->
      <view class="w-full">
        <view 
          @click="handleConfirm"
          class="w-full bg-stone-900 text-white py-3.5 rounded-xl flex items-center justify-center active:scale-[0.98] transition-transform shadow-lg shadow-stone-200"
        >
          <text class="text-sm font-bold tracking-wide">{{ confirmText || '知道了' }}</text>
        </view>
      </view>
    </view>
  </view>
</template>

<style scoped>
@keyframes modal-in {
  0% { opacity: 0; transform: scale(0.95) translateY(10px); }
  100% { opacity: 1; transform: scale(1) translateY(0); }
}
.animate-in {
  animation: modal-in 0.2s cubic-bezier(0.16, 1, 0.3, 1) forwards;
}
</style>