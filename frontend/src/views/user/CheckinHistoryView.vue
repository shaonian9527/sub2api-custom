<template>
  <AppLayout>
    <div class="mx-auto max-w-4xl space-y-6">
      <div class="card p-6">
        <div class="flex items-center justify-between gap-4">
          <div>
            <h1 class="text-xl font-semibold text-gray-900 dark:text-white">
              {{ t('checkinHistory.title') }}
            </h1>
            <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">
              {{ t('checkinHistory.description') }}
            </p>
          </div>
          <button class="btn btn-secondary" @click="loadHistory">
            {{ t('common.refresh') }}
          </button>
        </div>
      </div>

      <div class="card p-6">
        <div v-if="loading" class="flex items-center justify-center py-10 text-gray-500">
          {{ t('common.loading') }}
        </div>

        <div v-else-if="items.length === 0" class="py-10 text-center text-gray-500 dark:text-dark-400">
          {{ t('checkinHistory.empty') }}
        </div>

        <div v-else class="space-y-3">
          <div
            v-for="item in items"
            :key="item.id"
            class="rounded-xl border border-gray-100 p-4 dark:border-dark-700"
          >
            <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
              <div>
                <div class="flex items-center gap-2">
                  <span class="text-sm font-semibold text-gray-900 dark:text-white">
                    {{ item.day }}
                  </span>
                  <span class="rounded-full bg-primary-50 px-2 py-0.5 text-xs text-primary-700 dark:bg-primary-900/20 dark:text-primary-300">
                    {{ t('checkinHistory.streakDays', { days: item.consecutive_days }) }}
                  </span>
                </div>
                <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">
                  {{ t('checkinHistory.checkedInAt') }}: {{ formatDate(item.used_at, { year: 'numeric', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' }) }}
                </p>
              </div>
              <div class="text-right">
                <p class="text-sm font-semibold text-emerald-600 dark:text-emerald-400">
                  +{{ formatCurrency(item.reward_amount) }}
                </p>
                <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">
                  {{ t('checkinHistory.rewardBreakdown', { base: formatCurrency(item.base_reward_amount), bonus: formatCurrency(item.bonus_reward_amount) }) }}
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import { checkinAPI, type CheckinHistoryItem } from '@/api/checkin'
import { formatDate } from '@/utils/format'
import { useAppStore } from '@/stores'

const { t } = useI18n()
const appStore = useAppStore()
const loading = ref(false)
const items = ref<CheckinHistoryItem[]>([])

const formatCurrency = (v: number) => v.toFixed(2)

const loadHistory = async () => {
  loading.value = true
  try {
    const res = await checkinAPI.getCheckinHistory(1, 50)
    items.value = res.items
  } catch (error: any) {
    appStore.showError(error?.response?.data?.detail || t('checkinHistory.loadFailed'))
  } finally {
    loading.value = false
  }
}

onMounted(loadHistory)
</script>
