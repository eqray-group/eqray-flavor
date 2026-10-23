/**
 * API 类型：菜品模块（分类、菜品、规格、加料、门店可售与门店价）
 * 金额字段一律为「分」，与后端存储保持一致；雪花 ID 一律 string。
 */

declare namespace Api {
  namespace Dish {
    /** 菜品分类树节点，children 为空数组表示叶子节点 */
    interface CategoryNode {
      id: string
      /** 上级分类 ID，'0' 表示根分类 */
      parentId: string
      name: string
      /** 分类图标地址 */
      image: string
      sort: number
      /** 启停状态：0 禁用 / 1 启用 */
      status: number
      remark: string
      /** 本分类下的直属菜品数量，不含子分类 */
      spuCount: number
      children: CategoryNode[]
      createdAt?: string
      updatedAt?: string
    }

    /** 分类树查询参数：分类树一次返回全量，关键字与状态仅用于前端过滤 */
    interface CategorySearchParams {
      keyword?: string
      /** 启停状态：0 禁用 / 1 启用，空串表示不限 */
      status?: number | ''
    }

    /** 分类新增 / 编辑参数 */
    interface CategorySaveParams {
      parentId: string
      name: string
      image?: string
      sort?: number
      status: number
      remark?: string
    }

    /** 菜品列表行：菜品本体 + 分类名 + 规格数 */
    interface SpuItem {
      id: string
      categoryId: string
      categoryName: string
      name: string
      code: string
      unit: string
      description: string
      image: string
      /** 标签，逗号分隔 */
      tags: string
      /** 过敏原，逗号分隔 */
      allergens: string
      /** 上下架状态：0 下架 / 1 上架 */
      status: number
      sort: number
      remark: string
      skuCount: number
      createdAt?: string
      updatedAt?: string
    }

    /** 菜品列表查询参数 */
    interface SpuSearchParams {
      keyword?: string
      categoryId?: string
      /** 上下架状态，空表示不限 */
      status?: number | ''
      pageNo?: number
      pageSize?: number
    }

    /** 菜品分页结果 */
    interface SpuList {
      list: SpuItem[]
      total: number
    }

    /** 菜品规格，价格单位为分 */
    interface SkuItem {
      id: string
      spuId: string
      name: string
      code: string
      standardPrice: number
      memberPrice: number
      costPrice: number
      /** 上下架状态：0 下架 / 1 上架 */
      status: number
      sort: number
      remark: string
    }

    /** 菜品详情：菜品本体 + 规格列表 + 已挂载加料组 */
    interface SpuDetail {
      id: string
      categoryId: string
      name: string
      code: string
      unit: string
      description: string
      image: string
      tags: string
      allergens: string
      status: number
      sort: number
      remark: string
      skus: SkuItem[]
      addonGroupIds: string[]
    }

    /** 菜品保存时提交的单条规格，id 为空表示新增 */
    interface SkuSaveParams {
      id?: string
      name: string
      code?: string
      standardPrice: number
      memberPrice?: number
      costPrice?: number
      status: number
      sort?: number
      remark?: string
    }

    /** 菜品新增 / 编辑参数，规格与加料组整表提交 */
    interface SpuSaveParams {
      categoryId: string
      name: string
      code?: string
      unit?: string
      description?: string
      image?: string
      tags?: string
      allergens?: string
      status: number
      sort?: number
      remark?: string
      skus: SkuSaveParams[]
      addonGroupIds: string[]
    }

    /** 加料组列表行：加料组本体 + 加料项数量 + 被菜品引用次数 */
    interface AddonGroupItem {
      id: string
      name: string
      /** 是否必选：0 可选 / 1 必选 */
      required: number
      minSelect: number
      /** 0 表示不限 */
      maxSelect: number
      sort: number
      status: number
      remark: string
      itemCount: number
      spuCount: number
      createdAt?: string
      updatedAt?: string
    }

    /** 加料组列表查询参数 */
    interface AddonSearchParams {
      keyword?: string
      pageNo?: number
      pageSize?: number
    }

    /** 加料组分页结果 */
    interface AddonGroupList {
      list: AddonGroupItem[]
      total: number
    }

    /** 加料组新增 / 编辑参数 */
    interface AddonGroupSaveParams {
      name: string
      required: number
      minSelect?: number
      maxSelect?: number
      sort?: number
      status: number
      remark?: string
    }

    /** 加料项，price 单位为分，0 表示免费 */
    interface AddonItem {
      id: string
      groupId: string
      name: string
      price: number
      sort: number
      status: number
      remark: string
    }

    /** 加料组详情：加料组本体 + 加料项列表 */
    interface AddonGroupDetail extends AddonGroupItem {
      items: AddonItem[]
    }

    /** 加料项新增 / 编辑参数 */
    interface AddonItemSaveParams {
      name: string
      price: number
      sort?: number
      status: number
      remark?: string
    }

    /** 门店维度规格配置：品牌菜品信息 + 门店可售与门店价 */
    interface StoreSkuItem {
      skuId: string
      skuName: string
      skuCode: string
      spuId: string
      spuName: string
      categoryId: string
      categoryName: string
      /** 品牌标准价（分） */
      standardPrice: number
      /** 品牌会员价（分） */
      memberPrice: number
      skuStatus: number
      spuStatus: number
      /** 该门店是否配置过此规格，未配置按不可售处理 */
      configured: boolean
      available: number
      soldOut: number
      onlineOnly: number
      sort: number
      remark: string
      /** 门店价（分），0 表示沿用品牌价 */
      storePrice: number
      /** 门店会员价（分），0 表示沿用品牌会员价 */
      storeMemberPrice: number
      /** 生效区间，空串或 null 表示不限 */
      startAt?: string | null
      endAt?: string | null
    }

    /** 门店菜品配置列表查询参数 */
    interface StoreSkuSearchParams {
      keyword?: string
      categoryId?: string
      pageNo?: number
      pageSize?: number
    }

    /** 门店菜品配置分页结果 */
    interface StoreSkuList {
      list: StoreSkuItem[]
      total: number
    }

    /** 单规格门店配置保存参数，门店由请求头决定 */
    interface StoreSkuSaveParams {
      available: number
      soldOut: number
      onlineOnly: number
      sort?: number
      remark?: string
      storePrice?: number
      storeMemberPrice?: number
      /** yyyy-MM-dd HH:mm:ss，空表示立即生效 */
      startAt?: string
      /** yyyy-MM-dd HH:mm:ss，空表示长期有效 */
      endAt?: string
    }

    /** 批量调整门店可售开关，未传的开关不改 */
    interface StoreSkuBatchParams {
      skuIds: string[]
      available?: number
      soldOut?: number
      onlineOnly?: number
    }
  }
}
