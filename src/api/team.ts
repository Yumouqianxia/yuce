import { get, post, put, del } from './http'

export interface Team {
  id: number
  name: string
  shortName?: string
  logoUrl?: string
  isActive: boolean
}

export interface TeamPayload {
  name: string
  shortName?: string
  logoUrl?: string
  isActive: boolean
}

export const getTeams = (includeInactive = false) =>
  get<Team[]>('/api/teams', includeInactive ? { all: 1 } : undefined)

export const createTeam = (payload: TeamPayload) =>
  post<Team>('/api/teams', payload)

export const updateTeam = (id: number, payload: TeamPayload) =>
  put<Team>(`/api/teams/${id}`, payload)

export const deleteTeam = (id: number) => del(`/api/teams/${id}`)

