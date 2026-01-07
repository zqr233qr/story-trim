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
  prompts: Prompt[],
  isDarkMode?: boolean
}>()
const emit = defineEmits(['close', 'confirm'])

const selectedId = ref<number | string>('')

watch(() => props.prompts, (newPs) => {
  if (newPs.length > 0 && !selectedId.value) {
    selectedId.value = newPs[0].id
  }
}, { immediate: true })
</script>

<template>
  <view v-if="show" class="fixed inset-0 z-[100] flex items-center justify-center bg-black/40 backdrop-blur-sm p-4">
    <view :class="isDarkMode ? 'bg-stone-900 border border-stone-800' : 'bg-white shadow-xl'" 
          class="w-full max-w-sm rounded-3xl p-6 relative">
      <view class="text-lg font-bold mb-2 text-center" :class="isDarkMode ? 'text-stone-100' : 'text-stone-800'">全书后台处理</view>
      <view class="text-[10px] mb-6 text-center text-stone-400">将对《{{ bookTitle }}》进行批量精简。处理将在后台进行，您可以继续阅读。</view>

      <view class="space-y-3 mb-8 max-h-[300px] overflow-y-auto pr-1">
        <view v-for="prompt in prompts" :key="prompt.id"
          @click="selectedId = prompt.id"
          :class="selectedId === prompt.id 
            ? (isDarkMode ? 'border-teal-600 bg-teal-900/20 text-teal-500' : 'border-teal-500 bg-teal-50 text-teal-700') 
            : (isDarkMode ? 'border-stone-800 bg-stone-900/50 text-stone-500' : 'border-stone-200 text-stone-600')"
          class="flex items-center gap-3 p-4 border rounded-2xl transition-all">
          <view class="flex-1">
             <view class="text-sm font-bold">{{ prompt.name }}</view>
             <view class="text-[10px] opacity-70 mt-1">{{ prompt.description || '暂无描述' }}</view>
          </view>
        </view>
      </view>

      <view class="flex gap-3">
        <button @click="emit('close')" :class="isDarkMode ? 'bg-stone-800 text-stone-400' : 'bg-stone-100 text-stone-500'" class="flex-1 h-12 rounded-xl text-sm font-bold flex items-center justify-center">取消</button>
        <button @click="emit('confirm', selectedId)" class="flex-1 h-12 bg-stone-900 text-white rounded-xl text-sm font-bold shadow-lg flex items-center justify-center">启动任务</button>
      </view>
    </view>
  </view>
</template>