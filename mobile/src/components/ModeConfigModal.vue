<script setup lang="ts">
import { ref, watch } from 'vue'

interface Prompt {
  id: number;
  name: string;
  description?: string;
}

const props = defineProps<{ 
  show: boolean, 
  bookTitle: string,
  chapterTitle: string,
  prompts: Prompt[],
  trimmedIds?: number[],
  isDarkMode?: boolean
}>()
const emit = defineEmits(['close', 'start'])

const selectedId = ref<number | string>('')

watch(() => props.prompts, (newPs) => {
  if (newPs.length > 0 && !selectedId.value) {
    selectedId.value = newPs[0].id
  }
}, { immediate: true })
</script>

<template>
  <view class="fixed inset-0 z-[100] flex items-end sm:items-center justify-center pointer-events-none transition-all duration-300"
        :class="show ? 'opacity-100' : 'opacity-0 invisible'">
    
    <!-- Backdrop -->
    <view @click.stop="emit('close')" class="absolute inset-0 bg-black/40 backdrop-blur-[1px] pointer-events-auto"></view>

    <!-- Content Card -->
    <view :class="[
            isDarkMode ? 'bg-stone-900 border border-stone-800' : 'bg-white shadow-2xl',
            show ? 'translate-y-0 scale-100' : 'translate-y-10 scale-95'
          ]" 
          class="relative w-full max-w-md rounded-2xl overflow-hidden mb-4 sm:mb-0 flex flex-col max-h-[90vh] pointer-events-auto transition-all duration-500 cubic-bezier">
        
        <view class="p-6 pb-2 shrink-0">
          <view class="text-xl font-bold" :class="isDarkMode ? 'text-stone-100' : 'text-stone-800'">AI 精简设置</view>
          <view class="text-sm mt-1 text-stone-500">《{{ bookTitle }}》{{ chapterTitle }}</view>
        </view>

        <view class="p-6 space-y-3 overflow-y-auto">
          <view v-for="prompt in prompts" :key="prompt.id"
            @click="selectedId = prompt.id"
            :class="selectedId === prompt.id 
              ? (isDarkMode ? 'border-teal-600 bg-teal-900/20 ring-1 ring-teal-600' : 'border-teal-500 bg-teal-50 ring-1 ring-teal-500') 
              : (isDarkMode ? 'border-stone-800 bg-stone-900/50 hover:border-stone-700' : 'border-stone-200 hover:border-teal-200')"
            class="flex items-start gap-3 p-3 border rounded-xl cursor-pointer transition-all relative overflow-hidden">
            
            <view v-if="trimmedIds?.includes(prompt.id)" class="absolute top-0 right-0 bg-teal-500 text-white text-[10px] px-2 py-0.5 rounded-bl-lg font-bold">
              已缓存
            </view>

            <view>
              <view class="font-bold text-sm" :class="isDarkMode ? 'text-stone-200' : 'text-stone-800'">{{ prompt.name }}</view>
              <view class="text-xs mt-1 leading-relaxed text-stone-500">{{ prompt.description || '暂无描述' }}</view>
            </view>
          </view>
        </view>

        <view :class="isDarkMode ? 'bg-stone-900 border-stone-800' : 'bg-stone-50 border-stone-100'" class="p-4 border-t flex gap-3 shrink-0">
          <view @click="emit('close')" :class="isDarkMode ? 'text-stone-500 bg-stone-800' : 'text-stone-500 bg-stone-200'" class="flex-1 py-3 font-medium text-sm rounded-xl text-center">稍后</view>
          <view @click="emit('start', selectedId)" class="flex-1 py-3 bg-stone-900 text-white font-medium text-sm rounded-xl shadow-lg text-center active:scale-95 transition-transform">
            {{ trimmedIds?.includes(Number(selectedId)) ? '开始阅读' : '开始精简' }}
          </view>
        </view>
    </view>
  </view>
</template>

<style scoped>
.cubic-bezier { transition-timing-function: cubic-bezier(0.16, 1, 0.3, 1); }
</style>
