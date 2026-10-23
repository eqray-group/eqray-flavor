<!-- 菜品新增 / 编辑弹窗：菜品基本信息、规格整表维护与加料组挂载 -->
<template>
  <ElDialog
    v-model="dialogVisible"
    :title="isEdit ? '编辑菜品' : '新增菜品'"
    width="980px"
    align-center
    :close-on-click-modal="false"
    destroy-on-close
  >
    <ElForm ref="formRef" :model="form" :rules="rules" label-width="92px">
      <div class="grid grid-cols-2 gap-x-4">
        <ElFormItem label="所属分类" prop="categoryId">
          <ElTreeSelect
            v-model="form.categoryId"
            :data="categoryOptions"
            :props="{ label: 'name', value: 'id', children: 'children' }"
            check-strictly
            :render-after-expand="false"
            placeholder="请选择分类"
            style="width: 100%"
          />
        </ElFormItem>

        <ElFormItem label="菜品名称" prop="name">
          <ElInput v-model="form.name" placeholder="如宫保鸡丁" maxlength="64" />
        </ElFormItem>

        <ElFormItem label="菜品编码" prop="code">
          <ElInput v-model="form.code" placeholder="选填，非空时全局唯一" maxlength="64" />
        </ElFormItem>

        <ElFormItem label="销售单位" prop="unit">
          <ElInput v-model="form.unit" placeholder="如份、杯、例" maxlength="16" />
        </ElFormItem>

        <ElFormItem label="标签" prop="tags">
          <ElInput v-model="form.tags" placeholder="选填，逗号分隔，如招牌,微辣" maxlength="255" />
        </ElFormItem>

        <ElFormItem label="过敏原" prop="allergens">
          <ElInput
            v-model="form.allergens"
            placeholder="选填，逗号分隔，如花生,海鲜"
            maxlength="255"
          />
        </ElFormItem>

        <ElFormItem label="排序" prop="sort">
          <ElInputNumber v-model="form.sort" :min="0" :max="9999" style="width: 100%" />
        </ElFormItem>

        <ElFormItem label="状态" prop="status">
          <ElRadioGroup v-model="form.status">
            <ElRadio :value="1">上架</ElRadio>
            <ElRadio :value="0">下架</ElRadio>
          </ElRadioGroup>
        </ElFormItem>
      </div>

      <ElFormItem label="菜品主图" prop="image">
        <ElUpload
          :show-file-list="false"
          :before-upload="beforeImageUpload"
          :http-request="handleImageUpload"
          accept=".jpg,.jpeg,.png,.gif,.webp,.bmp"
        >
          <div
            v-loading="uploading"
            class="group relative flex h-20 w-20 items-center justify-center overflow-hidden rounded-lg border border-dashed border-[var(--el-border-color)] cursor-pointer"
          >
            <img
              v-if="form.image"
              class="h-full w-full object-cover"
              :src="form.image"
              alt="菜品主图"
            />
            <ArtSvgIcon v-else icon="ri:image-add-line" class="text-xl text-g-600" />
            <div
              class="absolute inset-0 flex items-center justify-center text-xs text-white transition-opacity bg-black/40 opacity-0 group-hover:opacity-100"
            >
              更换
            </div>
          </div>
        </ElUpload>
        <div class="ml-3 flex flex-col gap-1 text-xs text-g-600">
          <span>选填，建议 1:1 方图，不超过 2MB</span>
          <ElButton v-if="form.image" link type="danger" size="small" @click="form.image = ''">
            移除主图
          </ElButton>
        </div>
      </ElFormItem>

      <ElFormItem label="菜品介绍" prop="description">
        <ElInput
          v-model="form.description"
          type="textarea"
          :rows="2"
          maxlength="500"
          show-word-limit
          placeholder="选填，用于 C 端菜品详情展示"
        />
      </ElFormItem>

      <ElFormItem label="加料组" prop="addonGroupIds">
        <ElSelect
          v-model="form.addonGroupIds"
          multiple
          filterable
          clearable
          class="w-full"
          placeholder="选填，可挂载多个加料组"
          no-data-text="暂无可选加料组"
        >
          <ElOption
            v-for="option in addonOptions"
            :key="option.value"
            :label="option.label"
            :value="option.value"
          />
        </ElSelect>
      </ElFormItem>

      <ElFormItem label="备注" prop="remark">
        <ElInput
          v-model="form.remark"
          type="textarea"
          :rows="2"
          maxlength="255"
          show-word-limit
          placeholder="选填"
        />
      </ElFormItem>

      <!-- 规格整表提交：编辑时未提交的存量规格会被后端删除 -->
      <ElFormItem label="菜品规格">
        <div class="w-full">
          <div class="mb-2 flex items-center justify-between">
            <span class="text-xs text-g-600">
              价格按元填写，会员价留空表示未设置；编辑时删除的规格会同步删除门店配置。
            </span>
            <ElButton type="primary" plain size="small" @click="addSku">添加规格</ElButton>
          </div>

          <div class="w-full overflow-x-auto">
            <table class="sku-editor">
              <thead>
                <tr>
                  <th style="min-width: 150px">规格名称</th>
                  <th style="min-width: 130px">规格编码</th>
                  <th style="min-width: 120px">标准价（元）</th>
                  <th style="min-width: 120px">会员价（元）</th>
                  <th style="min-width: 110px">成本价（元）</th>
                  <th style="min-width: 90px">排序</th>
                  <th style="min-width: 100px">状态</th>
                  <th style="width: 70px">操作</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(sku, index) in form.skus" :key="sku.key">
                  <td>
                    <ElInput v-model="sku.name" placeholder="如大份" maxlength="64" />
                  </td>
                  <td>
                    <ElInput v-model="sku.code" placeholder="选填" maxlength="64" />
                  </td>
                  <td>
                    <ElInputNumber
                      v-model="sku.standardPrice"
                      :min="0"
                      :precision="2"
                      :step="1"
                      :controls="false"
                      class="w-full"
                      placeholder="必填"
                    />
                  </td>
                  <td>
                    <ElInputNumber
                      v-model="sku.memberPrice"
                      :min="0"
                      :precision="2"
                      :step="1"
                      :controls="false"
                      class="w-full"
                      placeholder="选填"
                    />
                  </td>
                  <td>
                    <ElInputNumber
                      v-model="sku.costPrice"
                      :min="0"
                      :precision="2"
                      :step="1"
                      :controls="false"
                      class="w-full"
                      placeholder="选填"
                    />
                  </td>
                  <td>
                    <ElInputNumber
                      v-model="sku.sort"
                      :min="0"
                      :max="9999"
                      :controls="false"
                      class="w-full"
                    />
                  </td>
                  <td>
                    <ElSelect v-model="sku.status" class="w-full">
                      <ElOption label="上架" :value="1" />
                      <ElOption label="下架" :value="0" />
                    </ElSelect>
                  </td>
                  <td>
                    <ElButton
                      link
                      type="danger"
                      :disabled="form.skus.length <= 1"
                      @click="removeSku(index)"
                    >
                      删除
                    </ElButton>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </ElFormItem>
    </ElForm>

    <template #footer>
      <ElButton @click="dialogVisible = false">取消</ElButton>
      <ElButton
        type="primary"
        :loading="submitting"
        :disabled="detailLoading"
        @click="handleSubmit"
      >
        提交
      </ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import { ElMessage } from 'element-plus'
  import type { FormInstance, FormRules, UploadRequestOptions } from 'element-plus'
  import { fetchUploadFile } from '@/api/infra/file'
  import { fetchCreateSpu, fetchGetSpuDetail, fetchUpdateSpu } from '@/api/dish/spu'

  defineOptions({ name: 'DishSpuDialog' })

  type SpuItem = Api.Dish.SpuItem
  type DialogType = 'add' | 'edit'
  type SelectOption = { label: string; value: string }

  /** 规格编辑行：价格用元，提交时转换为分 */
  interface SkuRow {
    key: string
    id?: string
    name: string
    code: string
    standardPrice: number | null
    memberPrice: number | null
    costPrice: number | null
    status: number
    sort: number
  }

  interface Props {
    visible: boolean
    type: DialogType
    spuData?: SpuItem | null
    categoryOptions?: Api.Dish.CategoryNode[]
    addonOptions?: SelectOption[]
  }

  const props = withDefaults(defineProps<Props>(), {
    spuData: null,
    categoryOptions: () => [],
    addonOptions: () => []
  })

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void
    (e: 'success'): void
  }>()

  const dialogVisible = computed({
    get: () => props.visible,
    set: (value: boolean) => emit('update:visible', value)
  })

  const isEdit = computed(() => props.type === 'edit')

  const formRef = ref<FormInstance>()
  const submitting = ref(false)
  const uploading = ref(false)
  const detailLoading = ref(false)

  const form = reactive({
    categoryId: '',
    name: '',
    code: '',
    unit: '份',
    description: '',
    image: '',
    tags: '',
    allergens: '',
    status: 1,
    sort: 0,
    remark: '',
    addonGroupIds: [] as string[],
    skus: [] as SkuRow[]
  })

  const rules: FormRules = {
    categoryId: [{ required: true, message: '请选择所属分类', trigger: 'change' }],
    name: [{ required: true, message: '请输入菜品名称', trigger: 'blur' }]
  }

  // 生成本地行标识：新增行用递增序号，编辑行沿用后端 ID
  let rowSeed = 0
  const createSkuRow = (): SkuRow => {
    rowSeed += 1
    return {
      key: `new-${rowSeed}`,
      name: '',
      code: '',
      standardPrice: null,
      memberPrice: null,
      costPrice: null,
      status: 1,
      sort: 0
    }
  }

  // 追加一条空规格
  const addSku = (): void => {
    form.skus.push(createSkuRow())
  }

  // 删除指定规格，至少保留一条
  const removeSku = (index: number): void => {
    if (form.skus.length <= 1) return
    form.skus.splice(index, 1)
  }

  // 元转分：按四舍五入取整，空值按 0 处理
  const toCent = (yuan: number | null): number => {
    if (yuan === null || Number.isNaN(yuan)) return 0
    return Math.round(yuan * 100)
  }

  // 分转元：用于表单回填
  const toYuan = (cent: number): number => Number((cent / 100).toFixed(2))

  // 校验图标类型与大小
  const beforeImageUpload = (file: File): boolean => {
    const allowed = ['image/jpeg', 'image/png', 'image/gif', 'image/webp', 'image/bmp']
    if (!allowed.includes(file.type)) {
      ElMessage.error('仅支持 jpg / png / gif / webp / bmp 图片')
      return false
    }
    if (file.size / 1024 / 1024 > 2) {
      ElMessage.error('图片不能超过 2MB')
      return false
    }
    return true
  }

  // 上传菜品主图：成功后写入表单地址
  const handleImageUpload = async (options: UploadRequestOptions): Promise<void> => {
    uploading.value = true
    try {
      const uploaded = await fetchUploadFile(options.file)
      form.image = uploaded.url
    } catch {
      // 错误提示已由请求层统一弹出
    } finally {
      uploading.value = false
    }
  }

  // 重置表单为新增默认值
  const resetForm = (): void => {
    Object.assign(form, {
      categoryId: '',
      name: '',
      code: '',
      unit: '份',
      description: '',
      image: '',
      tags: '',
      allergens: '',
      status: 1,
      sort: 0,
      remark: '',
      addonGroupIds: []
    })
    form.skus = [createSkuRow()]
  }

  // 拉取菜品详情并回填规格与加料组，编辑时列表行不含规格数据
  const loadDetail = async (id: string): Promise<void> => {
    detailLoading.value = true
    try {
      const detail = await fetchGetSpuDetail(id)
      Object.assign(form, {
        categoryId: detail.categoryId,
        name: detail.name,
        code: detail.code,
        unit: detail.unit,
        description: detail.description,
        image: detail.image,
        tags: detail.tags,
        allergens: detail.allergens,
        status: detail.status,
        sort: detail.sort,
        remark: detail.remark,
        addonGroupIds: detail.addonGroupIds || []
      })
      form.skus = (detail.skus || []).map((sku) => ({
        key: sku.id,
        id: sku.id,
        name: sku.name,
        code: sku.code,
        standardPrice: toYuan(sku.standardPrice),
        memberPrice: sku.memberPrice > 0 ? toYuan(sku.memberPrice) : null,
        costPrice: sku.costPrice > 0 ? toYuan(sku.costPrice) : null,
        status: sku.status,
        sort: sku.sort
      }))
      if (form.skus.length === 0) form.skus = [createSkuRow()]
    } catch {
      // 错误提示已由请求层统一弹出
    } finally {
      detailLoading.value = false
    }
  }

  // 每次打开弹窗初始化：编辑拉详情，新增直接重置
  watch(
    () => props.visible,
    async (visible) => {
      if (!visible) return

      formRef.value?.clearValidate()
      if (isEdit.value && props.spuData) {
        resetForm()
        Object.assign(form, {
          categoryId: props.spuData.categoryId,
          name: props.spuData.name,
          code: props.spuData.code,
          unit: props.spuData.unit || '份',
          status: props.spuData.status,
          sort: props.spuData.sort,
          remark: props.spuData.remark,
          image: props.spuData.image,
          tags: props.spuData.tags,
          allergens: props.spuData.allergens
        })
        await loadDetail(props.spuData.id)
      } else {
        resetForm()
      }

      nextTick(() => formRef.value?.clearValidate())
    },
    { immediate: true }
  )

  // 逐行校验规格：名称必填、标准价必须大于 0
  const validateSkus = (): string => {
    if (form.skus.length === 0) return '菜品至少需要一个规格'

    for (let index = 0; index < form.skus.length; index += 1) {
      const sku = form.skus[index]
      const label = `第 ${index + 1} 个规格`
      if (!sku.name.trim()) return `${label}的名称不能为空`
      if (sku.standardPrice === null || sku.standardPrice <= 0) {
        return `${label}的标准价必须大于 0`
      }
      if (sku.memberPrice !== null && sku.memberPrice > 0 && sku.memberPrice > sku.standardPrice) {
        return `${label}的会员价不能高于标准价`
      }
    }

    return ''
  }

  // 提交菜品：规格与加料组整表提交，价格由元转换为分
  const handleSubmit = async (): Promise<void> => {
    if (!formRef.value) return

    const valid = await formRef.value.validate().catch(() => false)
    if (!valid) return

    const skuError = validateSkus()
    if (skuError) {
      ElMessage.error(skuError)
      return
    }

    const payload: Api.Dish.SpuSaveParams = {
      categoryId: form.categoryId,
      name: form.name.trim(),
      code: form.code,
      unit: form.unit,
      description: form.description,
      image: form.image,
      tags: form.tags,
      allergens: form.allergens,
      status: form.status,
      sort: form.sort,
      remark: form.remark,
      skus: form.skus.map((sku) => ({
        id: sku.id,
        name: sku.name.trim(),
        code: sku.code,
        standardPrice: toCent(sku.standardPrice),
        memberPrice: toCent(sku.memberPrice),
        costPrice: toCent(sku.costPrice),
        status: sku.status,
        sort: sku.sort
      })),
      addonGroupIds: form.addonGroupIds
    }

    submitting.value = true
    try {
      if (isEdit.value && props.spuData) {
        await fetchUpdateSpu(props.spuData.id, payload)
        ElMessage.success('保存成功')
      } else {
        await fetchCreateSpu(payload)
        ElMessage.success('新增成功')
      }

      dialogVisible.value = false
      emit('success')
    } catch {
      // 错误提示已由请求层统一弹出
    } finally {
      submitting.value = false
    }
  }
</script>

<style lang="scss" scoped>
  .sku-editor {
    width: 100%;
    border-collapse: collapse;

    th,
    td {
      padding: 6px 4px;
      text-align: left;
      vertical-align: middle;
    }

    th {
      font-size: 12px;
      font-weight: 500;
      color: var(--el-text-color-secondary);
      border-bottom: 1px solid var(--el-border-color-lighter);
    }
  }
</style>
