/**
 * 解析任意时间为“北京时区视角”的 Date 对象。
 *
 * 说明：
 * - 如果传入的是带时区的 ISO 字符串（如 ...Z 或 +00:00/+08:00），先解析为 UTC 毫秒，再加上 8 小时，
 *   随后使用 getUTC* 系列方法格式化即可得到稳定的北京时间（不受客户端所在时区影响）。
 * - 如果传入的是不带时区的字符串（如 2025-09-07 18:00 或 2025-09-07T18:00:00），
 *   视为“北京时间”，按组件手动解析并构造一个以该时间作为 UTC 的 Date（即 Date.UTC(year, month-1, day, hour, ...)
 *   这样配合 getUTC* 系列方法显示时即为北京时间的原始数值）。
 * - 如果传入的是 Date 实例，表示一个“瞬时时刻”，直接加 8 小时后配合 getUTC* 显示北京时间。
 */
export const convertToBeijingTime = (input: Date | string): Date => {
  // 字符串处理
  if (typeof input === 'string') {
    const s = input.trim()

    // 含明确时区（Z 或 ±hh:mm）：直接解析为“瞬时”，后续格式化使用 UTC 视角即可
    if (/[zZ]|[+\-]\d{2}:?\d{2}$/.test(s)) {
      const parsed = Date.parse(s)
      if (!Number.isNaN(parsed)) {
        return new Date(parsed)
      }
    }

    // 不带时区，视为“北京时间墙上时间”
    const re = /^(\d{4})-(\d{2})-(\d{2})(?:[ T](\d{2}):(\d{2})(?::(\d{2}))?)?$/
    const m = s.match(re)
    if (m) {
      const year = Number(m[1])
      const month = Number(m[2])
      const day = Number(m[3])
      const hour = Number(m[4] ?? '0')
      const minute = Number(m[5] ?? '0')
      const second = Number(m[6] ?? '0')
      const ms = Date.UTC(year, month - 1, day, hour, minute, second)
      return new Date(ms)
    }

    // 回退：原生解析
    const fallbackMs = Date.parse(s)
    if (!Number.isNaN(fallbackMs)) {
      return new Date(fallbackMs + 8 * 60 * 60 * 1000)
    }

    return new Date()
  }

  // Date 实例：瞬时 -> 北京时间视角
  return new Date(input.getTime() + 8 * 60 * 60 * 1000)
}

/**
 * 格式化日期为YYYY-MM-DD（北京时间）
 */
export const formatDate = (date: Date | string): string => {
  const d = convertToBeijingTime(date)
  const year = d.getUTCFullYear()
  const month = String(d.getUTCMonth() + 1).padStart(2, '0')
  const day = String(d.getUTCDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

/**
 * 格式化日期时间为YYYY-MM-DD HH:MM（北京时间）
 */
export const formatDateTime = (date: Date | string): string => {
  const d = convertToBeijingTime(date)
  const year = d.getUTCFullYear()
  const month = String(d.getUTCMonth() + 1).padStart(2, '0')
  const day = String(d.getUTCDate()).padStart(2, '0')
  const hours = String(d.getUTCHours()).padStart(2, '0')
  const minutes = String(d.getUTCMinutes()).padStart(2, '0')
  return `${year}-${month}-${day} ${hours}:${minutes}`
}

/**
 * 格式化为相对时间（例如：3小时前，2天前等）
 */
export const formatRelativeTime = (date: Date | string): string => {
  const d = convertToBeijingTime(date)
  const now = convertToBeijingTime(new Date())
  const diffMs = now.getTime() - d.getTime()

  // 转换为秒
  const diffSec = Math.round(diffMs / 1000)

  // 小于60秒
  if (diffSec < 60) {
    return '刚刚'
  }

  // 分钟
  const diffMin = Math.round(diffSec / 60)
  if (diffMin < 60) {
    return `${diffMin}分钟前`
  }

  // 小时
  const diffHour = Math.round(diffMin / 60)
  if (diffHour < 24) {
    return `${diffHour}小时前`
  }

  // 天
  const diffDay = Math.round(diffHour / 24)
  if (diffDay < 30) {
    return `${diffDay}天前`
  }

  // 月
  const diffMonth = Math.round(diffDay / 30)
  if (diffMonth < 12) {
    return `${diffMonth}月前`
  }

  // 年
  const diffYear = Math.round(diffMonth / 12)
  return `${diffYear}年前`
}

/**
 * 计算倒计时（返回天、小时、分钟、秒）
 */
export const calculateCountdown = (targetDate: Date | string): { days: number; hours: number; minutes: number; seconds: number } => {
  const target = convertToBeijingTime(targetDate)
  const now = convertToBeijingTime(new Date())

  // 计算剩余毫秒数
  const diffMs = target.getTime() - now.getTime()

  // 如果目标时间已经过去，则返回0
  if (diffMs <= 0) {
    return { days: 0, hours: 0, minutes: 0, seconds: 0 }
  }

  // 计算剩余天数、小时、分钟和秒
  const days = Math.floor(diffMs / (1000 * 60 * 60 * 24))
  const hours = Math.floor((diffMs % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60))
  const minutes = Math.floor((diffMs % (1000 * 60 * 60)) / (1000 * 60))
  const seconds = Math.floor((diffMs % (1000 * 60)) / 1000)

  return { days, hours, minutes, seconds }
}

/**
 * 获取“北京时间”当天 00:00（使用 getUTCHours 进行归零）
 */
export const getBeijingStartOfDay = (date: Date | string): Date => {
  const d = convertToBeijingTime(date)
  const start = new Date(d.getTime())
  start.setUTCHours(0, 0, 0, 0)
  return start
}

/** 判断两时间是否同一天（北京时间） */
export const isSameBeijingDay = (a: Date | string, b: Date | string): boolean => {
  return getBeijingStartOfDay(a).getTime() === getBeijingStartOfDay(b).getTime()
}