<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(defineProps<{
  visible: boolean
  content?: string
  readingMode?: 'light' | 'dark' | 'sepia'
}>(), {
  readingMode: 'light'
})

const emit = defineEmits(['update:visible', 'confirm'])

const theme = computed(() => {
  switch (props.readingMode) {
    case 'dark':
      return {
        bg: 'bg-stone-900',
        textMain: 'text-stone-100',
        textSub: 'text-stone-400',
        iconBg: 'bg-stone-800',
        iconColor: 'text-stone-400', // SVG 用 class 控制颜色稍麻烦，这里主要控制容器
        btnPrimary: 'bg-stone-100 text-stone-900',
        btnSecondary: 'bg-stone-900 border border-stone-700 text-stone-400'
      }
    case 'sepia':
      return {
        bg: 'bg-[#f7f1e3]', // 经典的羊皮纸色
        textMain: 'text-[#5b4636]',
        textSub: 'text-[#887363]',
        iconBg: 'bg-[#ebe3d1]',
        iconColor: 'text-[#887363]',
        btnPrimary: 'bg-[#5b4636] text-[#f7f1e3]',
        btnSecondary: 'bg-[#f7f1e3] border border-[#dcd3bf] text-[#887363]'
      }
    default: // light
      return {
        bg: 'bg-white',
        textMain: 'text-stone-900',
        textSub: 'text-stone-500',
        iconBg: 'bg-stone-50',
        iconColor: 'text-stone-900',
        btnPrimary: 'bg-stone-900 text-white',
        btnSecondary: 'bg-white border border-stone-200 text-stone-500'
      }
  }
})

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
    <!-- Backdrop with Blur -->
    <view 
      class="absolute inset-0 bg-black/40 backdrop-blur-sm transition-opacity" 
      @click="handleClose"
    ></view>

    <!-- Modal Card -->
    <view 
      class="relative w-full max-w-[320px] rounded-[2rem] p-8 shadow-2xl flex flex-col items-center text-center animate-in transition-colors duration-300"
      :class="theme.bg"
    >
      <!-- Icon -->
      <view class="w-14 h-14 rounded-2xl flex items-center justify-center mb-5 rotate-3 shadow-sm border border-black/5" :class="theme.iconBg">
        <image src="/static/icons/lock.svg" class="w-6 h-6 opacity-60" :class="theme.textMain" />
      </view>

      <!-- Content -->
      <text class="text-lg font-bold tracking-tight mb-2" :class="theme.textMain">需要登录</text>
      <text class="text-sm leading-relaxed mb-8" :class="theme.textSub">
        {{ content || '此功能需要登录账号后才能使用，是否现在去登录？' }}
      </text>

      <!-- Actions -->
      <view class="w-full flex flex-col gap-3">
        <view 
          @click="handleConfirm"
          class="w-full py-3.5 rounded-xl flex items-center justify-center active:scale-[0.98] transition-transform shadow-lg shadow-black/5"
          :class="theme.btnPrimary"
        >
          <text class="text-sm font-bold tracking-wide">立即登录</text>
        </view>
        
        <view 
          @click="handleClose"
          class="w-full py-3.5 rounded-xl flex items-center justify-center active:opacity-80 transition-colors"
          :class="theme.btnSecondary"
        >
          <text class="text-sm font-bold">暂不登录</text>
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