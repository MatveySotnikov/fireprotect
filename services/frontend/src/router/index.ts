import { createRouter, createWebHistory } from 'vue-router';
import { useAuthStore } from '../stores/auth';

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('../views/LoginView.vue'),
    },
    {
      path: '/register',
      name: 'register',
      component: () => import('../views/RegisterView.vue'),
    },
    {
      path: '/',
      component: () => import('../components/layout/AppLayout.vue'),
      meta: { requiresAuth: true },
      children: [
        {
          path: '',
          name: 'calculator',
          component: () => import('../views/CalculatorView.vue'),
        },
        {
          path: 'history',
          name: 'history',
          component: () => import('../views/HistoryView.vue'),
        },
        {
          path: 'calculations/:id',
          name: 'calculation-detail',
          component: () => import('../views/CalculationDetailView.vue'),
        },
      ],
    },
  ],
});

router.beforeEach((to) => {
  const authStore = useAuthStore();
  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    return '/login';
  }
  if (authStore.isAuthenticated && (to.path === '/login' || to.path === '/register')) {
    return '/';
  }
  return true;
});

export default router;