/**
 * URL工具函数
 * 用于处理各种URL，避免硬编码
 */

// 获取API服务器地址
export const getApiServer = (): string => {
  return import.meta.env.VITE_API_SERVER || 'http://localhost:3000';
};

// 获取API基础URL
export const getApiBaseUrl = (): string => {
  return import.meta.env.VITE_API_BASE_URL || '/api';
};

// 获取前端URL
export const getFrontendUrl = (): string => {
  return import.meta.env.VITE_FRONTEND_URL || 'http://localhost:5173';
};

// 获取头像路径
export const getAvatarPath = (): string => {
  return '/api/uploads/avatar';
};

/**
 * 构建完整的头像URL
 * @param avatarPath 头像相对路径
 * @returns 完整的头像URL
 */
export const getFullAvatarUrl = (avatarPath: string): string => {
  if (!avatarPath) return '';

  // 如果已经是完整URL，直接返回
  if (avatarPath.startsWith('http')) {
    return avatarPath;
  }

  // 无论传入什么格式，提取文件名并拼 /api/uploads/avatar/{filename}
  const filename = avatarPath.split('/').pop() || '';
  return filename ? `/api/uploads/avatar/${filename}` : '';
};

/**
 * 为URL添加时间戳，避免缓存
 * @param url URL
 * @returns 带时间戳的URL
 */
export const addTimestamp = (url: string): string => {
  if (!url) return '';

  // 移除现有的时间戳参数
  let cleanUrl = url;
  if (url.includes('?')) {
    const [baseUrl, query] = url.split('?');
    const params = new URLSearchParams(query);
    params.delete('t');

    cleanUrl = params.toString()
      ? `${baseUrl}?${params.toString()}`
      : baseUrl;
  }

  // 添加新的时间戳
  const separator = cleanUrl.includes('?') ? '&' : '?';
  return `${cleanUrl}${separator}t=${Date.now()}`;
};

/**
 * 检查URL是否指向前端服务器
 * @param url URL
 * @returns 是否是前端URL
 */
export const isFrontendUrl = (url: string): boolean => {
  if (!url) return false;
  return url.includes(getFrontendUrl()) || url.includes('localhost:5173');
};

/**
 * 将前端URL转换为后端URL
 * @param url 前端URL
 * @returns 后端URL
 */
export const convertToBackendUrl = (url: string): string => {
  if (!url) return '';

  // 如果不是前端URL，直接返回
  if (!isFrontendUrl(url)) return url;

  // 提取文件名
  const parts = url.split('/');
  const filenameWithParams = parts[parts.length - 1];
  const filename = filenameWithParams.split('?')[0]; // 去除查询参数

  // 使用新的API端点
  return `/api/uploads/avatar/${filename}`;
};
