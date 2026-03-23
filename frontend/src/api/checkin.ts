import { apiClient } from './client'
import type { PaginatedResponse } from '@/types'

export interface CheckinStatus {
  enabled: boolean
  checked_in: boolean
  reward_amount: number
  base_reward_amount: number
  consecutive_days: number
  next_reward_amount: number
  timezone: string
  today: string
  checked_in_at?: string
  balance_after?: number
  checkin_code?: string
  history_page_path?: string
}

export interface CheckinResult {
  message: string
  reward_amount: number
  base_reward_amount: number
  bonus_reward_amount: number
  new_balance: number
  today: string
  timezone: string
  consecutive_days: number
  checked_in_at: string
  code: string
}

export interface CheckinHistoryItem {
  id: number
  code: string
  reward_amount: number
  base_reward_amount: number
  bonus_reward_amount: number
  consecutive_days: number
  day: string
  used_at: string
}

export async function getCheckinStatus(): Promise<CheckinStatus> {
  const { data } = await apiClient.get<CheckinStatus>('/user/checkin/status')
  return data
}

export async function performCheckin(): Promise<CheckinResult> {
  const { data } = await apiClient.post<CheckinResult>('/user/checkin', {})
  return data
}

export async function getCheckinHistory(page = 1, pageSize = 20): Promise<PaginatedResponse<CheckinHistoryItem>> {
  const { data } = await apiClient.get<PaginatedResponse<CheckinHistoryItem>>('/user/checkin/history', {
    params: { page, page_size: pageSize }
  })
  return data
}

export const checkinAPI = {
  getCheckinStatus,
  performCheckin,
  getCheckinHistory,
}

export default checkinAPI
