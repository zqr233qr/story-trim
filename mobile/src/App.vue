<script setup lang="ts">
import { onLaunch, onShow, onHide } from "@dcloudio/uni-app";
import { useBookStore } from "./stores/book";
import { useNetworkStore } from "./stores/network";



const bookStore = useBookStore();
const networkStore = useNetworkStore();


onLaunch(async () => {
  console.log("App Launch");
  // 初始化本地数据库
  // #ifdef APP-PLUS
  await bookStore.init();
  // #endif
  networkStore.startPing();
});
onShow(() => {
  console.log("App Show");
  networkStore.startPing();
});
onHide(() => {
  console.log("App Hide");
  networkStore.stopPing();
});

</script>

<style>
@import "@/style/tailwind.css";

/* 解决小程序中一些基础标签的样式问题 */
view, text {
  box-sizing: border-box;
}
</style>

