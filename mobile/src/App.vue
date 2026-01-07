<script setup lang="ts">
import { onLaunch, onShow, onHide } from "@dcloudio/uni-app";
import { useBookStore } from "./stores/book";

const bookStore = useBookStore();

onLaunch(async () => {
  console.log("App Launch");
  // 初始化本地数据库
  // #ifdef APP-PLUS
  await bookStore.init();
  // #endif
});
onShow(() => {
  console.log("App Show");
});
onHide(() => {
  console.log("App Hide");
});
</script>
<style>
@import "@/style/tailwind.css";

/* 解决小程序中一些基础标签的样式问题 */
view, text {
  box-sizing: border-box;
}
</style>
