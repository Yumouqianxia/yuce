/**
 * 显示消息提示
 * @param {string} message - 消息内容
 * @param {string} type - 消息类型: success, warning, error, info
 * @param {number} duration - 显示时长(毫秒)
 */
export function showMessage(message, type = 'info', duration = 3000) {
  // 创建消息元素
  const messageEl = document.createElement('div');
  messageEl.className = `custom-message message-${type}`;
  
  // 创建图标
  const iconEl = document.createElement('span');
  iconEl.className = 'message-icon';
  
  // 根据类型设置图标
  switch (type) {
    case 'success':
      iconEl.innerHTML = '✓';
      break;
    case 'warning':
      iconEl.innerHTML = '⚠';
      break;
    case 'error':
      iconEl.innerHTML = '✕';
      break;
    default:
      iconEl.innerHTML = 'ℹ';
  }
  
  // 创建消息文本
  const textEl = document.createElement('span');
  textEl.className = 'message-text';
  textEl.textContent = message;
  
  // 组装消息
  messageEl.appendChild(iconEl);
  messageEl.appendChild(textEl);
  
  // 添加到文档
  document.body.appendChild(messageEl);
  
  // 添加显示动画
  setTimeout(() => {
    messageEl.style.opacity = '1';
    messageEl.style.transform = 'translateY(0)';
  }, 10);
  
  // 设置自动关闭
  setTimeout(() => {
    messageEl.style.opacity = '0';
    messageEl.style.transform = 'translateY(-20px)';
    
    // 移除元素
    setTimeout(() => {
      document.body.removeChild(messageEl);
    }, 300);
  }, duration);
}

// 添加全局样式
const style = document.createElement('style');
style.textContent = `
  .custom-message {
    min-width: 240px;
    padding: 15px 20px;
    display: flex;
    align-items: center;
    border-radius: 6px;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.12);
    position: fixed;
    top: 20px;
    left: 50%;
    transform: translateX(-50%) translateY(-20px);
    z-index: 9999;
    opacity: 0;
    transition: opacity 0.3s, transform 0.3s;
    background-color: white;
  }
  
  .message-icon {
    margin-right: 10px;
    font-size: 18px;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 24px;
    height: 24px;
  }
  
  .message-success {
    border-left: 4px solid #67C23A;
  }
  
  .message-success .message-icon {
    color: #67C23A;
  }
  
  .message-warning {
    border-left: 4px solid #E6A23C;
  }
  
  .message-warning .message-icon {
    color: #E6A23C;
  }
  
  .message-error {
    border-left: 4px solid #F56C6C;
  }
  
  .message-error .message-icon {
    color: #F56C6C;
  }
  
  .message-info {
    border-left: 4px solid #909399;
  }
  
  .message-info .message-icon {
    color: #909399;
  }
`;

// 将样式添加到文档头部
document.head.appendChild(style);
