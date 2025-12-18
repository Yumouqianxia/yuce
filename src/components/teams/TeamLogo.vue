<template>
  <div class="team-logo" :class="size">
    <img v-if="getTeamLogo()" :src="getTeamLogo()" :alt="teamName" class="team-logo-img">
    <span v-else class="team-initial">{{ getInitial() }}</span>
  </div>
</template>

<script setup>
import { computed, onMounted } from 'vue';
import { useTeamStore } from '@/stores/team';

// 直接导入所有图片
import AG from '../../assets/teams/AG.jpg';
import DRG from '../../assets/teams/DRG.jpg';
import DYG from '../../assets/teams/DYG.jpg';
import EDG from '../../assets/teams/EDG.jpg';
import ESTAR from '../../assets/teams/ESTAR.jpg';
import HERO from '../../assets/teams/HERO.jpg';
import JDG from '../../assets/teams/JDG.jpg';
import KSG from '../../assets/teams/KSG.jpg';
import LGD from '../../assets/teams/LGD.jpg';
import RNG from '../../assets/teams/RNG.jpg';
import RW from '../../assets/teams/RW.jpg';
import TESA from '../../assets/teams/TES.A.jpg';
import TTG from '../../assets/teams/TTG.jpg';
import WB from '../../assets/teams/WB.jpg';
import WE from '../../assets/teams/WE.jpg';
import WOLVES from '../../assets/teams/WOLVES.jpg';

const props = defineProps({
  teamName: {
    type: String,
    required: true
  },
  size: {
    type: String,
    default: 'medium' // small, medium, large
  },
  logoUrl: {
    type: String,
    default: ''
  }
});

const teamStore = useTeamStore();
onMounted(() => {
  teamStore.ensureLoaded();
});



// 获取队伍首字母
function getInitial() {
  return props.teamName ? props.teamName.charAt(0).toUpperCase() : 'T';
}

// 直接获取队伍图标
function getTeamLogo() {
  // 优先使用传入或存储的图标
  const dynamicLogo = props.logoUrl || teamStore.getLogo(props.teamName);
  if (dynamicLogo) return dynamicLogo;

  // 如果没有队伍名称，返回AG图标作为默认
  if (!props.teamName) return AG;

  // 直接写出所有战队的图标路径
  switch (props.teamName) {
    case 'AG': return AG;
    case 'DRG': return DRG;
    case 'DYG': return DYG;
    case 'EDG': return EDG;
    case 'ESTAR': return ESTAR;
    case 'HERO': return HERO;
    case 'JDG': return JDG;
    case 'KSG': return KSG;
    case 'LGD': return LGD;
    case 'RNG': return RNG;
    case 'RW': return RW;
    case 'TES.A': return TESA;
    case 'TTG': return TTG;
    case 'WB': return WB;
    case 'WE': return WE;
    case '狼队': return WOLVES;
    // 全称匹配
    case '成都AG超玩会': return AG;
    case '佛山DRG': return DRG;
    case '深圳DYG': return DYG;
    case '上海EDG.M': return EDG;
    case '武汉eStarPro': return ESTAR;
    case 'Hero久竞': return HERO;
    case '北京JDG': return JDG;
    case '苏州KSG': return KSG;
    case '杭州NBW': return LGD;
    case '上海RNG.M': return RNG;
    case '济南RW侠': return RW;
    case '长沙TES.A': return TESA;
    case '广州TTG': return TTG;
    case '北京WB': return WB;
    case '西安WE': return WE;
    case '重庆狼队': return WOLVES;
    // 对于没有匹配的队伍，返回默认图标
    default:
      console.log('未匹配到队伍图标:', props.teamName);
      // 如果是狼队，返回狼队图标
      if (props.teamName.includes('狼队')) return WOLVES;
      return null;
  }
}
</script>

<style scoped>
.team-logo {
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  border-radius: 50%;
  background: linear-gradient(135deg, #60a5fa, #3b82f6);
}

.team-logo.small {
  width: 32px;
  height: 32px;
}

.team-logo.medium {
  width: 48px;
  height: 48px;
}

.team-logo.large {
  width: 64px;
  height: 64px;
}

.team-logo-img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.team-initial {
  font-weight: bold;
  color: white;
  font-size: 22px;
}

.team-logo.small .team-initial {
  font-size: 16px;
}

.team-logo.medium .team-initial {
  font-size: 22px;
}

.team-logo.large .team-initial {
  font-size: 28px;
}
</style>
