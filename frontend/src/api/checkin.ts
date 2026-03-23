import { apiClient } from './client'

export interface CheckinStatus {
  checked_in: boolean
  reward_amount: number
  today: string
  checked_in_at?: string
  balance_after?: number
  checkin_code?: string
}

export interface CheckinResult {
  message: string
  reward_amount: number
  new_balance: number
  today: string
  checked_in_at: string
  code: string
}

export async function getCheckinStatus(): Promise<CheckinStatus> {
  const { data } = await apiClient.get<CheckinStatus>('/user/checkin/status')
  return data
}

export async function performCheckin(): Promise<CheckinResult> {
  const { data } = await apiClient.post<CheckinResult>('/user/checkin', {})
  return data
}

export const checkinAPI = {
  getCheckinStatus,
  performCheckin,
}

export default checkinAPI
