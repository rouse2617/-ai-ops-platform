import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import ProactiveMessageCard from '../ProactiveMessageCard.vue'

describe('ProactiveMessageCard', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  const defaultProps = {
    type: 'morning_report',
    title: '早安报告',
    content: '系统运行正常',
    priority: 'low',
    timestamp: Math.floor(Date.now() / 1000)
  }

  it('renders correctly with default props', () => {
    const wrapper = mount(ProactiveMessageCard, {
      props: defaultProps
    })

    expect(wrapper.find('.proactive-message').exists()).toBe(true)
    expect(wrapper.find('.title-text').text()).toBe('早安报告')
    expect(wrapper.find('.monitoring-badge').text()).toContain('监控中')
  })

  it('displays correct priority tag', () => {
    const priorities = [
      { priority: 'critical', label: '紧急' },
      { priority: 'high', label: '重要' },
      { priority: 'medium', label: '一般' },
      { priority: 'low', label: '低' }
    ]

    priorities.forEach(({ priority, label }) => {
      const wrapper = mount(ProactiveMessageCard, {
        props: { ...defaultProps, priority }
      })
      expect(wrapper.find('.el-tag').text()).toBe(label)
    })
  })

  it('applies correct CSS class based on priority', () => {
    const wrapper = mount(ProactiveMessageCard, {
      props: { ...defaultProps, priority: 'critical' }
    })

    expect(wrapper.find('.proactive-message').classes()).toContain('priority-critical')
  })

  it('applies correct CSS class based on type', () => {
    const wrapper = mount(ProactiveMessageCard, {
      props: { ...defaultProps, type: 'morning_report' }
    })

    expect(wrapper.find('.proactive-message').classes()).toContain('type-morning_report')
  })

  it('renders hosts list when provided', () => {
    const wrapper = mount(ProactiveMessageCard, {
      props: {
        ...defaultProps,
        hosts: ['server-1', 'server-2', 'server-3']
      }
    })

    const hostTags = wrapper.findAll('.hosts-list .el-tag')
    expect(hostTags.length).toBe(3)
    expect(hostTags[0].text()).toBe('server-1')
  })

  it('renders metrics when provided', () => {
    const wrapper = mount(ProactiveMessageCard, {
      props: {
        ...defaultProps,
        metrics: {
          cpu: 45.5,
          memory: 60.2,
          disk: 75.0
        }
      }
    })

    expect(wrapper.find('.metrics-summary').exists()).toBe(true)
    const metricItems = wrapper.findAll('.metric-item')
    expect(metricItems.length).toBe(3)
  })

  it('renders action buttons when provided', () => {
    const wrapper = mount(ProactiveMessageCard, {
      props: {
        ...defaultProps,
        actions: [
          { id: '1', label: '查看详情' },
          { id: '2', label: '运行诊断', dangerous: true }
        ]
      }
    })

    const buttons = wrapper.findAll('.proactive-actions .el-button')
    expect(buttons.length).toBe(2)
  })

  it('emits action event when button clicked', async () => {
    const action = { id: '1', label: '查看详情' }
    const wrapper = mount(ProactiveMessageCard, {
      props: {
        ...defaultProps,
        actions: [action]
      }
    })

    await wrapper.find('.proactive-actions .el-button').trigger('click')
    expect(wrapper.emitted('action')).toBeTruthy()
    expect(wrapper.emitted('action')[0]).toEqual([action])
  })

  it('renders markdown content correctly', () => {
    const wrapper = mount(ProactiveMessageCard, {
      props: {
        ...defaultProps,
        content: '**加粗文本** 和 *斜体文本*'
      }
    })

    const contentHtml = wrapper.find('.content-text').html()
    expect(contentHtml).toContain('<strong>')
    expect(contentHtml).toContain('<em>')
  })

  it('formats timestamp correctly', () => {
    const timestamp = Math.floor(new Date('2024-01-15 10:30:00').getTime() / 1000)
    const wrapper = mount(ProactiveMessageCard, {
      props: {
        ...defaultProps,
        timestamp
      }
    })

    expect(wrapper.find('.timestamp').text()).toContain('01')
    expect(wrapper.find('.timestamp').text()).toContain('15')
  })
})

describe('ProactiveMessageCard - Message Types', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  const messageTypes = [
    { type: 'morning_report', expectedClass: 'type-morning_report' },
    { type: 'anomaly_alert', expectedClass: 'type-anomaly_alert' },
    { type: 'health_summary', expectedClass: 'type-health_summary' },
    { type: 'auto_fix', expectedClass: 'type-auto_fix' }
  ]

  messageTypes.forEach(({ type, expectedClass }) => {
    it(`renders ${type} message type correctly`, () => {
      const wrapper = mount(ProactiveMessageCard, {
        props: {
          type,
          title: 'Test Title',
          content: 'Test Content',
          priority: 'medium'
        }
      })

      expect(wrapper.find('.proactive-message').classes()).toContain(expectedClass)
    })
  })
})
