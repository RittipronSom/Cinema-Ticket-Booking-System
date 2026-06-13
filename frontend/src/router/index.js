import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import AdminView from '../views/AdminView.vue'
import { auth } from '../firebase'
import { onAuthStateChanged } from 'firebase/auth'

const ADMIN_EMAILS = ['foj3010@gmail.com'] // ใส่ email admin จริงๆ

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: HomeView,
    },
    {
      path: '/admin',
      name: 'admin',
      component: AdminView,
      meta: { requiresAdmin: true }
    },
  ],
})

router.beforeEach((to, from, next) => {
  if (to.meta.requiresAdmin) {
    // รอ Firebase โหลดเสร็จก่อน
    const unsubscribe = onAuthStateChanged(auth, (user) => {
      unsubscribe()
      console.log('user in guard:', user?.email)
      if (!user || !ADMIN_EMAILS.includes(user.email)) {
        next('/')
      } else {
        next()
      }
    })
  } else {
    next()
  }
})
export default router