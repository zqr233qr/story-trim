<script setup lang="ts">
import { ref, watch, nextTick } from 'vue'

const props = defineProps<{
  show: boolean,
  content: string,
  title: string,
  isDarkMode?: boolean
}>()

const emit = defineEmits(['close'])
const scrollIntoView = ref('')

// Auto-scroll logic for Uni-app
watch(() => props.content, async () => {
  await nextTick()
  scrollIntoView.value = 'bottom-marker'
})
</script>

<template>
  <view class="fixed inset-0 z-[100] pointer-events-none transition-all duration-300"
        :class="show ? 'opacity-100' : 'opacity-0 invisible'">
    
    <!-- Backdrop -->
    <view @click.stop="emit('close')" class="absolute inset-0 bg-black/40 backdrop-blur-[1px] pointer-events-auto"></view>

    <!-- Terminal Panel -->
    <view :class="[
            isDarkMode ? 'bg-stone-950 border-t border-stone-800 text-stone-300' : 'bg-white border-t border-stone-200 text-stone-800',
            show ? 'translate-y-0' : 'translate-y-full'
          ]" 
         class="absolute bottom-0 inset-x-0 w-full max-w-3xl mx-auto h-[70vh] rounded-t-2xl pointer-events-auto flex flex-col overflow-hidden shadow-2xl transition-transform duration-500 cubic-bezier(0.16, 1, 0.3, 1)">
      
      <!-- Header -->
      <view class="h-12 flex items-center px-4 border-b shrink-0 relative" 
           :class="isDarkMode ? 'border-stone-800 bg-stone-900/50' : 'border-stone-100 bg-stone-50/80'">
        
        <!-- Traffic Lights -->
        <view class="flex items-center gap-2 mr-4 group select-none">
          <view @click.stop="emit('close')" class="w-3 h-3 rounded-full bg-[#ff5f56] border border-[#e0443e] flex items-center justify-center transition-opacity hover:opacity-80 active:opacity-60">
            <text class="text-[8px] text-black/50 opacity-0 group-hover:opacity-100 transition-opacity">×</text>
          </view>
          <view class="w-3 h-3 rounded-full bg-[#ffbd2e] border border-[#dea123]"></view>
          <view class="w-3 h-3 rounded-full bg-[#27c93f] border border-[#1aab29]"></view>
        </view>

        <view class="flex items-center gap-2 opacity-70">
          <text class="text-xs font-mono font-bold tracking-widest uppercase">AI: {{ title }}</text>
        </view>
      </view>

      <!-- Content Area -->
      <scroll-view 
        scroll-y 
        :scroll-into-view="scrollIntoView" 
        scroll-with-animation
        class="flex-1 p-6 font-mono text-sm leading-relaxed whitespace-pre-wrap">
        <text>{{ content }}</text>
        <view class="inline-block w-2 h-4 bg-teal-500 align-middle ml-1"></view>
        <view id="bottom-marker" class="h-10"></view>
      </scroll-view>

      <!-- Gradient Fade Overlay -->
      <view class="absolute bottom-0 inset-x-0 h-12 bg-gradient-to-t pointer-events-none"
           :class="isDarkMode ? 'from-stone-950 to-transparent' : 'from-white to-transparent'"></view>
    </view>
  </view>
</template>

<style scoped>
.cubic-bezier { transition-timing-function: cubic-bezier(0.16, 1, 0.3, 1); }
</style>