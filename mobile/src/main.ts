import { createSSRApp } from "vue";
import * as Pinia from 'pinia';
import App from "./App.vue";
import '@/style/tailwind.css';

export function createApp() {
  const app = createSSRApp(App);
  
  app.use(Pinia.createPinia());
  
  return {
    app,
    Pinia, // Uni-app 推荐导出 Pinia
  };
}