import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
// 导入自定义全局样式覆盖
import './assets/styles.css'
import App from './App.vue'
import router from './router'

// 创建Vue应用实例
const app = createApp(App)

// 使用Pinia状态管理
app.use(createPinia())

// 使用路由
app.use(router)

// 使用Element Plus
app.use(ElementPlus, {
  // 在这里设置ElementPlus的全局配置
  zIndex: 9999
})

// 挂载应用
app.mount('#app') 