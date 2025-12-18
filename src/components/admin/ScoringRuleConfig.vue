<template>
  <el-dialog v-model="visible" :title="dialogTitle" width="900px" :close-on-click-modal="false"
    :close-on-press-escape="false" @close="handleClose">
    <el-form ref="formRef" :model="formData" :rules="formRules" label-width="140px" @submit.prevent>
      <!-- 基本信息 -->
      <el-card class="form-section" shadow="never">
        <template #header>
          <span class="section-title">基本信息</span>
        </template>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="规则名称" prop="name">
              <el-input v-model="formData.name" placeholder="请输入积分规则名称" maxlength="100" show-word-limit />
            </el-form-item>
          </el-col>

          <el-col :span="12">
            <el-form-item label="适用运动类型" prop="sport_type_id">
              <el-select v-model="formData.sport_type_id" placeholder="请选择运动类型" style="width: 100%" filterable>
                <el-option v-for="sport in accessibleSportTypes" :key="sport.id" :label="sport.name"
                  :value="sport.id" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="规则描述" prop="description">
          <el-input v-model="formData.description" type="textarea" :rows="3" placeholder="请输入积分规则描述" maxlength="500"
            show-word-limit />
        </el-form-item>
      </el-card>

      <!-- 基础积分设置 -->
      <el-card class="form-section" shadow="never">
        <template #header>
          <div class="section-header">
            <span class="section-title">基础积分设置</span>
            <el-button type="text" size="small" @click="showCalculationPreview = !showCalculationPreview">
              {{ showCalculationPreview ? '隐藏' : '显示' }}计算预览
            </el-button>
          </div>
        </template>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="基础积分" prop="base_points">
              <el-input-number v-model="formData.base_points" :min="1" :max="1000" placeholder="基础积分"
                style="width: 100%" />
              <div class="form-tip">预测正确时获得的基础积分</div>
            </el-form-item>

            <el-form-item>
              <template #label>
                <div class="config-label">
                  <span>启用难度系数</span>
                  <el-tooltip content="根据预测难度调整积分倍数" placement="top">
                    <el-icon class="help-icon">
                      <QuestionFilled />
                    </el-icon>
                  </el-tooltip>
                </div>
              </template>
              <el-switch v-model="formData.enable_difficulty" active-text="启用" inactive-text="禁用" />
            </el-form-item>

            <el-form-item v-if="formData.enable_difficulty" label="难度系数" prop="difficulty_multiplier">
              <el-input-number v-model="formData.difficulty_multiplier" :min="0.1" :max="10" :step="0.1" :precision="1"
                style="width: 100%" />
              <div class="form-tip">积分 = 基础积分 × 难度系数</div>
            </el-form-item>
          </el-col>

          <el-col :span="12">
            <!-- 计算预览 -->
            <div v-if="showCalculationPreview" class="calculation-preview">
              <h4>积分计算预览</h4>
              <div class="preview-item">
                <span class="preview-label">基础积分：</span>
                <span class="preview-value">{{ formData.base_points }}</span>
              </div>
              <div v-if="formData.enable_difficulty" class="preview-item">
                <span class="preview-label">难度系数：</span>
                <span class="preview-value">{{ formData.difficulty_multiplier }}</span>
              </div>
              <div class="preview-item total">
                <span class="preview-label">基础总分：</span>
                <span class="preview-value">{{ calculateBaseScore() }}</span>
              </div>
            </div>
          </el-col>
        </el-row>
      </el-card>

      <!-- 奖励组件 -->
      <el-card class="form-section" shadow="never">
        <template #header>
          <span class="section-title">奖励组件</span>
        </template>

        <el-row :gutter="20">
          <el-col :span="12">
            <!-- 投票奖励 -->
            <div class="reward-group">
              <el-form-item>
                <template #label>
                  <div class="config-label">
                    <span>投票奖励</span>
                    <el-tooltip content="根据预测获得的投票数给予额外积分" placement="top">
                      <el-icon class="help-icon">
                        <QuestionFilled />
                      </el-icon>
                    </el-tooltip>
                  </div>
                </template>
                <el-switch v-model="formData.enable_vote_reward" active-text="启用" inactive-text="禁用" />
              </el-form-item>

              <el-form-item v-if="formData.enable_vote_reward" label="每票积分" prop="vote_reward_points">
                <el-input-number v-model="formData.vote_reward_points" :min="0.1" :max="10" :step="0.1" :precision="1"
                  style="width: 120px" />
                <span class="unit-text">分/票</span>
              </el-form-item>

              <el-form-item v-if="formData.enable_vote_reward" label="最大奖励" prop="max_vote_reward">
                <el-input-number v-model="formData.max_vote_reward" :min="1" :max="100" style="width: 120px" />
                <span class="unit-text">分</span>
              </el-form-item>
            </div>
          </el-col>

          <el-col :span="12">
            <!-- 时间奖励 -->
            <div class="reward-group">
              <el-form-item>
                <template #label>
                  <div class="config-label">
                    <span>时间奖励</span>
                    <el-tooltip content="提前预测给予额外积分奖励" placement="top">
                      <el-icon class="help-icon">
                        <QuestionFilled />
                      </el-icon>
                    </el-tooltip>
                  </div>
                </template>
                <el-switch v-model="formData.enable_time_reward" active-text="启用" inactive-text="禁用" />
              </el-form-item>

              <el-form-item v-if="formData.enable_time_reward" label="奖励积分" prop="time_reward_points">
                <el-input-number v-model="formData.time_reward_points" :min="1" :max="50" style="width: 120px" />
                <span class="unit-text">分</span>
              </el-form-item>

              <el-form-item v-if="formData.enable_time_reward" label="奖励时间" prop="time_reward_hours">
                <el-input-number v-model="formData.time_reward_hours" :min="1" :max="168" style="width: 120px" />
                <span class="unit-text">小时前</span>
                <div class="form-tip">比赛开始前N小时预测可获得时间奖励</div>
              </el-form-item>
            </div>
          </el-col>
        </el-row>
      </el-card>

      <!-- 惩罚组件 -->
      <el-card class="form-section" shadow="never">
        <template #header>
          <span class="section-title">惩罚组件</span>
        </template>

        <el-row :gutter="20">
          <el-col :span="12">
            <!-- 修改惩罚 -->
            <div class="penalty-group">
              <el-form-item>
                <template #label>
                  <div class="config-label">
                    <span>修改惩罚</span>
                    <el-tooltip content="修改预测时扣除积分" placement="top">
                      <el-icon class="help-icon">
                        <QuestionFilled />
                      </el-icon>
                    </el-tooltip>
                  </div>
                </template>
                <el-switch v-model="formData.enable_modify_penalty" active-text="启用" inactive-text="禁用" />
              </el-form-item>

              <el-form-item v-if="formData.enable_modify_penalty" label="每次扣分" prop="modify_penalty_points">
                <el-input-number v-model="formData.modify_penalty_points" :min="0.1" :max="20" :step="0.1"
                  :precision="1" style="width: 120px" />
                <span class="unit-text">分/次</span>
              </el-form-item>

              <el-form-item v-if="formData.enable_modify_penalty" label="最大扣分" prop="max_modify_penalty">
                <el-input-number v-model="formData.max_modify_penalty" :min="1" :max="100" style="width: 120px" />
                <span class="unit-text">分</span>
              </el-form-item>
            </div>
          </el-col>

          <el-col :span="12">
            <!-- 积分计算示例 -->
            <div class="calculation-example">
              <h4>积分计算示例</h4>
              <div class="example-scenario">
                <p><strong>场景：</strong>用户预测正确，获得5票，提前24小时预测，修改了1次</p>
                <div class="calculation-steps">
                  <div class="step">
                    <span>基础积分：</span>
                    <span>{{ formData.base_points }}</span>
                  </div>
                  <div v-if="formData.enable_difficulty" class="step">
                    <span>难度加成：</span>
                    <span>{{ formData.base_points }} × {{ formData.difficulty_multiplier }} = {{ (formData.base_points *
                      formData.difficulty_multiplier).toFixed(1) }}</span>
                  </div>
                  <div v-if="formData.enable_vote_reward" class="step">
                    <span>投票奖励：</span>
                    <span>5票 × {{ formData.vote_reward_points }} = {{ Math.min(5 * formData.vote_reward_points,
                      formData.max_vote_reward) }}</span>
                  </div>
                  <div v-if="formData.enable_time_reward" class="step">
                    <span>时间奖励：</span>
                    <span>{{ formData.time_reward_points }}</span>
                  </div>
                  <div v-if="formData.enable_modify_penalty" class="step penalty">
                    <span>修改惩罚：</span>
                    <span>-{{ formData.modify_penalty_points }}</span>
                  </div>
                  <div class="step total">
                    <span><strong>总积分：</strong></span>
                    <span><strong>{{ calculateExampleScore() }}</strong></span>
                  </div>
                </div>
              </div>
            </div>
          </el-col>
        </el-row>
      </el-card>

      <!-- 规则模板 -->
      <el-card class="form-section" shadow="never">
        <template #header>
          <span class="section-title">规则模板</span>
        </template>

        <div class="template-buttons">
          <el-button type="primary" plain @click="applyTemplate('basic')">
            基础模板
          </el-button>
          <el-button type="success" plain @click="applyTemplate('competitive')">
            竞技模板
          </el-button>
          <el-button type="info" plain @click="applyTemplate('casual')">
            休闲模板
          </el-button>
          <el-button type="warning" plain @click="validateRule">
            验证规则
          </el-button>
        </div>

        <div class="template-description">
          <p><strong>基础模板：</strong>简单的积分计算，只有基础积分</p>
          <p><strong>竞技模板：</strong>完整的积分系统，包含所有奖励和惩罚机制</p>
          <p><strong>休闲模板：</strong>温和的积分系统，有奖励但无惩罚</p>
        </div>
      </el-card>
    </el-form>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleClose">取消</el-button>
        <el-button type="primary" :loading="loading" @click="handleSubmit">
          {{ isEdit ? '更新' : '创建' }}
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { QuestionFilled } from '@element-plus/icons-vue'
import { useAdminStore } from '@/stores/admin'
import { usePermissionStore } from '@/stores/permissions'
import type { ScoringRule, CreateScoringRuleRequest, UpdateScoringRuleRequest } from '@/types/admin'

interface Props {
  modelValue: boolean
  scoringRule?: ScoringRule | null
}

interface Emits {
  (e: 'update:modelValue', value: boolean): void
  (e: 'success'): void
}

const props = withDefaults(defineProps<Props>(), {
  scoringRule: null
})

const emit = defineEmits<Emits>()

const adminStore = useAdminStore()
const permissionStore = usePermissionStore()

// 表单引用
const formRef = ref<FormInstance>()

// 状态
const loading = ref(false)
const showCalculationPreview = ref(true)

// 计算属性
const visible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const isEdit = computed(() => !!props.scoringRule)

const dialogTitle = computed(() => {
  return isEdit.value ? '编辑积分规则' : '创建积分规则'
})

// 可访问的运动类型
const accessibleSportTypes = computed(() => {
  return permissionStore.filterAccessibleSportTypes(adminStore.allSportTypes)
})

// 表单数据
const formData = ref({
  name: '',
  sport_type_id: null as number | null,
  description: '',
  base_points: 10,
  enable_difficulty: false,
  difficulty_multiplier: 1.5,
  enable_vote_reward: false,
  vote_reward_points: 1,
  max_vote_reward: 10,
  enable_time_reward: false,
  time_reward_points: 5,
  time_reward_hours: 24,
  enable_modify_penalty: false,
  modify_penalty_points: 2,
  max_modify_penalty: 10
})

// 表单验证规则
const formRules: FormRules = {
  name: [
    { required: true, message: '请输入规则名称', trigger: 'blur' },
    { min: 2, max: 100, message: '规则名称长度在 2 到 100 个字符', trigger: 'blur' }
  ],
  sport_type_id: [
    { required: true, message: '请选择运动类型', trigger: 'change' }
  ],
  base_points: [
    { required: true, message: '请输入基础积分', trigger: 'blur' },
    { type: 'number', min: 1, max: 1000, message: '基础积分范围为 1-1000', trigger: 'blur' }
  ]
}

// 监听弹窗显示状态
watch(visible, (newVisible) => {
  if (newVisible) {
    initFormData()
    // 确保运动类型数据已加载
    if (adminStore.allSportTypes.length === 0) {
      adminStore.fetchAllSportTypes()
    }
  }
})

// 初始化表单数据
const initFormData = () => {
  if (isEdit.value && props.scoringRule) {
    // 编辑模式，填充现有数据
    formData.value = {
      name: props.scoringRule.name,
      sport_type_id: props.scoringRule.sport_type_id,
      description: props.scoringRule.description || '',
      base_points: props.scoringRule.base_points,
      enable_difficulty: props.scoringRule.enable_difficulty,
      difficulty_multiplier: props.scoringRule.difficulty_multiplier,
      enable_vote_reward: props.scoringRule.enable_vote_reward,
      vote_reward_points: props.scoringRule.vote_reward_points,
      max_vote_reward: props.scoringRule.max_vote_reward,
      enable_time_reward: props.scoringRule.enable_time_reward,
      time_reward_points: props.scoringRule.time_reward_points,
      time_reward_hours: props.scoringRule.time_reward_hours,
      enable_modify_penalty: props.scoringRule.enable_modify_penalty,
      modify_penalty_points: props.scoringRule.modify_penalty_points,
      max_modify_penalty: props.scoringRule.max_modify_penalty
    }
  } else {
    // 创建模式，使用默认值
    formData.value = {
      name: '',
      sport_type_id: null,
      description: '',
      base_points: 10,
      enable_difficulty: false,
      difficulty_multiplier: 1.5,
      enable_vote_reward: false,
      vote_reward_points: 1,
      max_vote_reward: 10,
      enable_time_reward: false,
      time_reward_points: 5,
      time_reward_hours: 24,
      enable_modify_penalty: false,
      modify_penalty_points: 2,
      max_modify_penalty: 10
    }
  }

  // 清除表单验证状态
  nextTick(() => {
    formRef.value?.clearValidate()
  })
}

// 计算基础积分
const calculateBaseScore = () => {
  let score = formData.value.base_points
  if (formData.value.enable_difficulty) {
    score *= formData.value.difficulty_multiplier
  }
  return score.toFixed(1)
}

// 计算示例积分
const calculateExampleScore = () => {
  let score = formData.value.base_points

  // 难度加成
  if (formData.value.enable_difficulty) {
    score *= formData.value.difficulty_multiplier
  }

  // 投票奖励
  if (formData.value.enable_vote_reward) {
    const voteReward = Math.min(5 * formData.value.vote_reward_points, formData.value.max_vote_reward)
    score += voteReward
  }

  // 时间奖励
  if (formData.value.enable_time_reward) {
    score += formData.value.time_reward_points
  }

  // 修改惩罚
  if (formData.value.enable_modify_penalty) {
    score -= formData.value.modify_penalty_points
  }

  return Math.max(0, score).toFixed(1)
}

// 应用规则模板
const applyTemplate = async (templateType: 'basic' | 'competitive' | 'casual') => {
  try {
    const templateNames = {
      basic: '基础模板',
      competitive: '竞技模板',
      casual: '休闲模板'
    }

    const result = await ElMessageBox.confirm(
      `应用${templateNames[templateType]}将覆盖当前配置，是否继续？`,
      '确认操作',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    if (result === 'confirm') {
      const templates = {
        basic: {
          base_points: 10,
          enable_difficulty: false,
          difficulty_multiplier: 1.0,
          enable_vote_reward: false,
          enable_time_reward: false,
          enable_modify_penalty: false
        },
        competitive: {
          base_points: 15,
          enable_difficulty: true,
          difficulty_multiplier: 1.5,
          enable_vote_reward: true,
          vote_reward_points: 1,
          max_vote_reward: 5,
          enable_time_reward: true,
          time_reward_points: 3,
          time_reward_hours: 24,
          enable_modify_penalty: true,
          modify_penalty_points: 2,
          max_modify_penalty: 6
        },
        casual: {
          base_points: 8,
          enable_difficulty: false,
          difficulty_multiplier: 1.0,
          enable_vote_reward: true,
          vote_reward_points: 1,
          max_vote_reward: 3,
          enable_time_reward: false,
          enable_modify_penalty: false
        }
      }

      Object.assign(formData.value, templates[templateType])
      ElMessage.success(`${templateNames[templateType]}已应用`)
    }
  } catch (error) {
    // 用户取消操作
  }
}

// 验证规则
const validateRule = () => {
  const warnings = []
  const suggestions = []

  // 检查基础积分
  if (formData.value.base_points < 5) {
    warnings.push('基础积分较低，可能影响用户积极性')
  }
  if (formData.value.base_points > 50) {
    warnings.push('基础积分较高，可能导致积分通胀')
  }

  // 检查难度系数
  if (formData.value.enable_difficulty && formData.value.difficulty_multiplier > 3) {
    warnings.push('难度系数过高，可能导致积分差距过大')
  }

  // 检查奖励平衡
  if (formData.value.enable_vote_reward && formData.value.enable_time_reward) {
    const maxReward = formData.value.max_vote_reward + formData.value.time_reward_points
    if (maxReward > formData.value.base_points) {
      suggestions.push('奖励积分总和超过基础积分，建议调整平衡')
    }
  }

  // 检查惩罚机制
  if (formData.value.enable_modify_penalty && formData.value.modify_penalty_points > formData.value.base_points * 0.5) {
    warnings.push('修改惩罚过重，可能影响用户体验')
  }

  // 显示验证结果
  let message = '规则验证完成！\n\n'
  if (warnings.length > 0) {
    message += '⚠️ 警告：\n' + warnings.map(w => `• ${w}`).join('\n') + '\n\n'
  }
  if (suggestions.length > 0) {
    message += '💡 建议：\n' + suggestions.map(s => `• ${s}`).join('\n') + '\n\n'
  }
  if (warnings.length === 0 && suggestions.length === 0) {
    message += '✅ 规则配置合理，没有发现问题'
  }

  ElMessageBox.alert(message, '规则验证结果', {
    confirmButtonText: '确定',
    type: warnings.length > 0 ? 'warning' : 'success'
  })
}

// 提交表单
const handleSubmit = async () => {
  if (!formRef.value) return

  try {
    // 验证表单
    await formRef.value.validate()

    loading.value = true

    if (isEdit.value && props.scoringRule) {
      // 更新积分规则
      const updateData: UpdateScoringRuleRequest = {
        name: formData.value.name,
        description: formData.value.description || undefined,
        base_points: formData.value.base_points,
        enable_difficulty: formData.value.enable_difficulty,
        difficulty_multiplier: formData.value.difficulty_multiplier,
        enable_vote_reward: formData.value.enable_vote_reward,
        vote_reward_points: formData.value.vote_reward_points,
        max_vote_reward: formData.value.max_vote_reward,
        enable_time_reward: formData.value.enable_time_reward,
        time_reward_points: formData.value.time_reward_points,
        time_reward_hours: formData.value.time_reward_hours,
        enable_modify_penalty: formData.value.enable_modify_penalty,
        modify_penalty_points: formData.value.modify_penalty_points,
        max_modify_penalty: formData.value.max_modify_penalty
      }

      await adminStore.updateScoringRule(props.scoringRule.id, updateData)
      ElMessage.success('积分规则更新成功')
    } else {
      // 创建积分规则
      const createData: CreateScoringRuleRequest = {
        sport_type_id: formData.value.sport_type_id!,
        name: formData.value.name,
        description: formData.value.description || undefined,
        base_points: formData.value.base_points,
        enable_difficulty: formData.value.enable_difficulty,
        difficulty_multiplier: formData.value.difficulty_multiplier,
        enable_vote_reward: formData.value.enable_vote_reward,
        vote_reward_points: formData.value.vote_reward_points,
        max_vote_reward: formData.value.max_vote_reward,
        enable_time_reward: formData.value.enable_time_reward,
        time_reward_points: formData.value.time_reward_points,
        time_reward_hours: formData.value.time_reward_hours,
        enable_modify_penalty: formData.value.enable_modify_penalty,
        modify_penalty_points: formData.value.modify_penalty_points,
        max_modify_penalty: formData.value.max_modify_penalty
      }

      await adminStore.createScoringRule(createData)
      ElMessage.success('积分规则创建成功')
    }

    emit('success')
    handleClose()

  } catch (error) {
    console.error('提交失败:', error)
    ElMessage.error(error instanceof Error ? error.message : '操作失败')
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
.form-section {
  margin-bottom: 20px;
}

.form-section:last-child {
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

.form-tip {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
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

.reward-group,
.penalty-group {
  border: 1px solid #EBEEF5;
  border-radius: 4px;
  padding: 16px;
  margin-bottom: 16px;
}

.calculation-preview {
  background: #f5f7fa;
  border-radius: 4px;
  padding: 16px;
  margin-top: 16px;
}

.calculation-preview h4 {
  margin: 0 0 12px 0;
  font-size: 14px;
  color: #303133;
}

.preview-item {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
}

.preview-item.total {
  border-top: 1px solid #DCDFE6;
  padding-top: 8px;
  font-weight: 600;
}

.preview-label {
  color: #606266;
}

.preview-value {
  color: #303133;
  font-weight: 500;
}

.calculation-example {
  background: #fff7e6;
  border: 1px solid #ffd591;
  border-radius: 4px;
  padding: 16px;
  margin-top: 16px;
}

.calculation-example h4 {
  margin: 0 0 12px 0;
  font-size: 14px;
  color: #d48806;
}

.example-scenario p {
  margin: 0 0 12px 0;
  font-size: 13px;
  color: #8c8c8c;
}

.calculation-steps {
  font-size: 12px;
}

.step {
  display: flex;
  justify-content: space-between;
  margin-bottom: 4px;
  padding: 2px 0;
}

.step.penalty {
  color: #f5222d;
}

.step.total {
  border-top: 1px solid #d9d9d9;
  padding-top: 8px;
  margin-top: 8px;
  font-size: 13px;
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