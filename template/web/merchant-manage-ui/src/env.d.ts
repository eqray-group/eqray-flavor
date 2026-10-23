/// <reference types="vite/client" />

declare module '*.vue' {
  import type { DefineComponent } from 'vue'

  const component: DefineComponent<object, object, unknown>
  export default component
}

// 样式按副作用导入，显式声明模块避免 TS 报找不到类型
declare module '*.css'
declare module 'element-plus/dist/index.css'
