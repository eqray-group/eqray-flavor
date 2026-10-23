/**
 * 文件下载工具
 * 后端导出接口返回二进制流，这里统一把 Blob 落成浏览器下载，避免各页面重复写 a 标签逻辑。
 *
 * @module utils/download
 */

/**
 * 触发浏览器下载一个 Blob。
 * @param blob 文件内容
 * @param filename 建议文件名（含扩展名）
 */
export function saveBlob(blob: Blob, filename: string): void {
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
}

/**
 * 生成带日期后缀的导出文件名。
 * @param prefix 文件名前缀，如 pay-orders
 * @param ext 扩展名，默认 csv
 */
export function exportFilename(prefix: string, ext = 'csv'): string {
  const date = new Date().toISOString().slice(0, 10)

  return `${prefix}-${date}.${ext}`
}
