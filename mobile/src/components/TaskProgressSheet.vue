<template>
  <view v-if="modelValue" class="fixed inset-0 z-[200] flex flex-col justify-end" @touchmove.stop.prevent>
    <!-- Backdrop -->
    <view 
      class="absolute inset-0 bg-black/40 backdrop-blur-sm transition-opacity duration-300"
      @click="close"
    ></view>

    <!-- Sheet Content -->
    <view 
      class="bg-white rounded-t-3xl relative z-10 w-full max-h-[80vh] flex flex-col transition-transform duration-300 transform translate-y-0"
      :class="{ 'translate-y-full': !animateShow }"
    >
      <!-- Header -->
      <view class="flex items-center justify-between px-6 pt-5 pb-4 border-b border-stone-100 shrink-0">
        <view>
          <text class="text-xl font-bold text-stone-800">任务中心</text>
          <text class="text-xs text-stone-400 block mt-1">{{ tasks.length > 0 ? 'AI 正在全力处理中' : '暂无进行中的任务' }}</text>
        </view>
        <view 
          @click="close"
          class="w-8 h-8 rounded-full bg-stone-100 flex items-center justify-center active:bg-stone-200 text-stone-500"
        >
          <text class="font-bold">×</text>
        </view>
      </view>

      <!-- Task List -->
      <scroll-view scroll-y class="flex-1 w-full p-6 overflow-hidden box-border">
        <view class="space-y-4 pb-10">
          <view 
            v-for="task in tasks" 
            :key="task.id"
            class="bg-stone-50 rounded-2xl p-4 border border-stone-100 w-full"
          >
            <!-- Task Header -->
            <view class="flex justify-between items-start mb-3 overflow-hidden">
              <view class="flex items-center gap-2 overflow-hidden">
                <text class="font-bold text-stone-800 truncate">{{ task.book_title }}</text>
                <view class="bg-teal-100 px-1.5 py-0.5 rounded text-[10px] text-teal-700 font-bold shrink-0">
                  {{ task.prompt_name || task.status }}
                </view>
              </view>
              <text class="text-sm font-mono font-bold text-teal-600 shrink-0 ml-2">
                {{ task.progress }}%
              </text>
            </view>

            <!-- Progress Bar -->
            <view class="h-1.5 w-full bg-stone-200 rounded-full overflow-hidden mb-3">
              <view 
                class="h-full bg-teal-500 transition-all duration-300 ease-out"
                :style="{ width: task.progress + '%' }"
              ></view>
            </view>

            <!-- Logs / Stage -->
            <view class="bg-white/50 rounded-lg p-2.5 flex items-start gap-2">
              <text class="text-[10px] text-stone-400 leading-relaxed shrink-0">></text>
              <text class="text-[10px] text-stone-500 font-mono leading-relaxed line-clamp-2">
                {{ getTaskStatusText(task) }}
              </text>
            </view>
          </view>

          <!-- Empty State -->
          <view v-if="tasks.length === 0" class="py-10 text-center">
            <text class="text-stone-300 text-sm">暂无进行中的任务</text>
          </view>
        </view>
      </scroll-view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, watch, onUnmounted } from 'vue';
import { taskApi } from '@/api/task';

interface TaskItem {
  id: string
  book_id: number
  book_title: string
  prompt_id: number
  prompt_name: string
  status: string
  progress: number
  error?: string
  created_at: string
}

const props = defineProps<{
  modelValue: boolean;
}>();

const emit = defineEmits(['update:modelValue', 'update:tasks']);

const animateShow = ref(false);
const tasks = ref<TaskItem[]>([]);
let pollTimer: ReturnType<typeof setInterval> | null = null;

const fetchTasks = async () => {
  try {
    const res = await taskApi.getActiveTasks();
    if (res.code === 0) {
      tasks.value = (res.data || []) as TaskItem[];
      emit('update:tasks', tasks.value);
      
      // 如果没有任务了，关闭弹窗
      if (tasks.value.length === 0) {
        close();
      }
    }
  } catch (e) {
    console.warn('[TaskProgress] Fetch tasks failed', e);
  }
};

const startPoll = () => {
  fetchTasks();
  pollTimer = setInterval(fetchTasks, 3000);
};

const stopPoll = () => {
  if (pollTimer) {
    clearInterval(pollTimer);
    pollTimer = null;
  }
};

watch(() => props.modelValue, (val) => {
  if (val) {
    setTimeout(() => {
      animateShow.value = true;
    }, 10);
    startPoll();
  } else {
    animateShow.value = false;
    stopPoll();
  }
});

onUnmounted(() => {
  stopPoll();
});

const close = () => {
  animateShow.value = false;
  setTimeout(() => {
    emit('update:modelValue', false);
  }, 300);
};

const getTaskStatusText = (task: TaskItem) => {
  if (task.status === 'running') {
    return `正在精简处理中... (进度: ${task.progress}%)`;
  }
  if (task.status === 'pending') {
    return '等待处理中...';
  }
  if (task.error) {
    return `处理失败: ${task.error}`;
  }
  return '准备就绪';
};
</script>
