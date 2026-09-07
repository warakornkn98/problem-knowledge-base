import { createApp } from 'vue'
import { createPinia } from 'pinia'
import naive from 'naive-ui'

import 'vfonts/Lato.css'
import 'vfonts/FiraCode.css'
import '@/assets/main.css'

import App from './App.vue'
import router from './router'
import { setUnauthorizedHandler } from '@/api/client'
import { useAuthStore } from '@/stores/auth'

const app = createApp(App)

app.use(createPinia())
app.use(router)
app.use(naive)

// Global 401 handling: drop the session and bounce to /login.
setUnauthorizedHandler(() => {
  const auth = useAuthStore()
  auth.clear()
  if (router.currentRoute.value.name !== 'login') {
    router.push({ name: 'login', query: { redirect: router.currentRoute.value.fullPath } })
  }
})

app.mount('#app')
