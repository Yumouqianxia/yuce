import type { RoutePermissionMeta } from './permissions'

declare module 'vue-router' {
  interface RouteMeta extends RoutePermissionMeta {}
}

