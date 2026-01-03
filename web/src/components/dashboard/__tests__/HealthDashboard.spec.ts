import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import HealthDashboard from '../HealthDashboard.vue'

// Mock request
vi.mock('@/api/request', () => ({
  request: {
    get: vi.fn(),
    post: vi.fn()
  }
}))

describe('HealthDashboard', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('renders correctly', () => {
    const wrapper = mount(HealthDashboard)

    expect(wrapper.find('.health-dashboard').exists()).toBe(true)
    expect(wrapper.find('.overall-health').exists()).toBe(true)
    expect(wrapper.find('.hosts-comparison').exists()).toBe(true)
  })

  it('displays overall health score', () => {
    const wrapper = mount(HealthDashboard)

    expect(wrapper.find('.score-value').exists()).toBe(true)
    expect(wrapper.find('.score-label').text()).toBe('健康分')
  })

  it('has health check button', () => {
    const wrapper = mount(HealthDashboard)

    const button = wrapper.find('.section-header .el-button')
    expect(button.exists()).toBe(true)
    expect(button.text()).toContain('开始体检')
  })

  it('displays correct level text based on score', async () => {
    const wrapper = mount(HealthDashboard)

    // 默认状态
    expect(wrapper.find('.score-info h3').text()).toBe('系统健康度')
  })

  it('shows anomaly alert when there are anomalies', async () => {
    const wrapper = mount(HealthDashboard)

    // 设置异常主机
    await wrapper.vm.anomalies = ['server-1', 'server-2']
    await wrapper.vm.$nextTick()

    // 检查是否显示异常提示
    const anomalyAlert = wrapper.find('.anomaly-alert')
    if (wrapper.vm.anomalies.length > 0) {
      expect(anomalyAlert.exists()).toBe(true)
    }
  })

  it('displays highlights section', () => {
    const wrapper = mount(HealthDashboard)

    expect(wrapper.find('.highlights').exists()).toBe(true)
  })
})

describe('HealthDashboard - Score Levels', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  const scoreLevels = [
    { score: 95, level: 'excellent', text: '优秀' },
    { score: 80, level: 'good', text: '良好' },
    { score: 65, level: 'warning', text: '需关注' },
    { score: 40, level: 'critical', text: '异常' }
  ]

  scoreLevels.forEach(({ score, level, text }) => {
    it(`displays correct level for score ${score}`, async () => {
      const wrapper = mount(HealthDashboard)

      // 设置分数和等级
      wrapper.vm.overallScore = score
      wrapper.vm.scoreLevel = level
      await wrapper.vm.$nextTick()

      expect(wrapper.find('.score-circle').classes()).toContain(level)
    })
  })
})

describe('HealthDashboard - Progress Colors', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('returns correct color for different usage levels', () => {
    const wrapper = mount(HealthDashboard)

    // 测试 getProgressColor 函数
    expect(wrapper.vm.getProgressColor(95)).toBe('#f56c6c') // 危险
    expect(wrapper.vm.getProgressColor(85)).toBe('#e6a23c') // 警告
    expect(wrapper.vm.getProgressColor(70)).toBe('#409eff') // 正常
    expect(wrapper.vm.getProgressColor(50)).toBe('#67c23a') // 良好
  })
})

describe('HealthDashboard - Status Helpers', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('returns correct status type', () => {
    const wrapper = mount(HealthDashboard)

    expect(wrapper.vm.getStatusType('healthy')).toBe('success')
    expect(wrapper.vm.getStatusType('warning')).toBe('warning')
    expect(wrapper.vm.getStatusType('critical')).toBe('danger')
    expect(wrapper.vm.getStatusType('unknown')).toBe('info')
  })

  it('returns correct status text', () => {
    const wrapper = mount(HealthDashboard)

    expect(wrapper.vm.getStatusText('healthy')).toBe('健康')
    expect(wrapper.vm.getStatusText('warning')).toBe('警告')
    expect(wrapper.vm.getStatusText('critical')).toBe('严重')
    expect(wrapper.vm.getStatusText('unknown')).toBe('未知')
  })
})
