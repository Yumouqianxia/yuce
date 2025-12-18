<template>
  <div class="admin-section">
    <h2 class="page-title">比赛管理</h2>

    <div class="action-row">
      <div class="search-filters">
        <div class="search-item">
          <label for="id-search">ID</label>
          <input
            id="id-search"
            v-model="idSearch"
            type="text"
            class="input-field"
            @input="handleSearch"
          />
        </div>

        <div class="search-item">
          <label for="name-search">比赛名称</label>
          <div class="input-with-button">
            <input
              id="name-search"
              v-model="searchQuery"
              type="text"
              class="input-field"
              @input="handleSearch"
            />
            <button class="search-button" @click="handleSearch">
              搜索
            </button>
          </div>
        </div>

        <div class="search-item">
          <div class="select-container">
            <select v-model="statusFilter" class="select-field" @change="handleSearch">
              <option value="">状态</option>
              <option value="not_started">未开始</option>
              <option value="in_progress">进行中</option>
              <option value="completed">已完成</option>
              <option value="cancelled">已取消</option>
            </select>
          </div>
        </div>

        <div class="search-item">
          <div class="select-container">
            <select v-model="typeFilter" class="select-field" @change="handleSearch">
              <option value="">比赛类型</option>
              <option value="regular">常规赛</option>
              <option value="playoff">季后赛</option>
              <option value="final">总决赛</option>
            </select>
          </div>
        </div>

        <div class="search-item">
          <div class="select-container">
            <select v-model="seriesFilter" class="select-field" @change="handleSearch">
              <option value="">比赛局数</option>
              <option value="BO2">BO2</option>
              <option value="BO3">BO3</option>
              <option value="BO5">BO5</option>
              <option value="BO7">BO7</option>
              <option value="BO9">BO9</option>
            </select>
          </div>
        </div>

        <div class="search-item">
          <div class="select-container">
            <select v-model="tournamentTypeFilter" class="select-field" @change="handleSearch">
              <option value="">赛事类型</option>
              <option value="spring">KPL春季赛</option>
              <option value="summer">KPL夏季赛</option>
              <option value="annual">KPL年度总决赛</option>
              <option value="challenger">KPL挑战者杯</option>
            </select>
          </div>
        </div>

        <div class="search-item">
          <div class="select-container">
            <select v-model="tournamentStageFilter" class="select-field" @change="handleSearch">
              <option value="">赛事阶段</option>
              <option v-if="['', 'spring', 'summer'].includes(tournamentTypeFilter)" value="regular">常规赛</option>
              <option v-if="['', 'spring', 'summer'].includes(tournamentTypeFilter)" value="playoff">季后赛</option>
              <option v-if="['', 'annual', 'challenger'].includes(tournamentTypeFilter)" value="group">小组赛</option>
              <option v-if="['', 'annual', 'challenger'].includes(tournamentTypeFilter)" value="knockout">淘汰赛</option>
            </select>
          </div>
        </div>

        <button class="add-button" @click="showAddForm = true">
          添加比赛
        </button>
        <button class="add-button secondary" @click="openCreateTeam">
          管理战队
        </button>
      </div>
    </div>

    <div class="table-container">
      <table class="data-table">
        <thead>
          <tr>
            <th>ID</th>
            <th>队伍A</th>
            <th>队伍B</th>
            <th>比赛类型</th>
            <th>比赛局数</th>
            <th>比赛时间</th>
            <th>状态</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody v-if="loading">
          <tr>
            <td colspan="8" class="loading-cell">加载中...</td>
          </tr>
        </tbody>
        <tbody v-else-if="currentPageMatches.length === 0">
          <tr>
            <td colspan="8" class="empty-cell">暂无数据</td>
          </tr>
        </tbody>
        <tbody v-else>
          <tr v-for="match in currentPageMatches" :key="match.id">
            <td>{{ match.id }}</td>
            <td>
              <div class="team-display">
                <TeamLogo :teamName="match.optionA" size="small" />
                <span>{{ match.optionA }}</span>
              </div>
            </td>
            <td>
              <div class="team-display">
                <TeamLogo :teamName="match.optionB" size="small" />
                <span>{{ match.optionB }}</span>
              </div>
            </td>
            <td>
              <span class="role-tag" :class="getMatchTypeClass(match.matchType)">
                {{ getMatchTypeText(match.matchType) }}
              </span>
            </td>
            <td>
              <span class="role-tag" :class="getSeriesClass(match.series)">
                {{ match.series }}
              </span>
            </td>
            <td>{{ formatMatchTime(match.matchTime) }}</td>
            <td>
              <span class="role-tag" :class="getStatusClass(match.status)">
                {{ getStatusText(match.status) }}
              </span>
            </td>
            <td class="action-cell">
              <button class="btn-edit" @click="editMatch(match)">编辑</button>
              <button
                class="btn-result"
                @click="setMatchResult(match)"
                :disabled="match.status === 'cancelled'">
                {{ match.status === 'completed' ? '重新设置结果' : '设置结果' }}
              </button>
              <button
                v-if="match.status === 'completed'"
                class="btn-result"
                @click="reverifyPredictions(match)">
                重新验证预测
              </button>
              <button
                size="small"
                class="danger-button"
                @click="deleteMatchConfirmed(match.id)"
              >
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
          @click="currentPage > 1 && handleCurrentChange(currentPage - 1)"
          :disabled="currentPage <= 1">上一页</button>
        <div class="page-numbers">
          <template v-if="totalPages <= 7">
            <button
              v-for="page in totalPages"
              :key="page"
              @click="handleCurrentChange(page)"
              class="page-number"
              :class="{ active: currentPage === page }">
              {{ page }}
            </button>
          </template>
          <template v-else>
            <!-- 当总页数较多时，显示较复杂的分页器 -->
            <button
              class="page-number"
              :class="{ active: currentPage === 1 }"
              @click="handleCurrentChange(1)">
              1
            </button>

            <!-- 左省略号 -->
            <span class="ellipsis" v-if="currentPage > 3">...</span>

            <!-- 当前页附近的页码 -->
            <template v-for="page in pageList" :key="page">
              <button
                v-if="page > 1 && page < totalPages"
                @click="handleCurrentChange(page)"
                class="page-number"
                :class="{ active: currentPage === page }">
                {{ page }}
              </button>
            </template>

            <!-- 右省略号 -->
            <span class="ellipsis" v-if="currentPage < totalPages - 2">...</span>

            <button
              class="page-number"
              :class="{ active: currentPage === totalPages }"
              @click="handleCurrentChange(totalPages)">
              {{ totalPages }}
            </button>
          </template>
        </div>
        <button
          class="next-btn"
          @click="currentPage < totalPages && handleCurrentChange(currentPage + 1)"
          :disabled="currentPage >= totalPages">下一页</button>
        <div class="page-size-select">
          <span>每页条数: </span>
          <select v-model="pageSize" @change="handleSizeChange" class="page-size-dropdown">
            <option :value="10">10</option>
            <option :value="20">20</option>
            <option :value="50">50</option>
            <option :value="100">100</option>
          </select>
        </div>
      </div>
    </div>

    <!-- 编辑/添加比赛对话框 -->
    <div v-if="showAddForm" class="modal-overlay">
      <div class="modal-content">
        <div class="modal-header">
          <h3>{{ currentMatch.id ? '编辑比赛' : '添加比赛' }}</h3>
          <button class="close-btn" @click="showAddForm = false">&times;</button>
        </div>
        <form @submit.prevent="saveMatch" class="user-form">
          <div class="form-group">
            <label for="option-a">队伍A</label>
            <div class="select-container">
              <select
                id="option-a"
                v-model="currentMatch.optionA"
                class="select-field"
                required
                @change="updateTitle"
              >
                <option value="" disabled>请选择队伍A</option>
                <option v-for="team in teamOptions" :key="team" :value="team">{{ team }}</option>
              </select>
            </div>
          </div>

          <div class="form-group">
            <label for="option-b">队伍B</label>
            <div class="select-container">
              <select
                id="option-b"
                v-model="currentMatch.optionB"
                class="select-field"
                required
                @change="updateTitle"
              >
                <option value="" disabled>请选择队伍B</option>
                <option v-for="team in teamsListForB" :key="team" :value="team">{{ team }}</option>
              </select>
            </div>
          </div>

          <div class="form-group" style="display: none;">
            <label for="title">比赛标题</label>
            <input
              id="title"
              v-model="currentMatch.title"
              type="text"
              class="input-field"
            />
          </div>

          <div class="form-group">
            <label for="match-type">比赛类型</label>
            <select
              id="match-type"
              v-model="currentMatch.matchType"
              class="select-field"
              required
            >
              <option value="regular">常规赛</option>
              <option value="playoff">季后赛</option>
              <option value="final">总决赛</option>
            </select>
          </div>

          <div class="form-group">
            <label for="series">比赛局数</label>
            <select
              id="series"
              v-model="currentMatch.series"
              class="select-field"
              required
            >
              <option value="BO2">BO2</option>
              <option value="BO3">BO3</option>
              <option value="BO5">BO5</option>
              <option value="BO7">BO7</option>
              <option value="BO9">BO9</option>
            </select>
          </div>

          <div class="form-group">
            <label for="match-time">比赛时间</label>
            <input
              id="match-time"
              v-model="matchTimeStr"
              type="datetime-local"
              class="input-field"
              required
            />
          </div>

          <div class="form-group">
            <label for="status">状态</label>
            <select
              id="status"
              v-model="currentMatch.status"
              class="select-field"
            >
              <option value="not_started">未开始</option>
              <option value="in_progress">进行中</option>
              <option value="completed">已完成</option>
              <option value="cancelled">已取消</option>
            </select>
          </div>

          <div class="form-group">
            <label for="tournament-type">赛事类型</label>
            <select
              id="tournament-type"
              v-model="currentMatch.tournamentType"
              class="select-field"
              required
              @change="updateTournamentStage"
            >
              <option value="spring">KPL春季赛</option>
              <option value="summer">KPL夏季赛</option>
              <option value="annual">KPL年度总决赛</option>
              <option value="challenger">KPL挑战者杯</option>
            </select>
          </div>

          <div class="form-group">
            <label for="tournament-stage">赛事阶段</label>
            <select
              id="tournament-stage"
              v-model="currentMatch.tournamentStage"
              class="select-field"
              required
            >
              <option v-if="['spring', 'summer'].includes(currentMatch.tournamentType)" value="regular">常规赛</option>
              <option v-if="['spring', 'summer'].includes(currentMatch.tournamentType)" value="playoff">季后赛</option>
              <option v-if="['annual', 'challenger'].includes(currentMatch.tournamentType)" value="group">小组赛</option>
              <option v-if="['annual', 'challenger'].includes(currentMatch.tournamentType)" value="knockout">淘汰赛</option>
            </select>
          </div>

          <div class="form-group">
            <label for="year">年份</label>
            <input
              id="year"
              v-model.number="currentMatch.year"
              type="number"
              class="input-field"
              min="2023"
              max="2030"
              required
            />
          </div>

          <div class="form-actions">
            <button type="button" class="cancel-btn" @click="showAddForm = false">取消</button>
            <button type="submit" class="submit-btn">保存</button>
          </div>
        </form>
      </div>
    </div>

    <!-- 设置比赛结果对话框 -->
    <div v-if="showResultForm" class="modal-overlay">
      <div class="modal-content">
        <div class="modal-header">
          <h3>设置比赛结果</h3>
          <button class="close-btn" @click="showResultForm = false">&times;</button>
        </div>
        <form @submit.prevent="submitMatchResult" class="user-form">


          <div class="form-group">
            <label>比赛结果</label>
            <div class="prediction-options">
              <div
                class="prediction-option"
                :class="{ selected: getWinner() === 'A' }"
              >
                <div class="team-display">
                  <TeamLogo :teamName="currentMatch.optionA" size="small" />
                  <span class="team-name">{{ currentMatch.optionA }}</span>
                </div>
                <input
                  type="number"
                  v-model.number="matchResult.scoreA"
                  min="0"
                  class="score-input"
                />
              </div>

              <div
                class="prediction-option"
                :class="{ selected: getWinner() === 'B' }"
              >
                <div class="team-display">
                  <TeamLogo :teamName="currentMatch.optionB" size="small" />
                  <span class="team-name">{{ currentMatch.optionB }}</span>
                </div>
                <input
                  type="number"
                  v-model.number="matchResult.scoreB"
                  min="0"
                  class="score-input"
                />
              </div>
            </div>
          </div>

          <div class="form-actions">
            <button type="button" class="cancel-btn" @click="showResultForm = false">取消</button>
            <button type="submit" class="submit-btn">提交结果</button>
          </div>
        </form>
      </div>
    </div>

    <!-- 战队管理 -->
    <div v-if="showTeamDialog" class="modal-overlay">
      <div class="modal-content">
        <button class="close-btn" @click="showTeamDialog = false">&times;</button>
        <h3 class="modal-title">{{ editingTeamId ? '编辑战队' : '新增战队' }}</h3>

        <form class="match-form" @submit.prevent="submitTeam">
          <div class="form-row">
            <div class="form-group">
              <label for="team-name">战队名称</label>
              <input
                id="team-name"
                v-model="teamForm.name"
                type="text"
                required
                placeholder="例如：北京JDG"
              />
            </div>
            <div class="form-group">
              <label for="team-short">简称（可选）</label>
              <input
                id="team-short"
                v-model="teamForm.shortName"
                type="text"
                placeholder="例如：JDG"
              />
            </div>
          </div>

          <div class="form-row">
            <div class="form-group full-width">
              <label for="team-logo">图标地址（URL，可外链）</label>
              <input
                id="team-logo"
                v-model="teamForm.logoUrl"
                type="text"
                placeholder="https://example.com/logo.png"
              />
            </div>
          </div>

          <div class="form-row">
            <div class="form-group checkbox-group">
              <label>
                <input type="checkbox" v-model="teamForm.isActive" />
                启用
              </label>
            </div>
          </div>

          <div class="form-actions">
            <button type="button" class="cancel-btn" @click="showTeamDialog = false">取消</button>
            <button type="submit" class="submit-btn" :disabled="teamFormLoading">
              {{ teamFormLoading ? '保存中...' : '保存' }}
            </button>
          </div>
        </form>

        <div class="team-list-block">
          <div class="team-list-header">
            <h4>已配置战队</h4>
            <button class="add-button" @click="openCreateTeam">新增</button>
          </div>
          <div v-if="teamStore.loading" class="loading-cell">加载中...</div>
          <div v-else-if="teamStore.teams.length === 0" class="empty-cell">暂无战队</div>
          <div v-else class="team-list">
            <div v-for="team in teamStore.teams" :key="team.id" class="team-row">
              <div class="team-info">
                <TeamLogo :teamName="team.name" :logoUrl="team.logoUrl" size="small" />
                <div class="team-text">
                  <div class="team-name">{{ team.name }}</div>
                  <div class="team-sub">
                    <span v-if="team.shortName">简称：{{ team.shortName }}</span>
                    <span>状态：{{ team.isActive ? '启用' : '停用' }}</span>
                  </div>
                </div>
              </div>
              <div class="team-actions">
                <button class="btn-edit" @click="openEditTeam(team)">编辑</button>
                <button class="btn-danger" @click="deleteTeam(team)">删除</button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, watch } from 'vue'
import axios from 'axios'
import { get } from '@/api/http'
import { useUserStore } from '@/stores/user'
import { useTeamStore } from '@/stores/team'
import type { Team, TeamPayload } from '@/api/team'
import { ElMessage, ElMessageBox } from 'element-plus'
import TeamLogo from '@/components/teams/TeamLogo.vue'

interface Match {
  id: number;
  title: string;
  optionA: string;
  optionB: string;
  matchTime: string;
  matchType: 'regular' | 'playoff' | 'final';
  series: 'BO2' | 'BO3' | 'BO5' | 'BO7' | 'BO9';
  status: 'not_started' | 'in_progress' | 'completed' | 'cancelled';
  description?: string;
  winner?: string;
  scoreA?: number;
  scoreB?: number;
  isActive?: boolean;
  tournamentType?: 'spring' | 'summer' | 'annual' | 'challenger';
  tournamentStage?: 'regular' | 'playoff' | 'group' | 'knockout';
  year?: number;
  createdAt?: string;
  updatedAt?: string;
}

interface MatchResult {
  winner?: 'A' | 'B';
  scoreA: number;
  scoreB: number;
}

const userStore = useUserStore()
const teamStore = useTeamStore()
const matches = ref<Match[]>([])
const loading = ref(false)
const showAddForm = ref(false)
const showResultForm = ref(false)
const idSearch = ref('')
const searchQuery = ref('')
const statusFilter = ref('')
const typeFilter = ref('')
const seriesFilter = ref('')
const tournamentTypeFilter = ref('')
const tournamentStageFilter = ref('')
const currentPage = ref(1)
const pageSize = ref(10)
const matchTimeStr = ref('')

// 动态战队列表（由后台配置）
const teamOptions = computed(() => teamStore.teams.map(t => t.name))

// 队伍B的可选列表，不包含已选的队伍A
const teamsListForB = computed(() => {
  return teamOptions.value.filter(team => team !== currentMatch.optionA)
})

// 战队管理
const showTeamDialog = ref(false)
const teamFormLoading = ref(false)
const editingTeamId = ref<number | null>(null)
const teamForm = reactive<TeamPayload>({
  name: '',
  shortName: '',
  logoUrl: '',
  isActive: true
})

const currentMatch = reactive<{
  id: number | null;
  title: string;
  optionA: string;
  optionB: string;
  matchTime: string | Date;
  matchType: 'regular' | 'playoff' | 'final';
  series: 'BO2' | 'BO3' | 'BO5' | 'BO7' | 'BO9';
  status: 'not_started' | 'in_progress' | 'completed' | 'cancelled';
  tournamentType: 'spring' | 'summer' | 'annual' | 'challenger';
  tournamentStage: 'regular' | 'playoff' | 'group' | 'knockout';
  year: number;
}>({
  id: null,
  title: '',
  optionA: '',
  optionB: '',
  matchTime: new Date(),
  matchType: 'regular',
  series: 'BO3',
  status: 'not_started',
  tournamentType: 'summer',
  tournamentStage: 'regular',
  year: 2025
})

const matchResult = reactive<MatchResult>({
  scoreA: 0,
  scoreB: 0
})

// 当编辑一个比赛时，将日期转换为input datetime-local可接受的字符串
watch(() => currentMatch.matchTime, (newVal) => {
  if (newVal) {
    try {
      let date: Date

      // 如果是字符串格式，尝试解析
      if (typeof newVal === 'string') {
        // 如果是 YYYY-MM-DD HH:MM 格式，手动解析
        if (newVal.match(/^\d{4}-\d{2}-\d{2} \d{2}:\d{2}$/)) {
          const [datePart, timePart] = newVal.split(' ')
          const [year, month, day] = datePart.split('-').map(Number)
          const [hours, minutes] = timePart.split(':').map(Number)

          // 创建Date对象，使用本地时间
          date = new Date(year, month - 1, day, hours, minutes)
          console.log('从字符串手动解析日期:', date)
        } else {
          // 其他字符串格式，直接解析
          date = new Date(newVal)
          console.log('从字符串直接解析日期:', date)
        }
      } else {
        // 如果已经是Date对象
        date = newVal as Date
      }

      // 检查日期是否有效
      if (isNaN(date.getTime())) {
        console.warn('无效的日期:', newVal)
        date = new Date() // 使用当前日期作为后备
      }

      // 格式化为 datetime-local 输入框需要的格式
      const year = date.getFullYear()
      const month = String(date.getMonth() + 1).padStart(2, '0')
      const day = String(date.getDate()).padStart(2, '0')
      const hours = String(date.getHours()).padStart(2, '0')
      const minutes = String(date.getMinutes()).padStart(2, '0')
      matchTimeStr.value = `${year}-${month}-${day}T${hours}:${minutes}`
      console.log('编辑比赛，日期字符串转换为:', matchTimeStr.value)
    } catch (error) {
      console.error('解析日期时出错:', error)
      // 出错时使用当前时间
      const now = new Date()
      const year = now.getFullYear()
      const month = String(now.getMonth() + 1).padStart(2, '0')
      const day = String(now.getDate()).padStart(2, '0')
      const hours = String(now.getHours()).padStart(2, '0')
      const minutes = String(now.getMinutes()).padStart(2, '0')
      matchTimeStr.value = `${year}-${month}-${day}T${hours}:${minutes}`
    }
  }
})

// 当datetime-local输入框变化时，更新当前比赛的matchTime
watch(matchTimeStr, (newVal) => {
  if (newVal) {
    try {
      // 创建Date对象用于格式化，但存储为字符串
      const date = new Date(newVal)

      // 检查日期是否有效
      if (isNaN(date.getTime())) {
        console.warn('无效的日期输入:', newVal)
        return
      }

      // 格式化为字符串格式（北京时间）
      const year = date.getFullYear()
      const month = String(date.getMonth() + 1).padStart(2, '0')
      const day = String(date.getDate()).padStart(2, '0')
      const hours = String(date.getHours()).padStart(2, '0')
      const minutes = String(date.getMinutes()).padStart(2, '0')

      // 存储为字符串格式，与后端保持一致
      currentMatch.matchTime = `${year}-${month}-${day} ${hours}:${minutes}`
      console.log('输入框日期变化，新的字符串格式:', currentMatch.matchTime)
    } catch (error) {
      console.error('处理日期输入时出错:', error)
    }
  }
})

// 过滤比赛列表
const filteredMatches = computed(() => {
  let result = [...matches.value]

  if (idSearch.value) {
    result = result.filter(match =>
      match.id.toString().includes(idSearch.value)
    )
  }

  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    result = result.filter(match =>
      match.optionA.toLowerCase().includes(query) ||
      match.optionB.toLowerCase().includes(query)
    )
  }

  if (statusFilter.value) {
    result = result.filter(match => match.status === statusFilter.value)
  }

  if (typeFilter.value) {
    result = result.filter(match => match.matchType === typeFilter.value)
  }

  if (seriesFilter.value) {
    result = result.filter(match => match.series === seriesFilter.value)
  }

  if (tournamentTypeFilter.value) {
    result = result.filter(match => match.tournamentType === tournamentTypeFilter.value)
  }

  if (tournamentStageFilter.value) {
    result = result.filter(match => match.tournamentStage === tournamentStageFilter.value)
  }

  return result
})

// 计算总页数
const totalPages = computed(() => {
  return Math.ceil(filteredMatches.value.length / pageSize.value) || 1
})

// 计算要显示的页码列表
const pageList = computed(() => {
  if (totalPages.value <= 7) {
    return Array.from({ length: totalPages.value }, (_, i) => i + 1)
  }

  const pages = []

  if (currentPage.value <= 3) {
    // 当前页靠近开始位置
    for (let i = 1; i <= 5; i++) {
      pages.push(i)
    }
  } else if (currentPage.value >= totalPages.value - 2) {
    // 当前页靠近结束位置
    for (let i = totalPages.value - 4; i <= totalPages.value; i++) {
      pages.push(i)
    }
  } else {
    // 当前页在中间位置
    for (let i = currentPage.value - 2; i <= currentPage.value + 2; i++) {
      pages.push(i)
    }
  }

  return pages
})

// 获取当前页的比赛数据
const currentPageMatches = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  const end = start + pageSize.value
  return filteredMatches.value.slice(start, end)
})

// 获取比赛数据
const fetchMatches = async () => {
  loading.value = true
  try {
    console.log('开始获取比赛列表...')

    // 使用封装的API调用
    try {
      const data = await get('/matches')
      console.log('API模块获取比赛列表成功:', JSON.stringify(data, null, 2))

      // 处理返回的数据
      if (Array.isArray(data)) {
        console.log('API模块返回了数组数据，长度:', data.length)
        matches.value = data.map(match => ({
          ...match,
          matchType: match.matchType || 'regular',
          series: match.series || 'BO3'
        }))
      } else {
        console.log('API模块返回了非数组数据:', typeof data)
        matches.value = []
      }
    } catch (apiError) {
      console.error('使用API模块获取比赛列表失败:', apiError)

      // 降级为直接使用axios
      console.log('尝试直接使用axios获取数据...')
      const response = await axios.get('/api/matches', {
        headers: {
          Authorization: `Bearer ${userStore.token}`
        }
      })
      console.log('获取比赛列表成功:', response)

      // 详细分析API响应结构
      console.log('API响应类型:', typeof response.data)
      console.log('API响应内容:', JSON.stringify(response.data, null, 2))

      // 临时数组存储处理后的比赛列表
      let fetchedMatches: Match[] = []

      // 根据API响应结构获取数据
      if (response.data && typeof response.data === 'object') {
        if (response.data.data && Array.isArray(response.data.data)) {
          // 标准结构：{ success, message, data: [...] }
          console.log('识别到标准API响应结构，数据条数:', response.data.data.length)
          fetchedMatches = response.data.data
        } else if (Array.isArray(response.data)) {
          // 数组结构：[...]
          console.log('识别到直接数组响应结构，数据条数:', response.data.length)
          fetchedMatches = response.data
        } else {
          // 其他可能的结构
          console.log('无法识别的API响应结构，尝试兼容处理')
          fetchedMatches = Array.isArray(response.data) ? response.data : []
        }
      } else {
        console.log('API响应为空或格式异常')
        fetchedMatches = []
      }

      // 处理每个比赛项，确保前端匹配类型字段的存在
      matches.value = fetchedMatches.map(match => {
        // 打印原始日期格式以便调试
        console.log(`比赛 ID ${match.id} 的原始日期:`, match.matchTime)

        return {
          ...match,
          // 如果后端没有这些字段，添加前端默认值
          matchType: match.matchType || 'regular',
          series: match.series || 'BO3'
        }
      })
    }

    // 打印处理后的匹配信息
    console.log('处理后的比赛列表:', matches.value)

    loading.value = false
  } catch (error: any) {
    console.error('获取比赛列表失败:', error)
    if (error.response) {
      console.error('错误状态:', error.response.status)
      console.error('错误数据:', JSON.stringify(error.response.data, null, 2))
    }
    ElMessage({
      message: '获取比赛列表失败',
      type: 'error',
      customClass: 'custom-message',
      offset: 80,
      duration: 3000
    })
    loading.value = false
    matches.value = []
  }
}

// 战队表单辅助
const resetTeamForm = () => {
  editingTeamId.value = null
  teamForm.name = ''
  teamForm.shortName = ''
  teamForm.logoUrl = ''
  teamForm.isActive = true
}

const openCreateTeam = () => {
  resetTeamForm()
  showTeamDialog.value = true
}

const openEditTeam = (team: Team) => {
  editingTeamId.value = team.id
  teamForm.name = team.name
  teamForm.shortName = team.shortName || ''
  teamForm.logoUrl = team.logoUrl || ''
  teamForm.isActive = team.isActive
  showTeamDialog.value = true
}

const submitTeam = async () => {
  teamFormLoading.value = true
  try {
    if (editingTeamId.value) {
      await teamStore.editTeam(editingTeamId.value, { ...teamForm })
      ElMessage.success('战队已更新')
    } else {
      await teamStore.addTeam({ ...teamForm })
      ElMessage.success('战队已创建')
    }
    showTeamDialog.value = false
  } catch (error: any) {
    ElMessage.error(error.response?.data?.message || error.message || '保存战队失败')
  } finally {
    teamFormLoading.value = false
  }
}

const deleteTeam = async (team: Team) => {
  try {
    await ElMessageBox.confirm(`确认删除战队「${team.name}」?`, '提示', { type: 'warning' })
    await teamStore.removeTeam(team.id)
    ElMessage.success('战队已删除')
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.message || error.message || '删除失败')
    }
  }
}

// 编辑比赛
const editMatch = (match: Match) => {
  console.log('编辑比赛，原始日期:', match.matchTime)

  // 直接使用字符串格式的日期时间
  let matchTimeValue = match.matchTime
  console.log('编辑比赛，原始日期值:', matchTimeValue)

  // 先设置除了optionB以外的所有字段
  Object.assign(currentMatch, {
    id: match.id,
    title: match.title || `${match.optionA} vs ${match.optionB}`,
    optionA: match.optionA,
    matchTime: matchTimeValue,
    matchType: match.matchType,
    series: match.series,
    status: match.status,
    tournamentType: match.tournamentType || 'summer',
    tournamentStage: match.tournamentStage || 'regular',
    year: match.year || 2025
  })

  // 等待下一个微任务周期，确保 teamsListForB 计算属性已更新
  setTimeout(() => {
    currentMatch.optionB = match.optionB
  }, 0)

  showAddForm.value = true
}

// 保存比赛
const saveMatch = async () => {
  try {
    // 确保title字段有值
    if (!currentMatch.title) {
      currentMatch.title = `${currentMatch.optionA} vs ${currentMatch.optionB}`
    }

    // 如果是Date对象，格式化为字符串；如果已经是字符串，直接使用
    let formattedDate: string

    if (typeof currentMatch.matchTime === 'string') {
      // 如果已经是字符串格式，检查是否符合要求的格式
      if (currentMatch.matchTime.match(/^\d{4}-\d{2}-\d{2} \d{2}:\d{2}$/)) {
        // 已经是正确的格式，直接使用
        formattedDate = currentMatch.matchTime
        console.log('使用现有的字符串格式:', formattedDate)
      } else {
        // 尝试解析并重新格式化
        try {
          const date = new Date(currentMatch.matchTime)
          const year = date.getFullYear()
          const month = String(date.getMonth() + 1).padStart(2, '0')
          const day = String(date.getDate()).padStart(2, '0')
          const hours = String(date.getHours()).padStart(2, '0')
          const minutes = String(date.getMinutes()).padStart(2, '0')
          formattedDate = `${year}-${month}-${day} ${hours}:${minutes}`
          console.log('从字符串解析并重新格式化:', formattedDate)
        } catch (error) {
          console.error('解析日期字符串出错:', error)
          // 出错时使用当前时间
          const now = new Date()
          const year = now.getFullYear()
          const month = String(now.getMonth() + 1).padStart(2, '0')
          const day = String(now.getDate()).padStart(2, '0')
          const hours = String(now.getHours()).padStart(2, '0')
          const minutes = String(now.getMinutes()).padStart(2, '0')
          formattedDate = `${year}-${month}-${day} ${hours}:${minutes}`
        }
      }
    } else {
      // 如果是Date对象，格式化为字符串
      const date = currentMatch.matchTime as Date
      const year = date.getFullYear()
      const month = String(date.getMonth() + 1).padStart(2, '0')
      const day = String(date.getDate()).padStart(2, '0')
      const hours = String(date.getHours()).padStart(2, '0')
      const minutes = String(date.getMinutes()).padStart(2, '0')
      formattedDate = `${year}-${month}-${day} ${hours}:${minutes}`
      console.log('从日期对象格式化:', formattedDate)
    }

    console.log('原始日期值:', currentMatch.matchTime)
    console.log('格式化后的日期字符串(北京时间):', formattedDate)

    // 将前端字段映射到后端 Go 接口期望的字段
    const tournamentMap: Record<string, string> = {
      spring: 'SPRING',
      summer: 'SUMMER',
      worlds: 'WORLDS'
    }
    // start_time 使用带时区的 ISO8601（保持东八区，不再转换为 UTC）
    const startIsoWithOffset = `${formattedDate.replace(' ', 'T')}:00+08:00`

    const matchData = {
      team_a: String(currentMatch.optionA).trim(),
      team_b: String(currentMatch.optionB).trim(),
      tournament: tournamentMap[currentMatch.tournamentType] || 'SUMMER',
      start_time: startIsoWithOffset
    }

    // 打印请求数据的完整JSON字符串
    console.log('请求数据的JSON字符串:', JSON.stringify(matchData))

    console.log('提交数据:', JSON.stringify(matchData, null, 2))

    // 统一请求配置
    const config = {
      headers: {
        'Authorization': `Bearer ${userStore.token}`,
        'Content-Type': 'application/json',
        'Accept': 'application/json'
      }
    }

    // 检查token
    console.log('使用的token:', userStore.token)

    let response
    let success = false

    try {
      if (currentMatch.id) {
        // 更新现有比赛 - 直接使用axios以获取完整响应
        console.log(`正在发送PUT请求到: /api/matches/${currentMatch.id}`)
        response = await axios.put(`/api/matches/${currentMatch.id}`, matchData, config)
        success = true
      } else {
        // 创建新比赛 - 直接使用axios以获取完整响应
        console.log('正在发送POST请求到: /api/matches')
        response = await axios.post('/api/matches', matchData, config)
        success = true
      }

      console.log('API响应成功:', response.status, response.statusText)
      console.log('API响应数据:', JSON.stringify(response.data, null, 2))
    } catch (requestError: any) {
      console.error('API请求错误:', requestError.message)
      if (requestError.config) {
        console.error('请求配置:', JSON.stringify(requestError.config, null, 2))
      }
      if (requestError.response) {
        console.error('错误响应:', requestError.response.status, requestError.response.statusText)
        console.error('错误数据:', JSON.stringify(requestError.response.data, null, 2))

        // 如果是400错误，尝试识别具体原因
        if (requestError.response.status === 400) {
          const errorData = requestError.response.data
          if (errorData.message && typeof errorData.message === 'string') {
            console.error('错误消息:', errorData.message)
          } else if (Array.isArray(errorData.message)) {
            console.error('验证错误:', errorData.message)
          }
        }
      }
      throw requestError
    }

    if (success) {
      // 使用Element Plus的消息提示
      ElMessage({
        message: currentMatch.id ? '比赛更新成功' : '比赛创建成功',
        type: 'success',
        customClass: 'custom-message',
        offset: 80,
        duration: 3000
      })

      showAddForm.value = false

      // 重置表单
      Object.assign(currentMatch, {
        id: null,
        title: '',
        optionA: '',
        optionB: '',
        matchTime: '',
        matchType: 'regular',
        series: 'BO3',
        status: 'not_started',
        tournamentType: 'summer',
        tournamentStage: 'regular',
        year: 2025
      })
      matchTimeStr.value = ''

      // 给UI一点时间更新，然后刷新比赛列表
      setTimeout(() => {
        console.log('开始刷新比赛列表...')
        fetchMatches()
      }, 500)
    } else {
      throw new Error('操作未成功完成: ' + (response?.data?.message || '未知错误'))
    }
  } catch (error: any) {
    console.error('保存比赛失败:', error)

    // 显示更详细的错误信息
    if (error.response) {
      console.error('错误状态:', error.response.status)
      console.error('错误数据:', JSON.stringify(error.response.data, null, 2))

      let errorMsg = `保存比赛失败`

      // 尝试提取具体错误信息
      if (error.response.data) {
        if (typeof error.response.data === 'string') {
          errorMsg = error.response.data
        } else if (error.response.data.message) {
          if (Array.isArray(error.response.data.message)) {
            errorMsg = error.response.data.message.join('\n')
          } else {
            errorMsg = error.response.data.message
          }
        } else if (error.response.data.details && Array.isArray(error.response.data.details)) {
          errorMsg = error.response.data.details.map((detail: any) => detail.message).join('\n')
        }
      }

      // 使用Element Plus的消息提示
      ElMessage({
        message: errorMsg,
        type: 'error',
        customClass: 'custom-message',
        offset: 80,
        duration: 3000
      })
    } else if (error.request) {
      ElMessage({
        message: `保存比赛失败: 请求已发送但未收到响应`,
        type: 'error',
        customClass: 'custom-message',
        offset: 80,
        duration: 3000
      })
    } else {
      ElMessage({
        message: `保存比赛失败: ${error.message || '未知错误'}`,
        type: 'error',
        customClass: 'custom-message',
        offset: 80,
        duration: 3000
      })
    }
  }
}

// 删除比赛直接确认
const deleteMatchConfirmed = async (matchId: number) => {
  try {
    const matchToDelete = matches.value.find(m => m.id === matchId);
    if (!matchToDelete) return;
    
    // 添加确认框
    await ElMessageBox.confirm(
      `确定要删除比赛 "${matchToDelete.optionA} vs ${matchToDelete.optionB}" 吗？`,
      '警告',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }
    );
    
    await axios.delete(`/api/matches/${matchId}`, {
      headers: { Authorization: `Bearer ${userStore.token}` }
    })
    ElMessage({
      message: '比赛删除成功',
      type: 'success',
      customClass: 'custom-message',
      offset: 80,
      duration: 3000
    })
    fetchMatches()
  } catch (error: any) {
    console.error('删除比赛失败:', error)
    
    // 如果是用户取消操作，显示不同的消息
    if (error === 'cancel' || error.toString().includes('cancel')) {
      ElMessage({
        message: '已取消删除',
        type: 'info',
        customClass: 'custom-message',
        offset: 80,
        duration: 3000
      })
      return;
    }
    
    ElMessage({
      message: '删除比赛失败',
      type: 'error',
      customClass: 'custom-message',
      offset: 80,
      duration: 3000
    })
  }
}

// 根据比分获取获胜方
const getWinner = () => {
  const scoreA = parseInt(matchResult.scoreA.toString()) || 0
  const scoreB = parseInt(matchResult.scoreB.toString()) || 0

  if (scoreA > scoreB) return 'A'
  if (scoreB > scoreA) return 'B'
  return '' // 平局或未设置比分
}

// 设置比赛结果
const setMatchResult = (match: Match) => {
  Object.assign(currentMatch, match)

  // 如果比赛已经有结果，使用现有结果初始化
  if (match.status === 'completed' && match.scoreA !== undefined && match.scoreB !== undefined) {
    Object.assign(matchResult, {
      scoreA: match.scoreA || 0,
      scoreB: match.scoreB || 0
    })
  } else {
    // 否则使用默认值
    Object.assign(matchResult, {
      scoreA: 0,
      scoreB: 0
    })
  }

  showResultForm.value = true
}

// 提交比赛结果
const submitMatchResult = async () => {
  try {
    // 根据比分自动判断获胜方
    const scoreA = parseInt(matchResult.scoreA.toString()) || 0
    const scoreB = parseInt(matchResult.scoreB.toString()) || 0

    // 确保比分不相等
    if (scoreA === scoreB) {
      ElMessage({
        message: '比分不能相等，请设置有效的比分',
        type: 'warning',
        customClass: 'custom-message',
        offset: 80,
        duration: 3000
      })
      return
    }

    const winner = scoreA > scoreB ? 'A' : 'B'
    matchResult.winner = winner

    console.log('开始设置比赛结果:', {
      matchId: currentMatch.id,
      optionA: currentMatch.optionA,
      optionB: currentMatch.optionB,
      winner: matchResult.winner,
      scoreA: matchResult.scoreA,
      scoreB: matchResult.scoreB
    })

    // 设置比赛结果（后端字段为 snake_case）
    const resultPayload = {
      score_a: scoreA,
      score_b: scoreB,
      winner
    }

    const resultResponse = await axios.post(`/api/matches/${currentMatch.id}/result`, resultPayload, {
      headers: { Authorization: `Bearer ${userStore.token}` }
    })

    console.log('比赛结果设置成功:', resultResponse.data)

    // 验证预测并计算积分
    try {
      console.log('开始验证预测并计算积分...')
      const verifyResponse = await axios.post(`/api/predictions/reverify/${currentMatch.id}`, {}, {
        headers: { Authorization: `Bearer ${userStore.token}` }
      })

      console.log('预测验证成功:', verifyResponse.data)

      ElMessage({
        message: '比赛结果设置成功，预测验证完成',
        type: 'success',
        customClass: 'custom-message',
        offset: 80,
        duration: 3000
      })
    } catch (verifyError: any) {
      console.error('验证预测失败:', verifyError)

      if (verifyError.response) {
        console.error('验证预测错误状态:', verifyError.response.status)
        console.error('验证预测错误数据:', JSON.stringify(verifyError.response.data, null, 2))
      }

      let errorMessage = '比赛结果设置成功，但预测验证失败'

      // 尝试提取具体错误信息
      if (verifyError.response && verifyError.response.data) {
        if (typeof verifyError.response.data === 'string') {
          errorMessage += ': ' + verifyError.response.data
        } else if (verifyError.response.data.message) {
          errorMessage += ': ' + verifyError.response.data.message
        }
      }

      ElMessage({
        message: errorMessage,
        type: 'warning',
        customClass: 'custom-message',
        offset: 80,
        duration: 3000
      })
    }

    showResultForm.value = false
    fetchMatches()
  } catch (error: any) {
    console.error('设置比赛结果失败:', error)

    if (error.response) {
      console.error('错误状态:', error.response.status)
      console.error('错误数据:', JSON.stringify(error.response.data, null, 2))
    }

    let errorMessage = '设置比赛结果失败'

    // 尝试提取具体错误信息
    if (error.response && error.response.data) {
      if (typeof error.response.data === 'string') {
        errorMessage += ': ' + error.response.data
      } else if (error.response.data.message) {
        errorMessage += ': ' + error.response.data.message
      }
    }

    ElMessage({
      message: errorMessage,
      type: 'error',
      customClass: 'custom-message',
      offset: 80,
      duration: 3000
    })
  }
}

// 获取状态类
const getStatusClass = (status: string): string => {
  const classes: Record<string, string> = {
    not_started: 'status-not-started',
    in_progress: 'status-in-progress',
    completed: 'status-completed',
    cancelled: 'status-cancelled'
  }
  return classes[status] || ''
}

// 获取状态文本
const getStatusText = (status: string): string => {
  const texts: Record<string, string> = {
    not_started: '未开始',
    in_progress: '进行中',
    completed: '已完成',
    cancelled: '已取消'
  }
  return texts[status] || '未知'
}

// 获取比赛类型样式类
const getMatchTypeClass = (type: string): string => {
  const classes: Record<string, string> = {
    regular: 'type-regular',
    playoff: 'type-playoff',
    final: 'type-final'
  }
  return classes[type] || ''
}

// 获取比赛类型文本
const getMatchTypeText = (type: string): string => {
  const texts: Record<string, string> = {
    regular: '常规赛',
    playoff: '季后赛',
    final: '总决赛'
  }
  return texts[type] || '未知'
}

// 获取比赛局数样式类
const getSeriesClass = (series: string): string => {
  const classes: Record<string, string> = {
    'BO2': 'series-bo2',
    'BO3': 'series-bo3',
    'BO5': 'series-bo5',
    'BO7': 'series-bo7',
    'BO9': 'series-bo9'
  }
  return classes[series] || ''
}

// 格式化比赛时间（北京时间）
const formatMatchTime = (dateString: string): string => {
  if (!dateString) return ''

  try {
    // 尝试直接解析日期字符串
    const date = new Date(dateString)

    // 检查日期是否有效
    if (isNaN(date.getTime())) {
      console.warn('无效的日期格式:', dateString)

      // 如果是 YYYY-MM-DD HH:MM 格式，直接返回
      if (typeof dateString === 'string' && dateString.match(/^\d{4}-\d{2}-\d{2} \d{2}:\d{2}$/)) {
        // 直接返回原始字符串，因为它已经是我们想要的格式
        return dateString
      }

      return dateString
    }

    // 格式化为 YYYY-MM-DD HH:MM 格式（北京时间）
    const year = date.getFullYear()
    const month = String(date.getMonth() + 1).padStart(2, '0')
    const day = String(date.getDate()).padStart(2, '0')
    const hours = String(date.getHours()).padStart(2, '0')
    const minutes = String(date.getMinutes()).padStart(2, '0')

    return `${year}-${month}-${day} ${hours}:${minutes}`
  } catch (error) {
    console.error('格式化日期出错:', error)
    return dateString
  }
}

// 搜索与过滤处理
const handleSearch = () => {
  currentPage.value = 1
}

// 分页处理
const handleSizeChange = () => {
  currentPage.value = 1
}

const handleCurrentChange = (val: number) => {
  currentPage.value = val
}

// 更新title字段
const updateTitle = () => {
  currentMatch.title = `${currentMatch.year}KPL${getTournamentTypeText(currentMatch.tournamentType)} ${getTournamentStageText(currentMatch.tournamentStage)} ${currentMatch.optionA} vs ${currentMatch.optionB}`
}

// 更新赛事阶段
const updateTournamentStage = () => {
  // 根据赛事类型设置默认的赛事阶段
  if (['spring', 'summer'].includes(currentMatch.tournamentType)) {
    if (!['regular', 'playoff'].includes(currentMatch.tournamentStage)) {
      currentMatch.tournamentStage = 'regular'
    }
  } else if (['annual', 'challenger'].includes(currentMatch.tournamentType)) {
    if (!['group', 'knockout'].includes(currentMatch.tournamentStage)) {
      currentMatch.tournamentStage = 'group'
    }
  }

  // 更新标题
  updateTitle()
}

// 获取赛事类型文本
const getTournamentTypeText = (type: string): string => {
  const types: Record<string, string> = {
    'spring': '春季赛',
    'summer': '夏季赛',
    'annual': '年度总决赛',
    'challenger': '挑战者杯'
  }
  return types[type] || '未知赛事'
}

// 获取赛事阶段文本
const getTournamentStageText = (stage: string): string => {
  const stages: Record<string, string> = {
    'regular': '常规赛',
    'playoff': '季后赛',
    'group': '小组赛',
    'knockout': '淘汰赛'
  }
  return stages[stage] || '未知阶段'
}

// 重新验证预测
const reverifyPredictions = async (match: Match) => {
  try {
    const response = await axios.post(`/api/predictions/reverify/${match.id}`, {}, {
      headers: { Authorization: `Bearer ${userStore.token}` }
    })

    console.log('预测重新验证成功:', response.data)

    ElMessage({
      message: '预测重新验证成功',
      type: 'success',
      customClass: 'custom-message',
      offset: 80,
      duration: 3000
    })
  } catch (error: any) {
    console.error('重新验证预测失败:', error)

    ElMessage({
      message: '重新验证预测失败: ' + (error.response?.data?.message || error.message),
      type: 'error',
      customClass: 'custom-message',
      offset: 80,
      duration: 3000
    })
  }
}

onMounted(() => {
  document.title = '比赛管理 | 预测系统'
  teamStore.ensureLoaded()
  fetchMatches()
})
</script>

<style scoped>
:global(.custom-message) {
  min-width: 240px;
  padding: 15px 20px;
  display: flex;
  align-items: center;
  border-radius: 6px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.12);
  position: fixed;
  top: 20px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 9999;
}
.admin-section {
  background-color: #fff;
  border-radius: 4px;
  padding: 20px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.05);
}

.team-display {
  display: flex;
  align-items: center;
  gap: 8px;
}

.page-title {
  margin-top: 0;
  margin-bottom: 20px;
  font-size: 20px;
  color: #303133;
  font-weight: 500;
}

.action-row {
  margin-bottom: 20px;
}

.search-filters {
  display: flex;
  align-items: flex-end;
  gap: 16px;
  flex-wrap: wrap;
}

.search-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 100px;
}

/* 搜索栏角色下拉框 */
.search-filters .search-item .select-container {
  width: 120px;
}

.search-filters .search-item .select-field {
  width: 120px;
}

.search-item label {
  font-size: 14px;
  color: #606266;
}

.input-field {
  height: 38px;
  padding: 0 15px;
  font-size: 14px;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  background-color: #fff;
  color: #606266;
  transition: border-color 0.2s;
  outline: none;
  box-sizing: border-box;
}

.input-field:focus {
  border-color: #409eff;
}

.textarea-field {
  height: auto;
  min-height: 80px;
  padding: 10px 15px;
  resize: vertical;
}

.select-container {
  position: relative;
  width: 100%;
  box-sizing: border-box;
}

.select-field {
  height: 38px;
  padding: 0 30px 0 15px;
  font-size: 14px;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  background-color: #fff;
  color: #606266;
  transition: border-color 0.2s;
  outline: none;
  width: 100%;
  appearance: none;
  box-sizing: border-box;
}

.select-field:focus {
  border-color: #409eff;
}

.select-container::after {
  content: "▼";
  font-size: 12px;
  color: #C0C4CC;
  position: absolute;
  right: 15px;
  top: 50%;
  transform: translateY(-50%);
  pointer-events: none;
}

.input-with-button {
  display: flex;
}

.search-button {
  padding: 0 20px;
  height: 38px;
  background-color: #409eff;
  border: 1px solid #409eff;
  color: #fff;
  font-size: 14px;
  border-radius: 0 4px 4px 0;
  cursor: pointer;
  transition: background-color 0.3s;
}

.search-button:hover {
  background-color: #66b1ff;
}

.add-button {
  padding: 0 20px;
  height: 38px;
  background-color: #409eff;
  border: 1px solid #409eff;
  color: #fff;
  font-size: 14px;
  border-radius: 4px;
  cursor: pointer;
  transition: background-color 0.3s;
}

.add-button.secondary {
  background-color: #10b981;
  border-color: #10b981;
}

.add-button:hover {
  background-color: #66b1ff;
}

.danger-button {
  padding: 5px 15px;
  background-color: #f56c6c;
  border: 1px solid #f56c6c;
  color: #fff;
  font-size: 14px;
  border-radius: 4px;
  cursor: pointer;
  transition: background-color 0.3s;
}

.danger-button:hover {
  background-color: #f78989;
}

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

.status-not-started {
  background-color: #909399;
  color: #fff;
}

.status-in-progress {
  background-color: #e6a23c;
  color: #fff;
}

.status-completed {
  background-color: #67c23a;
  color: #fff;
}

.status-cancelled {
  background-color: #f56c6c;
  color: #fff;
}

.action-cell {
  white-space: nowrap;
}

.btn-edit,
.btn-delete,
.btn-result {
  padding: 6px 12px;
  border-radius: 4px;
  border: 1px solid;
  background-color: transparent;
  font-size: 12px;
  cursor: pointer;
  margin: 0 4px;
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

.btn-result {
  color: #67c23a;
  border-color: #c2e7b0;
}

.btn-result:hover {
  color: #fff;
  background-color: #67c23a;
  border-color: #67c23a;
}

.btn-result:disabled,
.btn-delete:disabled,
.btn-edit:disabled {
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

.ellipsis {
  display: inline-block;
  width: 24px;
  text-align: center;
  font-weight: bold;
  letter-spacing: 2px;
}

.prev-btn,
.next-btn {
  min-width: 60px;
  height: 32px;
  padding: 0 10px;
  font-size: 14px;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  background-color: #fff;
  color: #606266;
  cursor: pointer;
}

.prev-btn:hover,
.next-btn:hover {
  color: #409eff;
  border-color: #c6e2ff;
  background-color: #ecf5ff;
}

.prev-btn:disabled,
.next-btn:disabled {
  color: #c0c4cc;
  cursor: not-allowed;
  border-color: #ebeef5;
  background-color: #f4f4f5;
}

.page-size-select {
  display: flex;
  align-items: center;
  margin-left: 15px;
  color: #606266;
  font-size: 14px;
}

.page-size-dropdown {
  margin-left: 5px;
  height: 32px;
  padding: 0 10px;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  outline: none;
}

/* 模态框 */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  width: 400px;
  max-width: 90%;
  background-color: #fff;
  border-radius: 4px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 15px 20px;
  border-bottom: 1px solid #ebeef5;
}

.modal-header h3 {
  font-size: 16px;
  font-weight: 500;
  color: #303133;
  margin: 0;
}

.close-btn {
  border: none;
  background: none;
  font-size: 20px;
  color: #909399;
  cursor: pointer;
}

.user-form {
  padding: 20px;
}

.form-group {
  margin-bottom: 20px;
  box-sizing: border-box;
  width: 100%;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  font-size: 14px;
  color: #606266;
}

.form-group .input-field,
.form-group .select-field {
  width: 100%;
  box-sizing: border-box;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 25px;
  box-sizing: border-box;
}

.cancel-btn,
.submit-btn {
  padding: 8px 20px;
  font-size: 14px;
  border-radius: 4px;
  cursor: pointer;
}

.cancel-btn {
  background-color: #fff;
  border: 1px solid #dcdfe6;
  color: #606266;
}

.submit-btn {
  background-color: #409eff;
  border: 1px solid #409eff;
  color: #fff;
}

.cancel-btn:hover {
  color: #409eff;
  border-color: #c6e2ff;
  background-color: #ecf5ff;
}

.submit-btn:hover {
  background-color: #66b1ff;
  border-color: #66b1ff;
}

.prediction-options {
  display: flex;
  justify-content: space-between;
  margin-bottom: 20px;
}

.prediction-option {
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 45%;
  padding: 15px;
  border-radius: 8px;
  border: 1px solid #eee;
  cursor: pointer;
  transition: all 0.2s;
}

.prediction-option:hover {
  background-color: #f9f9f9;
}

.prediction-option.selected {
  background-color: rgba(58, 123, 213, 0.1);
  border-color: #3a7bd5;
}

.score-input {
  width: 60px;
  height: 40px;
  margin-top: 10px;
  text-align: center;
  font-size: 18px;
  border: 1px solid #ddd;
  border-radius: 4px;
}

.type-regular {
  background-color: #409eff;
  color: #fff;
}

.type-playoff {
  background-color: #e6a23c;
  color: #fff;
}

.type-final {
  background-color: #9c27b0;
  color: #fff;
}

.series-bo2 {
  background-color: #42b983;
  color: #fff;
}

.series-bo3 {
  background-color: #3f51b5;
  color: #fff;
}

.series-bo5 {
  background-color: #ff9800;
  color: #fff;
}

.series-bo7 {
  background-color: #9c27b0;
  color: #fff;
}

.series-bo9 {
  background-color: #f44336;
  color: #fff;
}

/* 战队管理样式 */
.team-list-block {
  margin-top: 24px;
}

.team-list-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.team-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.team-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 12px;
  border: 1px solid #e5e7eb;
  border-radius: 10px;
  background: #fff;
}

.team-info {
  display: flex;
  align-items: center;
  gap: 10px;
}

.team-text .team-name {
  font-weight: 600;
}

.team-text .team-sub {
  font-size: 12px;
  color: #6b7280;
  display: flex;
  gap: 12px;
}

.team-actions {
  display: flex;
  gap: 8px;
}

.btn-danger {
  padding: 8px 12px;
  border: none;
  border-radius: 8px;
  background: #ef4444;
  color: white;
  cursor: pointer;
}

.btn-danger:hover {
  background: #dc2626;
}
</style>