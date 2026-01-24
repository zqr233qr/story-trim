import { defineStore } from "pinia";
import { ref } from "vue";

// 全局加载状态管理。
export const useLoadingStore = defineStore("loading", () => {
  const visible = ref(false);

  // 显示加载。
  const show = () => {
    visible.value = true;
  };

  // 隐藏加载。
  const hide = () => {
    visible.value = false;
  };

  return {
    visible,
    show,
    hide,
  };
});
