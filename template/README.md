# template · 前端工程

按**运行形态**分两层：`web/` 是 PC 浏览器端，`mobile/` 是 uni-app 移动端。
每个端一个独立工程，各自 `package.json`，互不引用源码。

```
template/
├── web/
│   ├── system-manage-ui/     # 管理后台（总部 + 门店，含 POS / KDS 路由）
│   └── merchant-manage-ui/   # 商户端（加盟商 / 门店老板）
└── mobile/
    ├── customer-uniapp/      # 顾客端（扫码点餐 / 外卖 / 自提 / 会员）
    └── rider-uniapp/         # 骑手端（抢单 / 取货 / 送达）
```

| 目录 | 使用者 | 技术栈 | 端口 | 接口前缀 | 状态 |
| --- | --- | --- | --- | --- | --- |
| `web/system-manage-ui` | 总部运营、财务、店长、收银、后厨 | Vue 3.5 + Vite 7 + TypeScript + Element Plus + Tailwind 4 + Pinia | 22222 | `/api/v1` | ✅ 可用 |
| `web/merchant-manage-ui` | 加盟商、门店老板 | 同上（建议复用同一套脚手架） | 22223 | `/api/v1` | ⬜ 占位 |
| `mobile/customer-uniapp` | 顾客 | uni-app（Vue 3 + TypeScript） | H5 `22224` | `/api/v1/client` | ⬜ 占位 |
| `mobile/rider-uniapp` | 骑手 | uni-app（Vue 3 + TypeScript） | H5 `22225` | `/api/v1` | ⬜ 占位 |

## 约定

- **管理后台**是唯一可用的前端；其余三个目录只有 `.gitkeep`，对应业务模块开工前不要写入代码。
- 端口从 `22222` 起按上表顺延；开发环境统一把 `/api` 代理到后端 `11111`。
- `web/*` 走后台令牌与权限点；`mobile/customer-uniapp` 走会员令牌（`/api/v1/client`）。
- **骑手端走后台账号体系**（门店员工角色 = 骑手），接口前缀与管理后台一致，不用 `/client`。
- 新增端一律按 `web/xxx` 或 `mobile/xxx` 平铺，不嵌套更深层级。
- POS 与 KDS 一期作为管理后台内的独立布局路由，不单独建工程；需要独立端时再拆到 `web/`。
