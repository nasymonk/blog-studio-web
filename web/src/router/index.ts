import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory('/studio/'),
  routes: [
    { path: '/', redirect: '/posts' },
    { name: 'login', path: '/login', component: () => import('@/views/LoginView.vue'), meta: { public: true } },
    { name: 'posts', path: '/posts', component: () => import('@/views/PostsView.vue') },
    { name: 'editor', path: '/posts/:slug', component: () => import('@/views/EditorView.vue') },
    { name: 'preview', path: '/posts/:slug/preview', component: () => import('@/views/PreviewView.vue') },
    { name: 'home', path: '/home', component: () => import('@/views/HomeView.vue') },
    { name: 'now', path: '/now', component: () => import('@/views/NowView.vue') },
    {
      name: 'settings',
      path: '/settings',
      component: () => import('@/views/settings/SettingsLayout.vue'),
      children: [
        { path: '', redirect: { name: 'settings-general' } },
        { name: 'settings-general', path: 'general', component: () => import('@/views/settings/GeneralTab.vue') },
        { name: 'settings-writing', path: 'writing', component: () => import('@/views/settings/WritingTab.vue') },
        { name: 'settings-security', path: 'security', component: () => import('@/views/settings/SecurityTab.vue') },
        { name: 'settings-audit', path: 'audit', component: () => import('@/views/settings/AuditTab.vue') },
      ]
    },
    { name: 'health', path: '/health', component: () => import('@/views/HealthView.vue') },
    { name: 'trash', path: '/trash', component: () => import('@/views/TrashView.vue') },
    { name: 'tags', path: '/tags', component: () => import('@/views/TagsView.vue') },
  ]
})

export default router
