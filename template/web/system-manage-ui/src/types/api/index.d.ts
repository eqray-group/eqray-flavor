/**
 * API 类型入口。
 *
 * 类型按模块拆在同目录：common / auth / infra / store / member / dish / dashboard / system。
 * 每个文件都用 `declare namespace Api { namespace Xxx { ... } }` 声明，TypeScript 会自动合并到
 * 全局命名空间 Api，业务代码直接写 Api.Xxx.Yyy，无需 import，也不要在文件顶部加 export。
 *
 * 新增模块：新建 <模块>.d.ts，照抄任意现有文件的头部，替换命名空间名即可。
 * 命名约定：system 模块的命名空间是 Api.SystemManage，不是 Api.System。
 * 雪花 ID 一律 string；金额字段在类型里注明是分还是元。
 */
