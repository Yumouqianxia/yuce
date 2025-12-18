<template>
  <el-dialog
    v-model="visible"
    :title="`${sportType?.name} - 功能配置`"
    width="700px"
    :close-on-click-modal="false"
    @close="handleClose"
  >
    <div v-if="sportType" class="config-container">
      <!-- 配置预览 -->
      <el-alert
        :title="previewText"
        type="info"
        :closable="false"
        show-icon
        class="preview-alert"
      />
      
      <el-form
        ref="formRef"
        :model="configData"
        label-width="140px"
        @submit.prevent
      >
        <!-- 基础功能配置 -->
        <el-card class="config-section" shadow="never">
          <template #header>
            <div class="section-header">
              <span class="section-title">基础功能</span>
              <el-button
                type="text"
                size="small"
                @click="resetToDefault"
              >
                恢复默认
              </el-button>
            </div>
          </template>
          
          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item>
                <template #label>
                  <div class="config-label">
                    <span>实时通信</span>
                    <el-tooltip content="启用实时更新功能" placement="top">
                      <el-icon class="help-icon"><QuestionFilled /></el-icon>
                    </el-tooltip>
                  </div>
                </template>
                <el-switch
                  v-model="configData.enable_realtime"
                  active-text="启用"
                  inactive-text="禁用"
                />
              </el-form-item>
              
              <el-form-item>
                <template #label>
                  <div class="config-label">
                    <span>聊天功能</span>
                    <el-tooltip content="允许用户在比赛页面聊天" placement="top">
                      <el-icon class="help-icon"><QuestionFilled /></el-icon>
                    </el-tooltip>
                  </div>
                </template>
                <el-switch
                  v-model="configData.enable_chat"
                  active-text="启用"
                  inactive-text="禁用"
                />
              </el-form-item>
              
              <el-form-item>
                <template #label>
                  <div class="config-label">
                    <span>投票功能</span>
                    <el-tooltip content="允许用户对预测进行投票" placement="top">
                      <el-icon class="help-icon"><QuestionFilled /></el-icon>
                    </el-tooltip>
                  </div>
                </template>
                <el-switch
                  v-model="configData.enable_voting"
                  active-text="启用"
                  inactive-text="禁用"
                />
              </el-form-item>
            </el-col>
            
            <el-col :span="12">
              <el-form-item>
                <template #label>
                  <div class="config-label">
                    <span>预测功能</span>
                    <el-tooltip content="允许用户创建预测" placement="top">
                      <el-icon class="help-icon"><QuestionFilled /></el-icon>
                    </el-tooltip>
                  </div>
                </template>
                <el-switch
                  v-model="configData.enable_prediction"
                  active-text="启用"
                  inactive-text="禁用"
                />
              </el-form-item>
              
              <el-form-item>
                <template #label>
                  <div class="config-label">
                    <span>排行榜</span>
                    <el-tooltip content="显示积分排行榜" placement="top">
                      <el-icon class="help-icon"><QuestionFilled /></el-icon>
                    </el-tooltip>
                  </div>
                </template>
                <el-switch
                  v-model="configData.enable_leaderboard"
                  active-text="启用"
                  inactive-text="禁用"
                />
              </el-form-item>
            </el-col>
          </el-row>
        </el-card>
        
        <!-- 预测设置 -->
        <el-card class="config-section" shadow="never">
          <template #header>
            <span class="section-title">预测设置</span>
          </template>
          
          <el-form-item>
            <template #label>
              <div class="config-label">
                <span>允许修改预测</span>
                <el-tooltip content="用户可以修改已创建的预测" placement="top">
                  <el-icon class="help-icon"><QuestionFilled /></el-icon>
                </el-tooltip>
              </div>
            </template>
            <el-switch
              v-model="configData.allow_modification"
              active-text="允许"
              inactive-text="禁止"
            />
          </el-form-item>
          
          <el-form-item v-if="configData.allow_modification">
            <template #label>
              <div class="config-label">
                <span>最大修改次数</span>
                <el-tooltip content="每个预测最多可以修改的次数" placement="top">
                  <el-icon class="help-icon"><QuestionFilled /></el-icon>
                </el-tooltip>
              </div>
            </template>
            <el-input-number
              v-model="configData.max_modifications"
              :min="1"
              :max="10"
              style="width: 150px"
            />
            <span class="unit-text">次</span>
          </el-form-item>
          
          <el-form-item v-if="configData.allow_modification">
            <template #label>
              <div class="config-label">
                <span>修改截止时间</span>
                <el-tooltip content="比赛开始前多少分钟禁止修改预测" placement="top">
                  <el-icon class="help-icon"><QuestionFilled /></el-icon>
                </el-tooltip>
              </div>
            </template>
            <el-input-number
              v-model="configData.modification_deadline"
              :min="0"
              :max="1440"
              style="width: 150px"
            />
            <span class="unit-text">分钟前</span>
          </el-form-item>
        </el-card>
        
        <!-- 投票设置 -->
        <el-card v-if="configData.enable_voting" class="config-section" shadow="never">
          <template #header>
            <span class="section-title">投票设置</span>
          </template>
          
          <el-form-item>
            <template #label>
              <div class="config-label">
                <span>允许给自己投票</span>
                <el-tooltip content="用户可以给自己的预测投票" placement="top">
                  <el-icon class="help-icon"><QuestionFilled /></el-icon>
                </el-tooltip>
              </div>
            </template>
            <el-switch
              v-model="configData.enable_self_voting"
              active-text="允许"
              inactive-text="禁止"
            />
          </el-form-item>
          
          <el-form-item>
            <template #label>
              <div class="config-label">
                <span>每用户最大投票数</span>
                <el-tooltip content="每个用户最多可以投票的数量" placement="top">
                  <el-icon class="help-icon"><QuestionFilled /></el-icon>
                </el-tooltip>
              </div>
            </template>
            <el-input-number
              v-model="configData.max_votes_per_user"
              :min="1"
              :max="100"
              style="width: 150px"
            />
            <span class="unit-text">票</span>
          </el-form-item>
          
          <el-form-item>
            <template #label>
              <div class="config-label">
                <span>投票截止时间</span>
                <el-tooltip content="比赛开始前多少分钟禁止投票，0表示无限制" placement="top">
                  <el-icon class="help-icon"><QuestionFilled /></el-icon>
                </el-tooltip>
              </div>
            </template>
            <el-input-number
              v-model="configData.voting_deadline"
              :min="0"
              :max="1440"
              style="width: 150px"
            />
            <span class="unit-text">分钟前</span>
            <div class="config-tip">设置为0表示无时间限制</div>
          </el-form-item>
        </el-card>
        
        <!-- 配置模板 -->
        <el-card class="config-section" shadow="never">
          <template #header>
            <span class="section-title">配置模板</span>
          </template>
          
          <div class="template-buttons">
            <el-button
              type="primary"
              plain
              @click="applyTemplate('esports')"
            >
              应用电竞模板
            </el-button>
            <el-button
              type="success"
              plain
              @click="applyTemplate('traditional')"
            >
              应用传统体育模板
            </el-button>
            <el-button
              type="info"
              plain
              @click="exportConfig"
            >
              导出配置
            </el-button>
          </div>
          
          <div class="template-description">
            <p><strong>电竞模板：</strong>启用所有功能，允许聊天，修改次数较多</p>
            <p><strong>传统体育模板：</strong>禁用聊天，修改次数较少，有投票时间限制</p>
          </div>
        </el-card>
      </el-form>
    </div>
    
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleClose">取消</el-button>
        <el-button
          type="primary"
          :loading="loading"
          @click="handleSave"
        >
          保存配置
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { QuestionFilled } from '@element-plus/icons-vue'
import { useAdminStore } from '@/stores/admin'
import type { SportType, SportConfiguration, UpdateSportConfigurationRequest } from '@/types/admin'

interface Props {
  modelValue: boolean
  sportType?: SportType | null
}

interface Emits {
  (e: 'update:modelValue', value: boolean): void
  (e: 'success'): void
}

const props = withDefaults(defineProps<Props>(), {
  sportType: null
})

const emit = defineEmits<Emits>()

const adminStore = useAdminStore()

// 状态
const loading = ref(false)
const formRef = ref()

// 计算属性
const visible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

// 配置数据
const configData = ref<UpdateSportConfigurationRequest>({
  enable_realtime: true,
  enable_chat: false,
  enable_voting: true,
  enable_prediction: true,
  enable_leaderboard: true,
  allow_modification: true,
  max_modifications: 3,
  modification_deadline: 30,
  enable_self_voting: false,
  max_votes_per_user: 10,
  voting_deadline: 0
})

// 配置预览文本
const previewText = computed(() => {
  const features = []
  if (configData.value.enable_realtime) features.push('实时通信')
  if (configData.value.enable_chat) features.push('聊天')
  if (configData.value.enable_voting) features.push('投票')
  if (configData.value.enable_prediction) features.push('预测')
  if (configData.value.enable_leaderboard) features.push('排行榜')
  
  return `当前启用功能: ${features.join('、') || '无'}`
})

// 监听弹窗显示状态
watch(visible, (newVisible) => {
  if (newVisible && props.sportType) {
    initConfigData()
  }
})

// 初始化配置数据
const initConfigData = () => {
  if (props.sportType?.configuration) {
    const config = props.sportType.configuration
    configData.value = {
      enable_realtime: config.enable_realtime,
      enable_chat: config.enable_chat,
      enable_voting: config.enable_voting,
      enable_prediction: config.enable_prediction,
      enable_leaderboard: config.enable_leaderboard,
      allow_modification: config.allow_modification,
      max_modifications: config.max_modifications,
      modification_deadline: config.modification_deadline,
      enable_self_voting: config.enable_self_voting,
      max_votes_per_user: config.max_votes_per_user,
      voting_deadline: config.voting_deadline
    }
  } else {
    // 使用默认配置
    resetToDefault()
  }
}

// 恢复默认配置
const resetToDefault = () => {
  configData.value = {
    enable_realtime: true,
    enable_chat: false,
    enable_voting: true,
    enable_prediction: true,
    enable_leaderboard: true,
    allow_modification: true,
    max_modifications: 3,
    modification_deadline: 30,
    enable_self_voting: false,
    max_votes_per_user: 10,
    voting_deadline: 0
  }
  ElMessage.success('已恢复默认配置')
}

// 应用配置模板
const applyTemplate = async (templateType: 'esports' | 'traditional') => {
  try {
    const result = await ElMessageBox.confirm(
      `应用${templateType === 'esports' ? '电竞' : '传统体育'}模板将覆盖当前配置，是否继续？`,
      '确认操作',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    if (result === 'confirm') {
      if (templateType === 'esports') {
        configData.value = {
          enable_realtime: true,
          enable_chat: true,
          enable_voting: true,
          enable_prediction: true,
          enable_leaderboard: true,
          allow_modification: true,
          max_modifications: 3,
          modification_deadline: 30,
          enable_self_voting: false,
          max_votes_per_user: 10,
          voting_deadline: 0
        }
      } else {
        configData.value = {
          enable_realtime: true,
          enable_chat: false,
          enable_voting: true,
          enable_prediction: true,
          enable_leaderboard: true,
          allow_modification: true,
          max_modifications: 2,
          modification_deadline: 60,
          enable_self_voting: false,
          max_votes_per_user: 5,
          voting_deadline: 30
        }
      }
      
      ElMessage.success(`${templateType === 'esports' ? '电竞' : '传统体育'}模板已应用`)
    }
  } catch (error) {
    // 用户取消操作
  }
}

// 导出配置
const exportConfig = () => {
  const configJson = JSON.stringify(configData.value, null, 2)
  const blob = new Blob([configJson], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${props.sportType?.code || 'sport'}-config.json`
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
  
  ElMessage.success('配置已导出')
}

// 保存配置
const handleSave = async () => {
  if (!props.sportType) return
  
  try {
    loading.value = true
    
    await adminStore.updateSportConfiguration(props.sportType.id, configData.value)
    
    ElMessage.success('配置保存成功')
    emit('success')
    handleClose()
    
  } catch (error) {
    console.error('保存配置失败:', error)
    ElMessage.error(error instanceof Error ? error.message : '保存失败')
  } finally {
    loading.value = false
  }
}

// 关闭弹窗
const handleClose = () => {
  visible.value = false
}
</script>

<style scoped>
.config-container {
  max-height: 600px;
  overflow-y: auto;
}

.preview-alert {
  margin-bottom: 20px;
}

.config-section {
  margin-bottom: 20px;
}

.config-section:last-child {
  margin-bottom: 0;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.section-title {
  font-weight: 600;
  color: #303133;
}

.config-label {
  display: flex;
  align-items: center;
  gap: 4px;
}

.help-icon {
  color: #909399;
  cursor: help;
}

.unit-text {
  margin-left: 8px;
  color: #909399;
  font-size: 12px;
}

.config-tip {
  font-size: 11px;
  color: #C0C4CC;
  margin-top: 4px;
}

.template-buttons {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
}

.template-description {
  font-size: 12px;
  color: #606266;
  line-height: 1.5;
}

.template-description p {
  margin: 4px 0;
}

.dialog-footer {
  text-align: right;
}

:deep(.el-card__header) {
  padding: 12px 20px;
  border-bottom: 1px solid #EBEEF5;
}

:deep(.el-card__body) {
  padding: 20px;
}

:deep(.el-form-item) {
  margin-bottom: 16px;
}

:deep(.el-switch) {
  margin-left: 12px;
}

:deep(.el-input-number .el-input__inner) {
  text-align: left;
}
</style>