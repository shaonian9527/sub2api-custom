<template>
  <AppLayout>
    <div class="mx-auto max-w-4xl space-y-6">
      <div class="grid grid-cols-1 gap-6 sm:grid-cols-3">
        <StatCard
          :title="t('profile.accountBalance')"
          :value="formatCurrency(user?.balance || 0)"
          :icon="WalletIcon"
          icon-variant="success"
        />
        <StatCard
          :title="t('profile.concurrencyLimit')"
          :value="user?.concurrency || 0"
          :icon="BoltIcon"
          icon-variant="warning"
        />
        <StatCard
          :title="t('profile.memberSince')"
          :value="formatDate(user?.created_at || '', { year: 'numeric', month: 'long' })"
          :icon="CalendarIcon"
          icon-variant="primary"
        />
      </div>

      <div class="card p-6" v-if="checkinStatus?.enabled !== false">
        <div class="flex flex-col gap-4">
          <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
            <div>
              <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
                {{ t('profile.dailyCheckin') }}
              </h3>
              <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">
                {{
                  checkinStatus?.checked_in
                    ? t('profile.checkedInToday')
                    : t('profile.checkinHint', { amount: checkinRewardText })
                }}
              </p>
              <p class="mt-2 text-xs text-gray-400 dark:text-dark-500">
                {{ t('profile.checkinTimezone') }}: {{ checkinStatus?.timezone || 'Asia/Shanghai' }}
              </p>
              <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">
                {{ t('profile.currentStreak', { days: checkinStatus?.consecutive_days || 0 }) }}
              </p>
              <p
                v-if="checkinStatus?.checked_in_at"
                class="mt-1 text-xs text-gray-400 dark:text-dark-500"
              >
                {{ t('profile.checkedInAt') }}:
                {{
                  formatDate(checkinStatus.checked_in_at, {
                    year: 'numeric',
                    month: 'short',
                    day: 'numeric',
                    hour: '2-digit',
                    minute: '2-digit'
                  })
                }}
              </p>
            </div>
            <button
              class="btn"
              :class="checkinStatus?.checked_in ? 'btn-secondary' : 'btn-primary'"
              :disabled="checkingIn || !!checkinStatus?.checked_in"
              @click="handleCheckin"
            >
              {{
                checkingIn
                  ? t('profile.checkingIn')
                  : checkinStatus?.checked_in
                    ? t('profile.alreadyCheckedIn')
                    : t('profile.checkinNow')
              }}
            </button>
          </div>

          <div class="flex flex-wrap items-center gap-3 text-xs text-gray-500 dark:text-dark-400">
            <span>{{ t('profile.baseReward', { amount: formatCurrency(checkinStatus?.base_reward_amount || 0) }) }}</span>
            <span>{{ t('profile.nextReward', { amount: formatCurrency(checkinStatus?.next_reward_amount || 0) }) }}</span>
            <router-link to="/checkin-history" class="text-primary-600 hover:underline dark:text-primary-400">
              {{ t('profile.viewCheckinHistory') }}
            </router-link>
          </div>
        </div>
      </div>

      <ProfileInfoCard :user="user" />

      <div v-if="contactInfo" class="card border-primary-200 bg-primary-50 p-6 dark:bg-primary-900/20">
        <div class="flex items-center gap-4">
          <div class="rounded-xl bg-primary-100 p-3 text-primary-600">
            <Icon name="chat" size="lg" />
          </div>
          <div>
            <h3 class="font-semibold text-primary-800 dark:text-primary-200">
              {{ t('common.contactSupport') }}
            </h3>
            <p class="text-sm font-medium">{{ contactInfo }}</p>
          </div>
        </div>
      </div>

      <ProfileEditForm :initial-username="user?.username || ''" />
      <ProfilePasswordForm />
      <ProfileTotpCard />
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, h, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { formatDate } from '@/utils/format'
import { authAPI, checkinAPI } from '@/api'
import AppLayout from '@/components/layout/AppLayout.vue'
import StatCard from '@/components/common/StatCard.vue'
import ProfileInfoCard from '@/components/user/profile/ProfileInfoCard.vue'
import ProfileEditForm from '@/components/user/profile/ProfileEditForm.vue'
import ProfilePasswordForm from '@/components/user/profile/ProfilePasswordForm.vue'
import ProfileTotpCard from '@/components/user/profile/ProfileTotpCard.vue'
import { Icon } from '@/components/icons'

const { t } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()
const user = computed(() => authStore.user)
const contactInfo = ref('')
const checkinStatus = ref<{
  enabled: boolean
  checked_in: boolean
  reward_amount: number
  base_reward_amount: number
  consecutive_days: number
  next_reward_amount: number
  timezone: string
  today: string
  checked_in_at?: string
} | null>(null)
const checkingIn = ref(false)

const WalletIcon = {
  render: () =>
    h(
      'svg',
      { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.5' },
      [h('path', { d: 'M21 12a2.25 2.25 0 00-2.25-2.25H15a3 3 0 11-6 0H5.25A2.25 2.25 0 003 12' })]
    )
}
const BoltIcon = {
  render: () =>
    h(
      'svg',
      { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.5' },
      [h('path', { d: 'm3.75 13.5 10.5-11.25L12 10.5h8.25L9.75 21.75 12 13.5H3.75z' })]
    )
}
const CalendarIcon = {
  render: () =>
    h(
      'svg',
      { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.5' },
      [h('path', { d: 'M6.75 3v2.25M17.25 3v2.25' })]
    )
}

const formatCurrency = (v: number) => `$${v.toFixed(2)}`
const checkinRewardText = computed(() => formatCurrency(checkinStatus.value?.reward_amount || 1))

const loadCheckinStatus = async () => {
  try {
    checkinStatus.value = await checkinAPI.getCheckinStatus()
  } catch (error) {
    console.error('Failed to load checkin status:', error)
  }
}

const handleCheckin = async () => {
  checkingIn.value = true
  try {
    const result = await checkinAPI.performCheckin()
    await authStore.refreshUser()
    await loadCheckinStatus()
    appStore.showSuccess(t('profile.checkinSuccess', { amount: formatCurrency(result.reward_amount) }))
  } catch (error: any) {
    appStore.showError(error?.response?.data?.message || error?.response?.data?.detail || t('profile.checkinFailed'))
  } finally {
    checkingIn.value = false
  }
}

onMounted(async () => {
  try {
    const s = await authAPI.getPublicSettings()
    contactInfo.value = s.contact_info || ''
  } catch (error) {
    console.error('Failed to load contact info:', error)
  }

  await loadCheckinStatus()
})
</script>
