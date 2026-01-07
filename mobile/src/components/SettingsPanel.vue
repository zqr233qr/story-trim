<script setup lang="ts">
import { ref, computed } from 'vue'

interface Prompt {
  id: number;
  name: string;
}

const props = defineProps<{
  show: boolean,
  modes: string[], 
  prompts: Prompt[], 
  activeMode: string,
  fontSize: number,
  isDarkMode: boolean,
  pageMode: 'scroll' | 'click'
}>()

const emit = defineEmits(['close', 'update:activeMode', 'update:fontSize', 'update:isDarkMode', 'update:pageMode', 'addMode'])

const getModeName = (id: string) => {
  if (id === 'original') return '原文'
  const prompt = props.prompts.find(p => p.id.toString() === id)
  return prompt ? prompt.name : id
}
</script>

<template>
  <view class="fixed inset-0 z-[100] flex items-end justify-center pointer-events-none transition-all duration-300"
        :class="show ? 'opacity-100' : 'opacity-0 invisible'">
    
    <!-- Backdrop -->
    <view @click.stop="emit('close')" class="absolute inset-0 bg-black/40 backdrop-blur-[1px] pointer-events-auto"></view>

    <!-- Panel -->
    <view :class="[
            isDarkMode ? 'bg-stone-900 border-stone-800' : 'bg-white',
            show ? 'translate-y-0' : 'translate-y-full'
          ]" 
          class="relative w-full max-w-lg rounded-t-3xl shadow-2xl p-6 pb-10 z-10 pointer-events-auto transition-transform duration-500 cubic-bezier">
      
      <view class="w-12 h-1 bg-stone-200 rounded-full mx-auto mb-6"></view>

      <!-- Mode Switcher -->
      <view class="mb-8">
        <view class="text-[10px] font-bold text-stone-400 uppercase tracking-widest mb-3">AI 阅读层</view>
        <scroll-view scroll-x class="w-full whitespace-nowrap">
          <view v-for="modeKey in modes" :key="modeKey"
            @click="emit('update:activeMode', modeKey)"
            :class="activeMode === modeKey ? 'bg-teal-500 text-white border-transparent shadow-md' : (isDarkMode ? 'bg-stone-800 text-stone-400 border-stone-700' : 'bg-white text-stone-600 border-stone-200')"
            class="px-4 py-2 rounded-full text-xs font-medium border inline-block mr-3 transition-all">
            {{ getModeName(modeKey) }}
          </view>
        </scroll-view>
      </view>

      <!-- Font Size -->
      <view class="mb-8">
        <view class="text-[10px] font-bold text-stone-400 uppercase tracking-widest mb-3">字号</view>
        <view :class="isDarkMode ? 'bg-stone-800' : 'bg-stone-50'" class="flex items-center justify-between rounded-xl p-3">
          <text class="text-sm px-2 text-stone-500">A</text>
          <slider 
            :value="fontSize" 
            @change="emit('update:fontSize', $event.detail.value)"
            min="14" max="30" step="1" 
            class="flex-1 mx-4"
            activeColor="#0d9488"
            block-size="20"
          />
          <text class="text-xl px-2 font-bold" :class="isDarkMode ? 'text-stone-300' : 'text-stone-800'">A</text>
        </view>
      </view>

      <!-- Page Mode -->
      <view class="mb-8">
         <view class="text-[10px] font-bold text-stone-400 uppercase tracking-widest mb-3">翻页模式</view>
         <view :class="isDarkMode ? 'bg-stone-800' : 'bg-stone-50'" class="rounded-xl p-1 flex">
           <view @click="emit('update:pageMode', 'scroll')" 
             :class="pageMode === 'scroll' ? (isDarkMode ? 'bg-stone-700 text-stone-100' : 'bg-white shadow-sm text-stone-800') : 'text-stone-400'"
             class="flex-1 py-2 rounded-lg text-xs font-bold text-center transition-all">
             滚动
           </view>
           <view @click="emit('update:pageMode', 'click')" 
             :class="pageMode === 'click' ? (isDarkMode ? 'bg-stone-700 text-stone-100' : 'bg-white shadow-sm text-stone-800') : 'text-stone-400'"
             class="flex-1 py-2 rounded-lg text-xs font-bold text-center transition-all">
             点击
           </view>
         </view>
      </view>

      <!-- Dark Mode Toggle -->
      <view class="flex items-center justify-between">
         <text class="text-[10px] font-bold text-stone-400 uppercase tracking-widest">夜间模式</text>
         <switch :checked="isDarkMode" @change="emit('update:isDarkMode', $event.detail.value)" color="#0d9488" />
      </view>
    </view>
  </view>
</template>

<style scoped>
.cubic-bezier { transition-timing-function: cubic-bezier(0.16, 1, 0.3, 1); }
</style>