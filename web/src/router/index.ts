import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory('/studio/'),
  routes: [
    { path: '/', redirect: '/posts' },
    { path: '/login', component: () => import('@/views/LoginView.vue'), meta: { public: true } },
    { path: '/posts', component: () => import('@/views/PostsView.vue') },
    { path: '/posts/:slug', component: () => import('@/views/EditorView.vue') },
    { path: '/posts/:slug/preview', component: () => import('@/views/PreviewView.vue') },
    { path: '/home', component: () => import('@/views/HomeView.vue') },
    { path: '/now', component: () => import('@/views/NowView.vue') },
    {
      path: '/settings',
      component: () => import('@/views/settings/SettingsLayout.vue'),
      children: [
        { path: '', redirect: 'general' },
        { path: 'general', component: () => import('@/views/settings/GeneralTab.vue') },
        { path: 'writing', component: () => import('@/views/settings/WritingTab.vue') },
        { path: 'wechat', component: () => import('@/views/settings/WechatTab.vue') },
        { path: 'security', component: () => import('@/views/settings/SecurityTab.vue') },
        { path: 'audit', component: () => import('@/views/settings/AuditTab.vue') },
      ]
    },
    { path: '/health', component: () => import('@/views/HealthView.vue') },
    { path: '/trash', component: () => import('@/views/TrashView.vue') },
  ]
})

export default router
