import mitt from 'mitt'

// 创建事件总线实例
const eventBus = mitt()

// 定义事件类型
export const EVENTS = {
  // 现有事件
  MATCH_DATA_UPDATED: 'match-data-updated'
  
  // WebSocket相关事件已移除
  // 实时数据更新功能已移除，改为定时刷新或手动刷新
}

export default eventBus
