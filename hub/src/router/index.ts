import { createRouter, createWebHistory } from 'vue-router'
import { applyHubRouteMeta } from '../hubSeo'
import HomeView from '../views/HomeView.vue'
import ExploreView from '../views/ExploreView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: HomeView,
    },
    {
      path: '/explore',
      name: 'explore',
      component: ExploreView,
    },
  ],
})

router.afterEach((to) => {
  applyHubRouteMeta(to)
})

export default router
