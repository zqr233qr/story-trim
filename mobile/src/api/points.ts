import { request } from './index'
import type { Response } from './index'

// 积分余额返回结构。
export interface PointsBalanceResponse {
  balance: number
}

// 积分流水项。
export interface PointsLedgerEntry {
  id: number
  change: number
  balance_after: number
  type: string
  reason: string
  ref_type?: string
  ref_id?: string
  extra?: Record<string, string>
  created_at: string
}

// 积分流水返回结构。
export interface PointsLedgerResponse {
  items: PointsLedgerEntry[]
}

export const pointsApi = {
  // 获取积分余额。
  getBalance: (): Promise<Response<PointsBalanceResponse>> => {
    return request({
      url: '/users/me/points',
      method: 'GET'
    })
  },

  // 获取积分流水。
  getLedger: (page = 1, size = 20): Promise<Response<PointsLedgerResponse>> => {
    return request({
      url: `/users/me/points/ledger?page=${page}&size=${size}`,
      method: 'GET'
    })
  }
}
