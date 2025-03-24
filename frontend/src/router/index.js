import { createRouter, createWebHistory } from 'vue-router'
import PasswordManager from '../views/PasswordManager.vue'

const routes = [
  {
    path: '/',
    name: 'Home',
    component: PasswordManager,
    meta: {
      title: '007密码管理器'
    }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 更新页面标题
router.beforeEach((to, from, next) => {
  document.title = to.meta.title || '007密码管理器'
  next()
})

export default router 