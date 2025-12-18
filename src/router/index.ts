import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { usePermissionStore } from '@/stores/permissions'
import { useAdminStore } from '@/stores/admin'
import { ADMIN_PERMISSIONS } from '@/types/admin'
import { ElMessage } from 'element-plus'

// 定义路由
const routes: Array<RouteRecordRaw> = [
  {
    path: '/',
    name: 'Home',
    component: () => import('@/views/Home.vue'),
    meta: { title: '首页' }
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/auth/Login.vue'),
    meta: { title: '登录', guest: true }
  },
  {
    path: '/register',
    name: 'Register',
    component: () => import('@/views/auth/Register.vue'),
    meta: { title: '注册', guest: true }
  },
  {
    path: '/matches',
    name: 'Matches',
    component: () => import('@/views/matches/MatchList.vue'),
    meta: { title: '比赛列表' },
    redirect: { name: 'TodayMatches' },
    children: [
      {
        path: 'today',
        name: 'TodayMatches',
        component: () => import('@/views/matches/TodayMatches.vue'),
        meta: { title: '今日赛程' }
      },
      {
        path: 'upcoming',
        name: 'UpcomingMatches',
        component: () => import('@/views/matches/UpcomingMatches.vue'),
        meta: { title: '未来赛程' }
      },
      {
        path: 'all',
        name: 'AllMatches',
        component: () => import('@/views/matches/AllMatches.vue'),
        meta: { title: '所有比赛 (调试)' }
      },
      {
        path: 'history',
        name: 'HistoryMatches',
        component: () => import('@/views/matches/HistoryMatches.vue'),
        meta: { title: '历史比赛' }
      }
    ]
  },
  // 春季赛
  {
    path: '/spring',
    name: 'SpringTournament',
    component: () => import('@/views/matches/MatchList.vue'),
    meta: { title: 'KPL春季赛', tournamentType: 'spring', year: 2025 },
    redirect: { name: 'SpringTodayMatches' },
    children: [
      {
        path: 'today',
        name: 'SpringTodayMatches',
        component: () => import('@/views/matches/TodayMatches.vue'),
        meta: { title: '春季赛今日赛程', tournamentType: 'spring', year: 2025 }
      },
      {
        path: 'upcoming',
        name: 'SpringUpcomingMatches',
        component: () => import('@/views/matches/UpcomingMatches.vue'),
        meta: { title: '春季赛未来赛程', tournamentType: 'spring', year: 2025 }
      },
      {
        path: 'history',
        name: 'SpringHistoryMatches',
        component: () => import('@/views/matches/HistoryMatches.vue'),
        meta: { title: '春季赛历史比赛', tournamentType: 'spring', year: 2025 }
      }
    ]
  },
  // 夏季赛
  {
    path: '/summer',
    name: 'SummerTournament',
    component: () => import('@/views/matches/MatchList.vue'),
    meta: { title: 'KPL夏季赛', tournamentType: 'summer', year: 2025 },
    redirect: { name: 'SummerTodayMatches' },
    children: [
      {
        path: 'today',
        name: 'SummerTodayMatches',
        component: () => import('@/views/matches/TodayMatches.vue'),
        meta: { title: '夏季赛今日赛程', tournamentType: 'summer', year: 2025 }
      },
      {
        path: 'upcoming',
        name: 'SummerUpcomingMatches',
        component: () => import('@/views/matches/UpcomingMatches.vue'),
        meta: { title: '夏季赛未来赛程', tournamentType: 'summer', year: 2025 }
      },
      {
        path: 'history',
        name: 'SummerHistoryMatches',
        component: () => import('@/views/matches/HistoryMatches.vue'),
        meta: { title: '夏季赛历史比赛', tournamentType: 'summer', year: 2025 }
      }
    ]
  },
  // 年度总决赛
  {
    path: '/annual',
    name: 'AnnualTournament',
    component: () => import('@/views/matches/MatchList.vue'),
    meta: { title: 'KPL年度总决赛', tournamentType: 'annual', year: 2025 },
    redirect: { name: 'AnnualTodayMatches' },
    children: [
      {
        path: 'today',
        name: 'AnnualTodayMatches',
        component: () => import('@/views/matches/TodayMatches.vue'),
        meta: { title: '年度总决赛今日赛程', tournamentType: 'annual', year: 2025 }
      },
      {
        path: 'upcoming',
        name: 'AnnualUpcomingMatches',
        component: () => import('@/views/matches/UpcomingMatches.vue'),
        meta: { title: '年度总决赛未来赛程', tournamentType: 'annual', year: 2025 }
      },
      {
        path: 'history',
        name: 'AnnualHistoryMatches',
        component: () => import('@/views/matches/HistoryMatches.vue'),
        meta: { title: '年度总决赛历史比赛', tournamentType: 'annual', year: 2025 }
      }
    ]
  },
  // 挑战者杯
  {
    path: '/challenger',
    name: 'ChallengerTournament',
    component: () => import('@/views/matches/MatchList.vue'),
    meta: { title: 'KPL挑战者杯', tournamentType: 'challenger', year: 2025 },
    redirect: { name: 'ChallengerTodayMatches' },
    children: [
      {
        path: 'today',
        name: 'ChallengerTodayMatches',
        component: () => import('@/views/matches/TodayMatches.vue'),
        meta: { title: '挑战者杯今日赛程', tournamentType: 'challenger', year: 2025 }
      },
      {
        path: 'upcoming',
        name: 'ChallengerUpcomingMatches',
        component: () => import('@/views/matches/UpcomingMatches.vue'),
        meta: { title: '挑战者杯未来赛程', tournamentType: 'challenger', year: 2025 }
      },
      {
        path: 'history',
        name: 'ChallengerHistoryMatches',
        component: () => import('@/views/matches/HistoryMatches.vue'),
        meta: { title: '挑战者杯历史比赛', tournamentType: 'challenger', year: 2025 }
      }
    ]
  },
  {
    path: '/matches/:id',
    name: 'MatchDetail',
    component: () => import('@/views/matches/MatchDetail.vue'),
    meta: { title: '比赛详情' }
  },
  {
    path: '/upcoming-matches',
    name: 'PredictionUpcomingMatches',
    component: () => import('@/views/predictions/UpcomingMatches.vue'),
    meta: { title: '即将开始的比赛' }
  },
  {
    path: '/prediction-history',
    name: 'PredictionHistory',
    component: () => import('@/views/predictions/PredictionHistory.vue'),
    meta: { title: '我的预测历史', requiresAuth: true }
  },
  // 积分排行榜路由
  {
    path: '/leaderboard',
    name: 'Leaderboard',
    component: () => import('@/views/leaderboard/LeaderboardPage.vue'),
    meta: { title: '积分排行榜' },
    redirect: { name: 'SummerLeaderboard' },
    children: []
  },
  // 春季赛积分榜
  {
    path: '/leaderboard/spring',
    name: 'SpringLeaderboard',
    component: () => import('@/views/leaderboard/LeaderboardPage.vue'),
    meta: { title: '2025KPL春季赛积分榜', tournamentType: 'spring' }
  },
  // 夏季赛积分榜
  {
    path: '/leaderboard/summer',
    name: 'SummerLeaderboard',
    component: () => import('@/views/leaderboard/LeaderboardPage.vue'),
    meta: { title: '2025KPL夏季赛积分榜', tournamentType: 'summer' }
  },
  // 年度总决赛积分榜
  {
    path: '/leaderboard/annual',
    name: 'AnnualLeaderboard',
    component: () => import('@/views/leaderboard/LeaderboardPage.vue'),
    meta: { title: '2025KPL年度总决赛积分榜', tournamentType: 'annual' }
  },
  // 挑战者杯积分榜
  {
    path: '/leaderboard/challenger',
    name: 'ChallengerLeaderboard',
    component: () => import('@/views/leaderboard/LeaderboardPage.vue'),
    meta: { title: '2025KPL挑战者杯积分榜', tournamentType: 'challenger' }
  },
  {
    path: '/prediction-rules',
    name: 'PredictionRules',
    component: () => import('@/views/predictions/PredictionRules.vue'),
    meta: { title: '积分规则' }
  },
  {
    path: '/profile',
    name: 'Profile',
    component: () => import('@/views/user/Profile.vue'),
    meta: { title: '个人资料', requiresAuth: true }
  },
  {
    path: '/profile/edit',
    name: 'ProfileEdit',
    component: () => import('@/views/user/ProfileEdit.vue'),
    meta: { title: '编辑个人资料', requiresAuth: true }
  },
  // 修改密码功能已集成到编辑资料页面
  {
    path: '/profile/points-history',
    name: 'PointsHistory',
    component: () => import('@/views/user/PointsHistory.vue'),
    meta: { title: '积分历史', requiresAuth: true }
  },
  // 管理员路由
  {
    path: '/admin',
    name: 'AdminDashboard',
    component: () => import('@/views/admin/AdminDashboard.vue'),
    meta: { 
      title: '管理中心', 
      requiresAuth: true, 
      requiresAdmin: true,
      permissions: []
    },
    children: [
      {
        path: '',
        name: 'AdminDashboardHome',
        redirect: { name: 'AdminSite' }
      },
      {
        path: 'site',
        name: 'AdminSite',
        component: () => import('@/views/admin/AdminSite.vue'),
        meta: { 
          title: '公告管理', 
          requiresAuth: true, 
          requiresAdmin: true,
          permissions: [ADMIN_PERMISSIONS.SYSTEM_CONFIG]
        }
      },
      {
        path: 'users',
        name: 'AdminUsers',
        component: () => import('@/views/admin/AdminUsers.vue'),
        meta: { 
          title: '用户管理', 
          requiresAuth: true, 
          requiresAdmin: true,
          permissions: [ADMIN_PERMISSIONS.USER_MANAGE]
        }
      },
      {
        path: 'matches',
        name: 'AdminMatches',
        component: () => import('@/views/admin/AdminMatches.vue'),
        meta: { 
          title: '比赛管理', 
          requiresAuth: true, 
          requiresAdmin: true,
          permissions: [ADMIN_PERMISSIONS.MATCH_MANAGE]
        }
      },
      {
        path: 'sport-types',
        name: 'AdminSportTypes',
        component: () => import('@/views/admin/AdminSportTypes.vue'),
        meta: { 
          title: '运动类型管理', 
          requiresAuth: true, 
          requiresAdmin: true,
          permissions: [ADMIN_PERMISSIONS.SPORT_TYPE_MANAGE]
        }
      },
      {
        path: 'scoring-rules',
        name: 'AdminScoringRules',
        component: () => import('@/views/admin/AdminScoringRules.vue'),
        meta: { 
          title: '积分规则管理', 
          requiresAuth: true, 
          requiresAdmin: true,
          permissions: [ADMIN_PERMISSIONS.SCORING_RULE_MANAGE]
        }
      },
      {
        path: 'admins',
        name: 'AdminManagement',
        component: () => import('@/views/admin/AdminManagement.vue'),
        meta: { 
          title: '管理员管理', 
          requiresAuth: true, 
          requiresAdmin: true,
          permissions: [ADMIN_PERMISSIONS.ADMIN_MANAGE],
          requiresSuperAdmin: true
        }
      },
      {
        path: 'audit-logs',
        name: 'AdminAuditLogs',
        component: () => import('@/views/admin/AdminAuditLogs.vue'),
        meta: { 
          title: '审计日志', 
          requiresAuth: true, 
          requiresAdmin: true,
          permissions: [ADMIN_PERMISSIONS.AUDIT_LOG_VIEW]
        }
      },
      {
        path: 'settings',
        name: 'AdminSettings',
        component: () => import('@/views/admin/AdminSettings.vue'),
        meta: { 
          title: '系统设置', 
          requiresAuth: true, 
          requiresAdmin: true,
          permissions: [ADMIN_PERMISSIONS.SYSTEM_CONFIG],
          requiresSuperAdmin: true
        }
      }
    ]
  },
  // 404页面
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: () => import('@/views/NotFound.vue'),
    meta: { title: '页面未找到' }
  }
]

// 创建路由实例
const router = createRouter({
  history: createWebHistory(),
  routes
})

// 导入守卫函数
import { combinedGuard } from './guards'

// 全局前置守卫
router.beforeEach(combinedGuard)

export default router