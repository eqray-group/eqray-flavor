<!-- 文件配置新增 / 编辑弹窗：支持本地磁盘与 S3 兼容存储（MinIO / OSS / COS 等） -->
<template>
  <ElDialog
    v-model="dialogVisible"
    :title="isEdit ? '编辑文件配置' : '新增文件配置'"
    width="620px"
    align-center
    :close-on-click-modal="false"
  >
    <ElForm ref="formRef" :model="form" :rules="rules" label-width="110px">
      <ElFormItem label="配置名称" prop="name">
        <ElInput v-model="form.name" placeholder="如：本地存储 / MinIO" maxlength="64" />
      </ElFormItem>

      <ElFormItem label="存储类型" prop="storage">
        <ElSelect v-model="form.storage" placeholder="请选择存储类型">
          <ElOption
            v-for="item in storageOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </ElSelect>
        <div class="mt-1 text-xs text-g-600"> 可同时存在多个配置，主配置只能有一个 </div>
      </ElFormItem>

      <template v-if="isS3">
        <ElFormItem label="存储桶" prop="bucket">
          <ElInput v-model="form.bucket" placeholder="Bucket 名称，如：eqray" maxlength="128" />
        </ElFormItem>

        <ElFormItem label="服务地址" prop="endpoint">
          <ElInput
            v-model="form.endpoint"
            placeholder="Endpoint，如：http://127.0.0.1:9000"
            maxlength="255"
          />
        </ElFormItem>

        <ElFormItem label="区域" prop="region">
          <ElInput v-model="form.region" placeholder="MinIO 可填 us-east-1" maxlength="64" />
        </ElFormItem>

        <ElFormItem label="访问密钥" prop="accessKey">
          <ElInput
            v-model="form.accessKey"
            placeholder="AccessKey，如：minioadmin"
            maxlength="128"
          />
        </ElFormItem>

        <ElFormItem label="私有密钥" prop="secretKey">
          <ElInput
            v-model="form.secretKey"
            type="password"
            show-password
            placeholder="SecretKey"
            maxlength="128"
          />
        </ElFormItem>

        <ElFormItem label="访问前缀" prop="url">
          <ElInput
            v-model="form.url"
            placeholder="选填，默认使用「服务地址/存储桶」拼接文件访问地址"
            maxlength="255"
          />
          <div class="mt-1 text-xs text-g-600">
            文件地址由该前缀拼接，存储桶需允许公开读或填写 CDN 域名
          </div>
        </ElFormItem>

        <ElFormItem label="路径风格" prop="usePathStyle">
          <ElSwitch v-model="form.usePathStyle" />
          <span class="ml-2 text-xs text-g-600">MinIO 需要开启，云厂商对象存储一般关闭</span>
        </ElFormItem>

        <ElFormItem label="目录前缀" prop="dir">
          <ElInput v-model="form.dir" placeholder="如：uploads" maxlength="128" />
        </ElFormItem>
      </template>

      <ElFormItem v-else label="存储目录" prop="dir">
        <ElInput
          v-model="form.dir"
          placeholder="相对 public 磁盘的目录，如 uploads"
          maxlength="128"
        />
      </ElFormItem>

      <ElFormItem label="主配置" prop="master">
        <ElSwitch v-model="form.master" :active-value="1" :inactive-value="0" />
        <span class="ml-2 text-xs text-g-600">新上传文件写入主配置对应的存储</span>
      </ElFormItem>
    </ElForm>

    <template #footer>
      <ElButton @click="dialogVisible = false">取消</ElButton>
      <ElButton type="primary" :loading="submitting" @click="handleSubmit">提交</ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import type { FormInstance, FormRules } from 'element-plus'
  import { ElDialog, ElMessage } from 'element-plus'
  import { fetchCreateFileConfig, fetchUpdateFileConfig } from '@/api/infra/file-config'

  defineOptions({ name: 'FileConfigDialog' })

  type FileConfigItem = Api.Infra.FileConfigItem
  type DialogType = 'add' | 'edit'
  type StorageType = 'local' | 's3'

  interface Props {
    visible: boolean
    type: DialogType
    configData?: FileConfigItem | null
  }

  const props = withDefaults(defineProps<Props>(), { configData: null })

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void
    (e: 'success'): void
  }>()

  const dialogVisible = computed({
    get: () => props.visible,
    set: (value: boolean) => emit('update:visible', value)
  })

  const isEdit = computed(() => props.type === 'edit')

  const storageOptions = [
    { label: '本地磁盘', value: 'local' },
    { label: 'S3 兼容存储（MinIO / OSS / COS）', value: 's3' }
  ]

  const formRef = ref<FormInstance>()
  const submitting = ref(false)

  const form = reactive({
    name: '',
    storage: 'local' as StorageType,
    dir: 'uploads',
    bucket: '',
    endpoint: '',
    region: 'us-east-1',
    accessKey: '',
    secretKey: '',
    url: '',
    usePathStyle: true,
    master: 0
  })

  const isS3 = computed(() => form.storage === 's3')

  const rules = computed<FormRules>(() => {
    const base: FormRules = {
      name: [
        { required: true, message: '请输入配置名称', trigger: 'blur' },
        { max: 64, message: '配置名称不能超过 64 个字符', trigger: 'blur' }
      ],
      storage: [{ required: true, message: '请选择存储类型', trigger: 'change' }]
    }

    if (!isS3.value) return base

    return {
      ...base,
      bucket: [{ required: true, message: '请输入存储桶名称', trigger: 'blur' }],
      endpoint: [{ required: true, message: '请输入服务地址', trigger: 'blur' }],
      accessKey: [{ required: true, message: '请输入访问密钥', trigger: 'blur' }],
      secretKey: [{ required: true, message: '请输入私有密钥', trigger: 'blur' }]
    }
  })

  // 打开弹窗时回填数据，配置 JSON 按存储类型解析
  watch(
    () => [props.visible, props.configData] as const,
    ([visible]) => {
      if (!visible) return

      const data = props.configData
      let conf: Record<string, any> = {}
      if (data?.config) {
        try {
          conf = JSON.parse(data.config) ?? {}
        } catch {
          conf = {}
        }
      }

      Object.assign(form, {
        name: data?.name ?? '',
        storage: (data?.storage === 's3' ? 's3' : 'local') as StorageType,
        dir: conf.dir || 'uploads',
        bucket: conf.bucket ?? '',
        endpoint: conf.endpoint ?? '',
        region: conf.region || 'us-east-1',
        accessKey: conf.accessKey ?? '',
        secretKey: conf.secretKey ?? '',
        url: conf.url ?? '',
        usePathStyle: conf.usePathStyle ?? true,
        master: data?.master ?? 0
      })

      nextTick(() => formRef.value?.clearValidate())
    },
    { immediate: true }
  )

  // 按存储类型组装配置 JSON，字段与后端 S3StorageConfig / LocalStorageConfig 对齐
  const buildConfig = (): string => {
    if (form.storage === 's3') {
      return JSON.stringify({
        bucket: form.bucket.trim(),
        endpoint: form.endpoint.trim(),
        region: form.region.trim() || 'us-east-1',
        accessKey: form.accessKey.trim(),
        secretKey: form.secretKey.trim(),
        url: form.url.trim(),
        usePathStyle: form.usePathStyle,
        dir: form.dir.trim() || 'uploads'
      })
    }

    return JSON.stringify({ dir: form.dir.trim() || 'uploads' })
  }

  const handleSubmit = async (): Promise<void> => {
    if (!formRef.value) return

    const valid = await formRef.value.validate().catch(() => false)
    if (!valid) return

    const payload: Api.Infra.FileConfigParams = {
      name: form.name.trim(),
      storage: form.storage,
      master: form.master,
      config: buildConfig()
    }

    submitting.value = true
    try {
      if (isEdit.value && props.configData) {
        await fetchUpdateFileConfig(props.configData.id, payload)
        ElMessage.success('保存成功')
      } else {
        await fetchCreateFileConfig(payload)
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
