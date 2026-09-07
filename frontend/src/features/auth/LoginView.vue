<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { NButton, NCard, NForm, NFormItem, NInput, NTabs, NTabPane, useMessage } from 'naive-ui'
import { useAuthStore } from '@/stores/auth'
import { ApiError } from '@/api/client'

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()
const message = useMessage()

const tab = ref<'login' | 'register'>('login')
const loading = ref(false)

const loginForm = reactive({ login: '', password: '' })
const registerForm = reactive({ username: '', email: '', display_name: '', password: '' })

function redirectTarget() {
  const r = route.query.redirect
  return typeof r === 'string' ? r : { name: 'dashboard' as const }
}

async function submitLogin() {
  if (!loginForm.login || !loginForm.password) {
    message.warning('Enter your username and password')
    return
  }
  loading.value = true
  try {
    await auth.login(loginForm.login, loginForm.password)
    router.push(redirectTarget())
  } catch (e) {
    message.error(e instanceof ApiError ? e.message : 'Login failed')
  } finally {
    loading.value = false
  }
}

async function submitRegister() {
  if (!registerForm.username || !registerForm.email || registerForm.password.length < 8) {
    message.warning('Username, email and a password of at least 8 characters are required')
    return
  }
  loading.value = true
  try {
    await auth.register({ ...registerForm })
    message.success('Account created')
    router.push(redirectTarget())
  } catch (e) {
    message.error(e instanceof ApiError ? e.message : 'Registration failed')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="wrap">
    <div class="hero">
      <div class="brand">
        <span class="brand__mark">PKB</span>
        <span class="brand__name">Problem Knowledge Base</span>
      </div>
      <p class="tagline">
        บันทึกปัญหา วิธีแก้ และขั้นตอน Troubleshooting เพื่อให้ทีมค้นเจอและนำกลับมาใช้ได้เร็ว
      </p>
    </div>

    <NCard class="card">
      <NTabs v-model:value="tab" justify-content="space-evenly" type="line">
        <NTabPane name="login" tab="Sign in">
          <NForm @submit.prevent="submitLogin">
            <NFormItem label="Username or email">
              <NInput v-model:value="loginForm.login" placeholder="admin" autofocus />
            </NFormItem>
            <NFormItem label="Password">
              <NInput
                v-model:value="loginForm.password"
                type="password"
                show-password-on="click"
                placeholder="••••••••"
                @keyup.enter="submitLogin"
              />
            </NFormItem>
            <NButton type="primary" block :loading="loading" attr-type="submit">Sign in</NButton>
          </NForm>
        </NTabPane>

        <NTabPane name="register" tab="Create account">
          <NForm @submit.prevent="submitRegister">
            <NFormItem label="Username">
              <NInput v-model:value="registerForm.username" placeholder="jdoe" />
            </NFormItem>
            <NFormItem label="Email">
              <NInput v-model:value="registerForm.email" placeholder="jdoe@example.com" />
            </NFormItem>
            <NFormItem label="Display name">
              <NInput v-model:value="registerForm.display_name" placeholder="Jane Doe" />
            </NFormItem>
            <NFormItem label="Password">
              <NInput
                v-model:value="registerForm.password"
                type="password"
                show-password-on="click"
                placeholder="at least 8 characters"
              />
            </NFormItem>
            <NButton type="primary" block :loading="loading" attr-type="submit">Create account</NButton>
          </NForm>
        </NTabPane>
      </NTabs>
    </NCard>
  </div>
</template>

<style scoped>
.wrap {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 24px;
  padding: 24px;
}
.hero {
  text-align: center;
  max-width: 420px;
}
.brand {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
}
.brand__mark {
  font-weight: 800;
  background: #4f46e5;
  color: #fff;
  border-radius: 6px;
  padding: 3px 7px;
}
.brand__name {
  font-size: 18px;
  font-weight: 700;
}
.tagline {
  margin-top: 10px;
  color: #6b7280;
  font-size: 13px;
  line-height: 1.6;
}
.card {
  width: 100%;
  max-width: 380px;
}
</style>
