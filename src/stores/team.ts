import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Team, TeamPayload } from '@/api/team'
import { getTeams, createTeam, updateTeam, deleteTeam } from '@/api/team'

export const useTeamStore = defineStore('team', () => {
  const teams = ref<Team[]>([])
  const loading = ref(false)
  const loaded = ref(false)

  const logoMap = computed(() => {
    const map: Record<string, string> = {}
    teams.value.forEach(t => {
      if (t.logoUrl) {
        map[t.name.toLowerCase()] = t.logoUrl
        if (t.shortName) {
          map[t.shortName.toLowerCase()] = t.logoUrl
        }
      }
    })
    return map
  })

  const getLogo = (teamName?: string) => {
    if (!teamName) return ''
    return logoMap.value[teamName.toLowerCase()] || ''
  }

  const fetchTeams = async (force = false) => {
    if (loaded.value && !force) return
    loading.value = true
    try {
      teams.value = await getTeams(true)
      loaded.value = true
    } finally {
      loading.value = false
    }
  }

  const addTeam = async (payload: TeamPayload) => {
    const created = await createTeam(payload)
    teams.value.push(created)
    return created
  }

  const editTeam = async (id: number, payload: TeamPayload) => {
    const updated = await updateTeam(id, payload)
    const idx = teams.value.findIndex(t => t.id === id)
    if (idx !== -1) {
      teams.value[idx] = updated
    }
    return updated
  }

  const removeTeam = async (id: number) => {
    await deleteTeam(id)
    teams.value = teams.value.filter(t => t.id !== id)
  }

  const ensureLoaded = async () => {
    if (!loaded.value && !loading.value) {
      await fetchTeams()
    }
  }

  return {
    teams,
    loading,
    loaded,
    fetchTeams,
    ensureLoaded,
    addTeam,
    editTeam,
    removeTeam,
    getLogo,
  }
})

