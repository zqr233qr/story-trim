<script setup lang="ts">
import { onLaunch, onShow, onHide } from "@dcloudio/uni-app";
onLaunch(() => {
  console.log("App Launch");
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
