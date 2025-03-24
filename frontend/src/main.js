import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import './style.css'
import axios from 'axios'
import { create, NButton, NInput, NForm, NFormItem, NCard, NSpace, NModal, NLayout, NLayoutSider, NMenu, NIcon, NLayoutHeader, NLayoutContent, NDataTable, NPopconfirm, NSwitch, NPagination, NDrawer, NDrawerContent, NSelect, NTag, NMessageProvider } from 'naive-ui'

// 创建Naive UI
const naive = create({
  components: [
    NButton,
    NInput,
    NForm,
    NFormItem,
    NCard,
    NSpace,
    NModal,
    NLayout,
    NLayoutSider,
    NMenu,
    NIcon,
    NLayoutHeader,
    NLayoutContent,
    NDataTable,
    NPopconfirm,
    NSwitch,
    NPagination,
    NDrawer,
    NDrawerContent,
    NSelect,
    NTag,
    NMessageProvider
  ]
})

const app = createApp(App)

// 注册Naive UI
app.use(naive)
app.use(router)

// 配置axios
axios.defaults.baseURL = '/api'

// 自定义全局属性
app.config.globalProperties.$axios = axios

// 挂载应用
app.mount('#app') 