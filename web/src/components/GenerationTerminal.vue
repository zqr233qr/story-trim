<script setup lang="ts">
import { ref, watch, nextTick } from 'vue'

const props = defineProps<{
  show: boolean,
  content: string,
  title: string,
  isDarkMode?: boolean
}>()

const emit = defineEmits(['close'])
const viewport = ref<HTMLElement | null>(null)

// Auto-scroll
watch(() => props.content, async () => {
  await nextTick()
  if (viewport.value) {
    viewport.value.scrollTop = viewport.value.scrollHeight
  }
})
</script>

<template>
  <transition name="slide-up">
    <div v-if="show" class="fixed inset-0 z-[60] flex items-end justify-center pointer-events-none">
      <!-- Backdrop -->
      <div @click.stop="emit('close')" class="absolute inset-0 bg-black/10 backdrop-blur-[1px] pointer-events-auto transition-opacity"></div>

      <!-- Terminal Panel -->
      <div :class="isDarkMode ? 'bg-stone-950 border-t border-stone-800 text-stone-300' : 'bg-white border-t border-stone-200 text-stone-800 shadow-2xl'" 
           class="w-full max-w-3xl h-[70vh] rounded-t-2xl pointer-events-auto flex flex-col overflow-hidden relative shadow-2xl">
        
        <!-- Header -->
        <div class="h-12 flex items-center px-4 border-b shrink-0 relative" 
             :class="isDarkMode ? 'border-stone-800 bg-stone-900/50' : 'border-stone-100 bg-stone-50/80'">
          
          <!-- Traffic Lights -->
          <div class="flex items-center gap-2 mr-4 group select-none">
            <button @click.stop="emit('close')" class="w-3 h-3 rounded-full bg-[#ff5f56] border border-[#e0443e] flex items-center justify-center transition-opacity hover:opacity-80 active:opacity-60">
              <svg class="w-2 h-2 text-black/50 opacity-0 group-hover:opacity-100 transition-opacity" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="4"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12"></path></svg>
            </button>
            <div class="w-3 h-3 rounded-full bg-[#ffbd2e] border border-[#dea123]"></div>
            <div class="w-3 h-3 rounded-full bg-[#27c93f] border border-[#1aab29]"></div>
          </div>

          <div class="flex items-center gap-2 opacity-70">
            <span class="text-xs font-mono font-bold tracking-widest uppercase">AI: {{ title }}</span>
          </div>
        </div>

        <!-- Content Area -->
        <div ref="viewport" class="flex-1 overflow-y-auto p-6 font-mono text-sm leading-relaxed whitespace-pre-wrap no-scrollbar">
          {{ content }}<span class="animate-pulse inline-block w-2 h-4 bg-teal-500 align-middle ml-1"></span>
        </div>

        <!-- Gradient Fade Overlay at bottom -->
        <div class="absolute bottom-0 inset-x-0 h-12 bg-gradient-to-t pointer-events-none"
             :class="isDarkMode ? 'from-stone-950 to-transparent' : 'from-white to-transparent'"></div>
      </div>
    </div>
  </transition>
</template>

<style scoped>
.slide-up-enter-active, .slide-up-leave-active { transition: transform 0.4s cubic-bezier(0.16, 1, 0.3, 1); }
.slide-up-enter-from, .slide-up-leave-to { transform: translateY(100%); }
</style>
