<template>
  <div>
    <div class="table-container">
      <table class="data-table">
        <thead>
          <tr>
            <th>ID</th>
            <th>用户名</th>
            <th>昵称</th>
            <th>邮箱</th>
            <th>角色</th>
            <th>注册时间</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody v-if="loading">
          <tr>
            <td colspan="7" class="loading-cell">加载中...</td>
          </tr>
        </tbody>
        <tbody v-else-if="users.length === 0">
          <tr>
            <td colspan="7" class="empty-cell">暂无数据</td>
          </tr>
        </tbody>
        <tbody v-else>
          <tr v-for="user in users" :key="user.id">
            <td>{{ user.id }}</td>
            <td>{{ user.username }}</td>
            <td>{{ user.nickname }}</td>
            <td>{{ user.email }}</td>
            <td>
              <span class="role-tag" :class="user.role === 'admin' ? 'admin-role' : 'user-role'">
                {{ user.role === 'admin' ? '管理员' : '用户' }}
              </span>
            </td>
            <td>{{ formatDate(user.createdAt ?? (user as any).created_at) }}</td>
            <td class="action-cell">
              <button class="btn-reset" @click="$emit('reset-password', user)">重置密码</button>
              <button class="btn-edit" @click="$emit('edit', user)">编辑</button>
              <button
                class="btn-delete"
                @click="$emit('delete', user)"
                :disabled="user.role === 'admin' || user.username === 'root'">
                删除
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div class="pagination-container">
      <div class="pagination-box">
        <button
          class="prev-btn"
          @click="$emit('page-change', Math.max(1, currentPage - 1))"
          :disabled="currentPage <= 1">上一页</button>
        <div class="page-numbers">
          <template v-if="totalPages <= 7">
            <button
              v-for="page in totalPages"
              :key="page"
              @click="$emit('page-change', page)"
              class="page-number"
              :class="{ active: currentPage === page }">
              {{ page }}
            </button>
          </template>
          <template v-else>
            <button
              class="page-number"
              :class="{ active: currentPage === 1 }"
              @click="$emit('page-change', 1)">
              1
            </button>

            <span class="ellipsis" v-if="currentPage > 3">...</span>

            <template v-for="page in pageList" :key="page">
              <button
                v-if="page > 1 && page < totalPages"
                @click="$emit('page-change', page)"
                class="page-number"
                :class="{ active: currentPage === page }">
                {{ page }}
              </button>
            </template>

            <span class="ellipsis" v-if="currentPage < totalPages - 2">...</span>

            <button
              class="page-number"
              :class="{ active: currentPage === totalPages }"
              @click="$emit('page-change', totalPages)">
              {{ totalPages }}
            </button>
          </template>
        </div>
        <button
          class="next-btn"
          @click="$emit('page-change', Math.min(totalPages, currentPage + 1))"
          :disabled="currentPage >= totalPages">下一页</button>
        <div class="page-size-select">
          <span>每页条数: </span>
          <select :value="pageSize" @change="onSizeChange" class="page-size-dropdown">
            <option :value="10">10</option>
            <option :value="20">20</option>
            <option :value="50">50</option>
            <option :value="100">100</option>
          </select>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { User } from '@/types/user'

const props = defineProps<{
  users: User[]
  loading: boolean
  currentPage: number
  totalPages: number
  pageList: number[]
  pageSize: number
}>()

const emit = defineEmits<{
  (e: 'edit', user: User): void
  (e: 'delete', user: User): void
  (e: 'reset-password', user: User): void
  (e: 'page-change', page: number): void
  (e: 'page-size-change', size: number): void
}>()

const onSizeChange = (event: Event) => {
  const value = Number((event.target as HTMLSelectElement).value)
  emit('page-size-change', value)
}

const formatDate = (value?: string) => {
  if (!value) return ''
  const date = new Date(value)
  return isNaN(date.getTime()) ? '' : date.toLocaleString()
}
</script>

<style scoped>
.table-container {
  margin-bottom: 20px;
  overflow-x: auto;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
  border: 1px solid #ebeef5;
  text-align: center;
}

.data-table th,
.data-table td {
  padding: 12px;
  border-bottom: 1px solid #ebeef5;
}

.data-table th {
  background-color: #f5f7fa;
  font-weight: 500;
  color: #606266;
}

.data-table tbody tr:nth-child(even) {
  background-color: #fafafa;
}

.data-table tbody tr:hover {
  background-color: #f5f7fa;
}

.loading-cell,
.empty-cell {
  text-align: center;
  padding: 30px;
  color: #909399;
}

.role-tag {
  display: inline-block;
  padding: 4px 10px;
  font-size: 12px;
  border-radius: 4px;
}

.admin-role {
  background-color: #f56c6c;
  color: #fff;
}

.user-role {
  background-color: #409eff;
  color: #fff;
}

.action-cell {
  white-space: nowrap;
}

.btn-edit,
.btn-delete,
.btn-reset {
  padding: 6px 12px;
  border-radius: 4px;
  border: 1px solid;
  background-color: transparent;
  font-size: 12px;
  cursor: pointer;
  margin: 0 4px;
}

.btn-reset {
  color: #909399;
  border-color: #dcdfe6;
}

.btn-reset:hover {
  color: #606266;
  border-color: #c0c4cc;
}

.btn-edit {
  color: #409eff;
  border-color: #c6e2ff;
}

.btn-edit:hover {
  color: #fff;
  background-color: #409eff;
  border-color: #409eff;
}

.btn-delete {
  color: #f56c6c;
  border-color: #fbc4c4;
}

.btn-delete:hover {
  color: #fff;
  background-color: #f56c6c;
  border-color: #f56c6c;
}

.btn-delete:disabled {
  color: #c0c4cc;
  border-color: #e4e7ed;
  cursor: not-allowed;
  background-color: #fff;
}

.pagination-container {
  display: flex;
  justify-content: center;
  margin-top: 20px;
}

.pagination-box {
  display: flex;
  align-items: center;
  justify-content: center;
  background: #fff;
  padding: 8px 10px;
  border-radius: 4px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.page-numbers {
  display: flex;
  align-items: center;
  margin: 0 10px;
}

.page-number {
  min-width: 32px;
  height: 32px;
  margin: 0 4px;
  padding: 0 4px;
  font-size: 14px;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  background-color: #fff;
  color: #606266;
  cursor: pointer;
}

.page-number:hover {
  color: #409eff;
}

.page-number.active {
  background-color: #409eff;
  color: #fff;
  border-color: #409eff;
}

.prev-btn,
.next-btn {
  padding: 6px 12px;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  background-color: #fff;
  color: #606266;
  cursor: pointer;
}

.prev-btn:disabled,
.next-btn:disabled {
  color: #c0c4cc;
  cursor: not-allowed;
}

.page-size-select {
  display: flex;
  align-items: center;
  margin-left: 10px;
}

.page-size-dropdown {
  margin-left: 6px;
  padding: 4px 8px;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
}

.ellipsis {
  margin: 0 4px;
  color: #909399;
}
</style>

