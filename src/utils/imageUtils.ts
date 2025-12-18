/**
 * 处理头像URL，确保返回完整的URL路径
 * @param avatarPath 头像路径
 * @returns 完整的头像URL
 */
export const getFullAvatarUrl = (avatarPath: string | null | undefined): string => {
  if (!avatarPath) return '';

  // 如果已经是完整URL，直接返回
  if (avatarPath.startsWith('http')) {
    return avatarPath;
  }

  // 尝试不同的路径组合
  const baseUrl = window.location.origin;

  // 如果路径已经以斜杠开头，则移除它
  const cleanPath = avatarPath.startsWith('/') ? avatarPath.substring(1) : avatarPath;

  // 尝试直接访问
  const directUrl = `${baseUrl}/${cleanPath}`;

  // 尝试添加api前缀
  const apiUrl = `${baseUrl}/api/${cleanPath}`;

  // 尝试使用相对路径
  const relativeUrl = `/${cleanPath}`;

  // 尝试使用静态资源路径
  const staticUrl = `${baseUrl}/static/${cleanPath}`;

  // 尝试使用公共资源路径
  const publicUrl = `${baseUrl}/public/${cleanPath}`;

  // 输出调试信息
  console.log('头像路径处理:', {
    original: avatarPath,
    cleanPath,
    directUrl,
    apiUrl,
    relativeUrl,
    staticUrl,
    publicUrl
  });

  // 尝试使用直接访问上传目录
  const uploadsUrl = `${baseUrl}/uploads/${cleanPath}`;

  // 输出所有可能的URL
  console.log('所有可能的URL:', {
    directUrl,
    apiUrl,
    relativeUrl,
    staticUrl,
    publicUrl,
    uploadsUrl,
    originalPath: avatarPath,
    cleanPath
  });

  // 如果是头像文件，统一返回 /api/uploads/avatar/{filename}
  if (cleanPath.includes('avatars/')) {
    const filename = cleanPath.split('/').pop();
    return filename ? `/api/uploads/avatar/${filename}` : '';
  }

  // 默认返回上传目录路径（加 /api 前缀）
  return cleanPath ? `/api/${cleanPath}` : '';
};
