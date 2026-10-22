<!-- 高德地图选点：点击地图取经纬度，AK 从参数配置读取，不写在前端环境变量里 -->
<template>
  <div class="art-map-picker flex items-center gap-2">
    <ElInput :model-value="displayValue" readonly :placeholder="placeholder" :disabled="disabled">
      <template #append>
        <ElButton :disabled="disabled" @click="openDialog">地图选点</ElButton>
      </template>
    </ElInput>

    <ElDialog
      v-model="visible"
      title="地图选点"
      width="760px"
      align-center
      append-to-body
      :close-on-click-modal="false"
      @closed="destroyMap"
    >
      <ElAlert
        v-if="keyMissing"
        type="warning"
        :closable="false"
        show-icon
        title="未配置高德地图 Key"
        description="请到「基础设施 → 参数配置」填写 amap_js_key（Web 端 JS API Key），如开启了安全密钥再填写 amap_security_code。"
      />

      <template v-else>
        <div class="mb-3 flex items-center gap-2">
          <ElInput
            v-model="keyword"
            placeholder="输入地址或地点名称搜索"
            clearable
            @keyup.enter="searchPlace"
          />
          <ElButton :loading="searching" @click="searchPlace">搜索</ElButton>
        </div>

        <div class="map-wrapper">
          <div class="map-container" ref="mapRef"></div>
          <div v-if="mapLoading" class="map-mask">地图加载中…</div>
        </div>

        <div class="mt-3 flex items-center justify-between text-sm">
          <span class="text-g-700">
            当前坐标：{{ draftLongitude.toFixed(6) }}, {{ draftLatitude.toFixed(6) }}
          </span>
          <span class="text-xs text-g-600">点击地图可重新选点</span>
        </div>
        <div v-if="draftAddress" class="mt-1 text-xs text-g-600">地址：{{ draftAddress }}</div>
      </template>

      <template #footer>
        <ElButton @click="visible = false">取消</ElButton>
        <ElButton type="primary" :disabled="keyMissing || mapLoading" @click="confirm">
          确定
        </ElButton>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { ElAlert, ElButton, ElDialog, ElInput, ElMessage } from 'element-plus'
  import { fetchGetConfigValue } from '@/api/infra/config'

  defineOptions({ name: 'ArtMapPicker' })

  interface Props {
    /** 经度，0 表示未设置 */
    longitude?: number
    /** 纬度，0 表示未设置 */
    latitude?: number
    placeholder?: string
    disabled?: boolean
  }

  const props = withDefaults(defineProps<Props>(), {
    longitude: 0,
    latitude: 0,
    placeholder: '点击右侧按钮在地图上选点',
    disabled: false
  })

  const emit = defineEmits<{
    (e: 'update:longitude', value: number): void
    (e: 'update:latitude', value: number): void
    /** 选点完成时回传坐标与解析出的地址，便于上层回填详细地址 */
    (e: 'pick', payload: { longitude: number; latitude: number; address: string }): void
  }>()

  // 默认中心点：杭州，未选点时用于初始化地图视野
  const DEFAULT_CENTER: [number, number] = [120.15507, 30.274085]

  const visible = ref(false)
  const keyMissing = ref(false)
  const mapLoading = ref(false)
  const searching = ref(false)
  const keyword = ref('')
  const mapRef = ref<HTMLElement>()
  const draftLongitude = ref(0)
  const draftLatitude = ref(0)
  const draftAddress = ref('')

  let map: any = null
  let marker: any = null
  let geocoder: any = null
  let placeSearch: any = null

  // 已选坐标的回显：两个都为 0 视为未设置
  const displayValue = computed(() =>
    props.longitude || props.latitude ? `${props.longitude}, ${props.latitude}` : ''
  )

  // 参数与地图脚本都是全局资源，模块级缓存，避免每次开弹窗都重复请求
  let configPromise: Promise<{ key: string; securityCode: string }> | null = null
  const loadConfig = (): Promise<{ key: string; securityCode: string }> => {
    if (!configPromise) {
      configPromise = Promise.all([
        fetchGetConfigValue('amap_js_key'),
        fetchGetConfigValue('amap_security_code')
      ]).then(([key, security]) => ({ key: key.value, securityCode: security.value }))
    }

    return configPromise
  }

  const AMAP_SCRIPT_ID = 'amap-jsapi-script'
  let amapPromise: Promise<any> | null = null

  // 手动注入高德 JS API：避免为一个脚本引入额外依赖，AK 由参数配置下发
  const loadAMap = (key: string, securityCode: string): Promise<any> => {
    if ((window as any).AMap) return Promise.resolve((window as any).AMap)
    if (amapPromise) return amapPromise

    // JS API 2.0 开启安全密钥时必须在脚本加载前注入
    if (securityCode) {
      ;(window as any)._AMapSecurityConfig = { securityJsCode: securityCode }
    }

    amapPromise = new Promise((resolve, reject) => {
      const plugins = ['AMap.Geocoder', 'AMap.PlaceSearch', 'AMap.Scale', 'AMap.ToolBar'].join(',')
      const src = `https://webapi.amap.com/maps?v=2.0&key=${encodeURIComponent(key)}&plugin=${plugins}`

      const script = document.createElement('script')
      script.id = AMAP_SCRIPT_ID
      script.src = src
      script.async = true
      script.onload = () => resolve((window as any).AMap)
      script.onerror = () => {
        amapPromise = null
        script.remove()
        reject(new Error('AMap script load failed'))
      }
      document.head.appendChild(script)
    })

    return amapPromise
  }

  // 打开弹窗：先取 AK，再初始化地图
  const openDialog = async (): Promise<void> => {
    draftLongitude.value = props.longitude || 0
    draftLatitude.value = props.latitude || 0
    draftAddress.value = ''
    keyword.value = ''
    keyMissing.value = false
    visible.value = true

    mapLoading.value = true
    try {
      const config = await loadConfig()
      if (!config.key) {
        keyMissing.value = true
        return
      }

      const AMap = await loadAMap(config.key, config.securityCode)
      await nextTick()
      initMap(AMap)
    } catch {
      ElMessage.error('高德地图加载失败，请检查 Key 与网络')
    } finally {
      mapLoading.value = false
    }
  }

  // 初始化地图、标记与插件；已有实例时只重置视野
  const initMap = (AMap: any): void => {
    if (!mapRef.value) return

    const center: [number, number] =
      draftLongitude.value || draftLatitude.value
        ? [draftLongitude.value, draftLatitude.value]
        : DEFAULT_CENTER

    if (!map) {
      map = new AMap.Map(mapRef.value, {
        zoom: 16,
        center,
        resizeEnable: true
      })
      map.addControl(new AMap.Scale())
      map.addControl(new AMap.ToolBar({ position: 'RB' }))

      marker = new AMap.Marker({ position: center, draggable: true })
      map.add(marker)
      marker.on('dragend', (event: any) => applyPoint(event.lnglat, false))
      map.on('click', (event: any) => applyPoint(event.lnglat, true))

      geocoder = new AMap.Geocoder()
      placeSearch = new AMap.PlaceSearch({ pageSize: 1 })
    } else {
      map.setCenter(center)
      marker.setPosition(center)
    }

    if (draftLongitude.value || draftLatitude.value) resolveAddress()
  }

  // 应用选点坐标：syncMarker 为 true 时把标记移到点击位置
  const applyPoint = (lnglat: any, syncMarker: boolean): void => {
    draftLongitude.value = Number(lnglat.getLng().toFixed(6))
    draftLatitude.value = Number(lnglat.getLat().toFixed(6))
    if (syncMarker && marker) marker.setPosition(lnglat)
    resolveAddress()
  }

  // 逆地理编码取地址，失败时留空不影响坐标提交
  const resolveAddress = (): void => {
    if (!geocoder) return

    geocoder.getAddress([draftLongitude.value, draftLatitude.value], (status: string, result: any) => {
      if (status !== 'complete' || !result?.regeocode) {
        draftAddress.value = ''
        return
      }
      const regeocode = result.regeocode
      draftAddress.value = regeocode.formattedAddress ?? ''
    })
  }

  // 关键字搜索：命中后把地图与标记移到该点
  const searchPlace = (): void => {
    if (!placeSearch || !keyword.value.trim()) return

    searching.value = true
    placeSearch.search(keyword.value.trim(), (status: string, result: any) => {
      searching.value = false
      const poi = result?.poiList?.pois?.[0]
      if (status !== 'complete' || !poi) {
        ElMessage.warning('没有找到匹配的地点')
        return
      }

      draftLongitude.value = Number(poi.location.getLng().toFixed(6))
      draftLatitude.value = Number(poi.location.getLat().toFixed(6))
      draftAddress.value = poi.address || poi.name || ''

      if (map && marker) {
        map.setCenter(poi.location)
        map.setZoom(17)
        marker.setPosition(poi.location)
      }
    })
  }

  // 确定：把草稿坐标回传父组件，地址交给父组件按需使用
  const confirm = (): void => {
    if (!draftLongitude.value && !draftLatitude.value) {
      ElMessage.warning('请先在地图上选择一个位置')
      return
    }

    emit('update:longitude', draftLongitude.value)
    emit('update:latitude', draftLatitude.value)
    emit('pick', {
      longitude: draftLongitude.value,
      latitude: draftLatitude.value,
      address: draftAddress.value
    })
    visible.value = false
  }

  // 关闭弹窗时销毁地图实例，避免多次打开造成内存泄漏
  const destroyMap = (): void => {
    if (map) {
      map.destroy()
      map = null
      marker = null
      geocoder = null
      placeSearch = null
    }
  }

  onBeforeUnmount(destroyMap)
</script>

<style scoped>
  .map-wrapper {
    position: relative;
    width: 100%;
    height: 360px;
  }

  .map-container {
    width: 100%;
    height: 100%;
    border-radius: 8px;
    overflow: hidden;
    background: var(--el-fill-color-light);
  }

  .map-mask {
    position: absolute;
    inset: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 8px;
    background: var(--el-fill-color-lighter);
    color: var(--el-text-color-secondary);
    font-size: 13px;
  }
</style>
