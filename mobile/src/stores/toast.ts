import { defineStore } from "pinia";
import { ref } from "vue";

export type ToastType = "success" | "error" | "info";

// Toast 提示状态管理。
export const useToastStore = defineStore("toast", () => {
  const visible = ref(false);
  const message = ref("");
  const type = ref<ToastType>("info");
  const duration = ref(2000);
  let timer: ReturnType<typeof setTimeout> | null = null;

  // 显示提示。
  const show = (payload: { message: string; type?: ToastType; duration?: number }) => {
    message.value = payload.message;
    type.value = payload.type || "info";
    duration.value = payload.duration ?? 2000;
    visible.value = true;
    if (timer) {
      clearTimeout(timer);
    }
    timer = setTimeout(() => {
      visible.value = false;
    }, duration.value);
  };

  // 关闭提示。
  const hide = () => {
    visible.value = false;
    if (timer) {
      clearTimeout(timer);
      timer = null;
    }
  };

  return {
    visible,
    message,
    type,
    duration,
    show,
    hide,
  };
});
