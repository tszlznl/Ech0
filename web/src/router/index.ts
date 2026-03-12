import { createRouter, createWebHistory } from 'vue-router'
import { useInitStore, useUserStore } from '@/stores'

// 所有路由组件使用懒加载，优化首屏加载性能
const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: () => import('../views/home/HomeView.vue'),
      meta: {
        title: 'Home',
        optionalAuth: true,
      },
    },
    {
      path: '/panel',
      name: 'panel',
      component: () => import('../views/panel/PanelView.vue'),
      redirect: '/panel/dashboard',
      meta: {
        requiresAuth: true,
      },
      children: [
        {
          path: 'dashboard',
          name: 'panel-dashboard',
          component: () => import('../views/panel/modules/TheDashboard.vue'),
        },
        {
          path: 'setting',
          name: 'panel-setting',
          component: () => import('../views/panel/modules/TheSetting.vue'),
        },
        {
          path: 'user',
          name: 'panel-user',
          component: () => import('../views/panel/modules/TheUser.vue'),
        },
        {
          path: 'storage',
          name: 'panel-storage',
          component: () => import('../views/panel/modules/TheStorage.vue'),
        },
        {
          path: 'data-management',
          name: 'panel-data-management',
          component: () => import('../views/panel/modules/TheDataManagement.vue'),
        },
        {
          path: 'sso',
          name: 'panel-sso',
          component: () => import('../views/panel/modules/TheSSO.vue'),
        },
        {
          path: 'extension',
          name: 'panel-extension',
          component: () => import('../views/panel/modules/TheExtension.vue'),
        },
        {
          path: 'comment',
          name: 'panel-comment',
          component: () => import('../views/panel/modules/TheCommentManager.vue'),
        },
        {
          path: 'advance',
          name: 'panel-advance',
          component: () => import('../views/panel/modules/TheAdvance.vue'),
        },
        {
          path: 'system-log',
          name: 'panel-system-log',
          component: () => import('../views/panel/modules/TheSystemLog.vue'),
        },
      ],
      // beforeEnter: (to, from, next) => {
      //   const userStore = useUserStore()
      //   if (userStore.isLogin) {
      //     next()
      //   } else {
      //     next({ name: 'auth' })
      //   }
      // },
    },
    {
      path: '/auth',
      name: 'auth',
      component: () => import('../views/auth/AuthView.vue'),
    },
    {
      path: '/widget',
      name: 'widget',
      component: () => import('../views/widget/WidgetView.vue'),
    },
    {
      path: '/init',
      name: 'init',
      component: () => import('../views/init/InitView.vue'),
    },
    {
      path: '/connect',
      name: 'connect',
      component: () => import('../views/connect/ConnectView.vue'),
    },
    {
      path: '/hub',
      name: 'hub',
      component: () => import('../views/hub/HubView.vue'),
    },
    {
      path: '/zone/:echoId?',
      name: 'zone',
      component: () => import('../views/zone/ZoneView.vue'),
    },
    {
      path: '/echo/:echoId',
      name: 'echo',
      component: () => import('../views/echo/EchoView.vue'),
    },
    {
      path: '/:pathMatch(.*)*',
      name: 'not-found',
      component: () => import('../views/404/NotFoundView.vue'),
    },
  ],
})

// 全局路由守卫
router.beforeEach(async (to) => {
  const initStore = useInitStore()
  const userStore = useUserStore()

  if (!initStore.ready) {
    await initStore.init()
  }

  if (!initStore.initialized && to.name !== 'init') {
    return { name: 'init' }
  }

  if (initStore.initialized && to.name === 'init') {
    return { name: 'auth' }
  }

  // 等待用户信息初始化完成
  if (!userStore.initialized) {
    await userStore.init()
  }

  const token = localStorage.getItem('token')
  const needRedirect = localStorage.getItem('needLoginRedirect')

  //  强制鉴权页面或 token 无效
  if (
    (to.meta.requiresAuth && !userStore.isLogin) || // 需要鉴权但未登录
    (to.meta.optionalAuth && token && !userStore.isLogin && needRedirect === 'true') // 可选鉴权且有token但未登录且需要重定向
  ) {
    localStorage.removeItem('needLoginRedirect')
    localStorage.removeItem('token')
    return { name: 'auth' }
  }

  return true
})

export default router
