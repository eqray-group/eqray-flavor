<!-- 菜品分类新增 / 编辑弹窗：维护上级分类、名称、图标、排序与启停状态 -->
<template>
  <ElDialog
    v-model="dialogVisible"
    :title="isEdit ? '编辑菜品分类' : '新增菜品分类'"
    width="560px"
    align-center
    :close-on-click-modal="false"
  >
    <ElForm ref="formRef" :model="form" :rules="rules" label-width="92px">
      <ElFormItem label="上级分类" prop="parentId">
        <ElTreeSelect
          v-model="form.parentId"
          :data="parentOptions"
          :props="{ label: 'name', value: 'id', children: 'children' }"
          check-strictly
          clearable
          :render-after-expand="false"
          placeholder="不选表示根分类"
          style="width: 100%"
          @clear="form.parentId = '0'"
        />
      </ElFormItem>

      <ElFormItem label="分类名称" prop="name">
        <ElInput v-model="form.name" placeholder="如热菜、凉菜、饮品" maxlength="64" />
      </ElFormItem>

      <ElFormItem label="分类图标" prop="image">
        <ElUpload
          :show-file-list="false"
          :before-upload="beforeIconUpload"
          :http-request="handleIconUpload"
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
              alt="分类图标"
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
          <span>选填，支持 jpg / png / gif / webp / bmp，不超过 2MB</span>
          <ElButton v-if="form.image" link type="danger" size="small" @click="form.image = ''">
            移除图标
          </ElButton>
        </div>
      </ElFormItem>

      <ElFormItem label="排序" prop="sort">
        <ElInputNumber v-model="form.sort" :min="0" :max="9999" style="width: 100%" />
      </ElFormItem>

      <ElFormItem label="状态" prop="status">
        <ElRadioGroup v-model="form.status">
          <ElRadio :value="1">启用</ElRadio>
          <ElRadio :value="0">禁用</ElRadio>
        </ElRadioGroup>
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

      <div class="-mt-2 pl-23 text-xs text-g-600">
        分类名称在同一上级下唯一；上级分类不能是自身或自身的下级。
      </div>
    </ElForm>

    <template #footer>
      <ElButton @click="dialogVisible = false">取消</ElButton>
      <ElButton type="primary" :loading="submitting" @click="handleSubmit">提交</ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import { ElMessage } from 'element-plus'
  import type { FormInstance, FormRules, UploadRequestOptions } from 'element-plus'
  import { fetchUploadFile } from '@/api/infra/file'
  import { fetchCreateCategory, fetchUpdateCategory } from '@/api/dish/category'

  defineOptions({ name: 'DishCategoryDialog' })

  type CategoryNode = Api.Dish.CategoryNode
  type DialogType = 'add' | 'edit'

  interface Props {
    visible: boolean
    type: DialogType
    categoryData?: CategoryNode | null
    /** 新增子分类时预置的上级分类 ID，'0' 表示根分类 */
    parentId?: string
    /** 完整分类树，用于上级分类下拉与自身子树排除 */
    tree?: CategoryNode[]
  }

  const props = withDefaults(defineProps<Props>(), {
    categoryData: null,
    parentId: '0',
    tree: () => []
  })

  const dialogVisible = computed({
    get: () => props.visible,
    set: (value: boolean) => emit('update:visible', value)
  })

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void
    (e: 'success'): void
  }>()

  const isEdit = computed(() => props.type === 'edit')

  const formRef = ref<FormInstance>()
  const submitting = ref(false)
  const uploading = ref(false)

  const form = reactive({
    parentId: '0',
    name: '',
    image: '',
    sort: 0,
    status: 1,
    remark: ''
  })

  const rules: FormRules = {
    name: [{ required: true, message: '请输入分类名称', trigger: 'blur' }]
  }

  // 递归剔除指定节点及其子孙，避免把分类的上级选成自身或下级
  const excludeSubtree = (nodes: CategoryNode[], excludeId: string): CategoryNode[] => {
    return nodes.reduce<CategoryNode[]>((acc, node) => {
      if (node.id === excludeId) return acc
      acc.push({ ...node, children: excludeSubtree(node.children || [], excludeId) })
      return acc
    }, [])
  }

  // 上级分类选项：编辑时排除自身子树，其余场景直接用完整树
  const parentOptions = computed<CategoryNode[]>(() => {
    const tree = props.tree || []
    if (isEdit.value && props.categoryData) {
      return excludeSubtree(tree, props.categoryData.id)
    }
    return tree
  })

  // 校验图标类型与大小
  const beforeIconUpload = (file: File): boolean => {
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

  // 上传分类图标：成功后把返回地址写入表单，由提交时统一保存
  const handleIconUpload = async (options: UploadRequestOptions): Promise<void> => {
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

  // 每次打开弹窗回填数据：编辑用行数据，新增用预置上级
  watch(
    () => props.visible,
    (visible) => {
      if (!visible) return

      const data = props.categoryData
      Object.assign(form, {
        parentId: isEdit.value ? (data?.parentId ?? '0') : (props.parentId ?? '0'),
        name: data?.name ?? '',
        image: data?.image ?? '',
        sort: data?.sort ?? 0,
        status: data?.status ?? 1,
        remark: data?.remark ?? ''
      })

      nextTick(() => formRef.value?.clearValidate())
    },
    { immediate: true }
  )

  // 提交分类：上级为空时按根分类处理，成功后关闭弹窗并通知父页面刷新
  const handleSubmit = async (): Promise<void> => {
    if (!formRef.value) return

    const valid = await formRef.value.validate().catch(() => false)
    if (!valid) return

    const payload: Api.Dish.CategorySaveParams = {
      parentId: form.parentId || '0',
      name: form.name.trim(),
      image: form.image,
      sort: form.sort,
      status: form.status,
      remark: form.remark
    }

    submitting.value = true
    try {
      if (isEdit.value && props.categoryData) {
        await fetchUpdateCategory(props.categoryData.id, payload)
        ElMessage.success('保存成功')
      } else {
        await fetchCreateCategory(payload)
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
