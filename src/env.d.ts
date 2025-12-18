/// <reference types="vite/client" />

interface ImportMetaEnv {
  /** API基础URL */
  readonly VITE_API_BASE_URL: string
  /** 应用名称 */
  readonly VITE_APP_TITLE: string
  /** 应用环境 */
  readonly VITE_APP_ENV: 'development' | 'production' | 'test'
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}

declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<{}, {}, any>
  export default component
} 