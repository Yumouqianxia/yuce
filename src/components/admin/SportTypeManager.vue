<template>
  <el-dialog
    v-model="visible"
    :title="dialogTitle"
    width="800px"
    :close-on-click-modal="false"
    :close-on-press-escape="false"
    @close="handleClose"
  >
    <el-form
      ref="formRef"
      :model="formData"
      :rules="formRules"
      label-width="120px"
      @submit.prevent
    >
      <!-- 基本信息 -->
      <el-card class="form-section" shadow="never">
        <template #header>
          <span class="section-title">基本信息</span>
        </template>
        
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="运动名称" prop="name">
              <el-input
                v-model="formData.name"
                placeholder="请输入运动名称"
                maxlength="100"
                show-word-limit
              />
            </el-form-item>
          </el-col>
          
          <el-col :span="12">
            <el-form-item label="运动代码" prop="code">
              <el-input
                v-model="formData.code"
                placeholder="请输入运动代码"
                maxlength="20"
                show-word-limit
                :disabled="isEdit"
              />
              <div class="form-tip">代码创建后不可修改，用于系统内部标识</div>
            </el-form-item>
          </el-col>
        </el-row>
        
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="运动类别" prop="category">
              <el-select
                v-model="formData.category"
                placeholder="请选择运动类别"
                style="width: 100%"
              >
                <el-option
                  v-for="(label, value) in categoryOptions"
                  :key="value"
                  :label="label"
                  :value="value"
                />
              </el-select>
            </el-form-item>
          </el-col>
          
          <el-col :span="12">
            <el-form-item label="排序权重" prop="sort_order">
              <el-input-number
                v-model="formData.sort_order"
                :min="0"
                :max="999"
                placeholder="排序权重"
                style="width: 100%"
              />
              <div class="form-tip">数值越大排序越靠前</div>
            </el-form-item>
          </el-col>
        </el-row>
        
        <el-form-item label="运动描述" prop="description">
          <el-input
            v-model="formData.description"
            type="textarea"
            :rows="3"
            placeholder="请输入运动描述"
            maxlength="500"
            show-word-limit
          />
        </el-form-item>
        
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="运动图标" prop="icon">
              <el-input
                v-model="formData.icon"
                placeholder="请输入图标URL"
              />
            </el-form-item>
          </el-col>
          
          <el-col :span="12">
            <el-form-item label="横幅图片" prop="banner">
              <el-input
                v-model="formData.banner"
                placeholder="请输入横幅图片URL"
              />
            </el-form-item>
          </el-col>
        </el-row>
      </el-card>
      
      <!-- 功能配置 -->
      <el-card class="form-section" shadow="never">
        <template #header>
          <span class="section-title">功能配置</span>
          <el-button
            type="text"
            size="small"
            @click="applyPresetConfig"
          >
            应用预设配置
          </el-button>
        </template>
        
        <el-row :gutter="20">
          <el-col :span="12">
            <div class="config-group">
              <h4>基础功能</h4>
              <el-form-item>
                <el-checkbox v-model="formData.configuration.enable_realtime">
                  启用实时通信
                </el-checkbox>
                <div class="config-tip">启用实时更新功能</div>
              </el-form-item>
              
              <el-form-item>
                <el-checkbox v-model="formData.configuration.enable_chat">
                  启用聊天功能
                </el-checkbox>
                <div class="config-tip">允许用户在比赛页面聊天</div>
              </el-form-item>
              
              <el-form-item>
                <el-checkbox v-model="formData.configuration.enable_voting">
                  启用投票功能
                </el-checkbox>
                <div class="config-tip">允许用户对预测进行投票</div>
              </el-form-item>
              
              <el-form-item>
                <el-checkbox v-model="formData.configuration.enable_prediction">
                  启用预测功能
                </el-checkbox>
                <div class="config-tip">允许用户创建预测</div>
              </el-form-item>
              
              <el-form-item>
                <el-checkbox v-model="formData.configuration.enable_leaderboard">
                  启用排行榜
                </el-checkbox>
                <div class="config-tip">显示积分排行榜</div>
              </el-form-item>
            </div>
          </el-col>
          
          <el-col :span="12">
            <div class="config-group">
              <h4>预测设置</h4>
              <el-form-item>
                <el-checkbox v-model="formData.configuration.allow_modification">
                  允许修改预测
                </el-checkbox>
                <div class="config-tip">用户可以修改已创建的预测</div>
              </el-form-item>
              
              <el-form-item v-if="formData.configuration.allow_modification">
                <label class="config-label">最大修改次数</label>
                <el-input-number
                  v-model="formData.configuration.max_modifications"
                  :min="1"
                  :max="10"
                  size="small"
                  style="width: 120px"
                />
              </el-form-item>
              
              <el-form-item v-if="formData.configuration.allow_modification">
                <label class="config-label">修改截止时间</label>
                <el-input-number
                  v-model="formData.configuration.modification_deadline"
                  :min="0"
                  :max="1440"
                  size="small"
                  style="width: 120px"
                />
                <span class="config-unit">分钟前</span>
                <div class="config-tip">比赛开始前N分钟禁止修改</div>
              </el-form-item>
            </div>
            
            <div class="config-group">
              <h4>投票设置</h4>
              <el-form-item v-if="formData.configuration.enable_voting">
                <el-checkbox v-model="formData.configuration.enable_self_voting">
                  允许给自己投票
                </el-checkbox>
                <div class="config-tip">用户可以给自己的预测投票</div>
              </el-form-item>
              
              <el-form-item v-if="formData.configuration.enable_voting">
                <label class="config-label">每用户最大投票数</label>
                <el-input-number
                  v-model="formData.configuration.max_votes_per_user"
                  :min="1"
                  :max="100"
                  size="small"
                  style="width: 120px"
                />
              </el-form-item>
              
              <el-form-item v-if="formData.configuration.enable_voting">
                <label class="config-label">投票截止时间</label>
                <el-input-number
                  v-model="formData.configuration.voting_deadline"
                  :min="0"
                  :max="1440"
                  size="small"
                  style="width: 120px"
                />
                <span class="config-unit">分钟前</span>
                <div class="config-tip">比赛开始前N分钟禁止投票，0表示无限制</div>
              </el-form-item>
            </div>
          </el-col>
        </el-row>
      </el-card>
    </el-form>
    
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleClose">取消</el-button>
        <el-button
          type="primary"
          :loading="loading"
          @click="handleSubmit"
        >
          {{ isEdit ? '更新' : '创建' }}
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { useAdminStore } from '@/stores/admin'
import { usePermissionStore } from '@/stores/permissions'
import { SportCategory, SPORT_CATEGORY_NAMES } from '@/types/admin'
import type { SportType, CreateSportTypeRequest, UpdateSportTypeRequest } from '@/types/admin'

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
const permissionStore = usePermissionStore()

// 表单引用
const formRef = ref<FormInstance>()

// 状态
const loading = ref(false)

// 计算属性
const visible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const isEdit = computed(() => !!props.sportType)

const dialogTitle = computed(() => {
  return isEdit.value ? '编辑运动类型' : '创建运动类型'
})

// 类别选项
const categoryOptions = computed(() => SPORT_CATEGORY_NAMES)

// 表单数据
const formData = ref({
  name: '',
  code: '',
  category: SportCategory.ESPORTS,
  icon: '',
  banner: '',
  description: '',
  sort_order: 0,
  configuration: {
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
})

// 表单验证规则
const formRules: FormRules = {
  name: [
    { required: true, message: '请输入运动名称', trigger: 'blur' },
    { min: 2, max: 100, message: '运动名称长度在 2 到 100 个字符', trigger: 'blur' }
  ],
  code: [
    { required: true, message: '请输入运动代码', trigger: 'blur' },
    { min: 2, max: 20, message: '运动代码长度在 2 到 20 个字符', trigger: 'blur' },
    { pattern: /^[a-zA-Z0-9_-]+$/, message: '运动代码只能包含字母、数字、下划线和横线', trigger: 'blur' }
  ],
  category: [
    { required: true, message: '请选择运动类别', trigger: 'change' }
  ]
}

// 监听弹窗显示状态
watch(visible, (newVisible) => {
  if (newVisible) {
    initFormData()
  }
})

// 初始化表单数据
const initFormData = () => {
  if (isEdit.value && props.sportType) {
    // 编辑模式，填充现有数据
    formData.value = {
      name: props.sportType.name,
      code: props.sportType.code,
      category: props.sportType.category,
      icon: props.sportType.icon || '',
      banner: props.sportType.banner || '',
      description: props.sportType.description || '',
      sort_order: props.sportType.sort_order,
      configuration: {
        enable_realtime: props.sportType.configuration?.enable_realtime ?? true,
        enable_chat: props.sportType.configuration?.enable_chat ?? false,
        enable_voting: props.sportType.configuration?.enable_voting ?? true,
        enable_prediction: props.sportType.configuration?.enable_prediction ?? true,
        enable_leaderboard: props.sportType.configuration?.enable_leaderboard ?? true,
        allow_modification: props.sportType.configuration?.allow_modification ?? true,
        max_modifications: props.sportType.configuration?.max_modifications ?? 3,
        modification_deadline: props.sportType.configuration?.modification_deadline ?? 30,
        enable_self_voting: props.sportType.configuration?.enable_self_voting ?? false,
        max_votes_per_user: props.sportType.configuration?.max_votes_per_user ?? 10,
        voting_deadline: props.sportType.configuration?.voting_deadline ?? 0
      }
    }
  } else {
    // 创建模式，使用默认值
    formData.value = {
      name: '',
      code: '',
      category: SportCategory.ESPORTS,
      icon: '',
      banner: '',
      description: '',
      sort_order: 0,
      configuration: {
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
    }
  }
  
  // 清除表单验证状态
  nextTick(() => {
    formRef.value?.clearValidate()
  })
}

// 应用预设配置
const applyPresetConfig = async () => {
  try {
    const result = await ElMessageBox.confirm(
      '应用预设配置将覆盖当前的功能配置，是否继续？',
      '确认操作',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    if (result === 'confirm') {
      // 根据运动类别应用预设配置
      if (formData.value.category === SportCategory.ESPORTS) {
        formData.value.configuration = {
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
        formData.value.configuration = {
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
      
      ElMessage.success('预设配置已应用')
    }
  } catch (error) {
    // 用户取消操作
  }
}

// 提交表单
const handleSubmit = async () => {
  if (!formRef.value) return
  
  try {
    // 验证表单
    await formRef.value.validate()
    
    loading.value = true
    
    if (isEdit.value && props.sportType) {
      // 更新运动类型
      const updateData: UpdateSportTypeRequest = {
        name: formData.value.name,
        icon: formData.value.icon || undefined,
        banner: formData.value.banner || undefined,
        description: formData.value.description || undefined,
        sort_order: formData.value.sort_order
      }
      
      await adminStore.updateSportType(props.sportType.id, updateData)
      
      // 更新配置
      await adminStore.updateSportConfiguration(props.sportType.id, formData.value.configuration)
      
      ElMessage.success('运动类型更新成功')
    } else {
      // 创建运动类型
      const createData: CreateSportTypeRequest = {
        name: formData.value.name,
        code: formData.value.code,
        category: formData.value.category,
        icon: formData.value.icon || undefined,
        banner: formData.value.banner || undefined,
        description: formData.value.description || undefined,
        sort_order: formData.value.sort_order,
        configuration: formData.value.configuration
      }
      
      await adminStore.createSportType(createData)
      
      ElMessage.success('运动类型创建成功')
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

.section-title {
  font-weight: 600;
  color: #303133;
}

.form-tip {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}

.config-group {
  margin-bottom: 20px;
}

.config-group h4 {
  margin: 0 0 12px 0;
  font-size: 14px;
  font-weight: 600;
  color: #303133;
}

.config-label {
  display: inline-block;
  width: 100px;
  font-size: 13px;
  color: #606266;
  margin-right: 8px;
}

.config-unit {
  margin-left: 8px;
  font-size: 12px;
  color: #909399;
}

.config-tip {
  font-size: 11px;
  color: #C0C4CC;
  margin-top: 2px;
  line-height: 1.2;
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

:deep(.el-checkbox) {
  margin-bottom: 8px;
}

:deep(.el-input-number .el-input__inner) {
  text-align: left;
}
</style>