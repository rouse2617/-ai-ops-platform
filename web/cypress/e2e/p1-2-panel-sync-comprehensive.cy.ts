import { ChatPage } from '../support/page-objects'

describe('P1-2: Panel Synchronization - Comprehensive', () => {
  let chatPage: ChatPage

  beforeEach(() => {
    chatPage = new ChatPage()
    chatPage.visit()
  })

  describe('TC01: System load message syncs to right panel', () => {
    it('should display CPU/Memory charts when sending system load query', () => {
      chatPage.selectHosts(['host-001'])
      chatPage.sendMessage('查看系统负载')

      chatPage.getRightPanel().should('be.visible')
      cy.get('[data-testid="cpu-chart"]').should('exist')
      cy.get('[data-testid="memory-chart"]').should('exist')
    })

    it('should update charts with real-time data', () => {
      chatPage.selectHosts(['host-001'])
      chatPage.sendMessage('查看系统负载')

      cy.get('[data-testid="cpu-value"]').should('contain', '%')
      cy.get('[data-testid="memory-value"]').should('contain', '%')
    })

    it('should display host name in monitoring panel', () => {
      chatPage.selectHosts(['host-001'])
      chatPage.sendMessage('查看系统负载')

      cy.get('[data-testid="monitor-host-name"]').should('contain', 'host-001')
    })
  })

  describe('TC02: Log analysis message syncs to right panel', () => {
    it('should display log analysis panel when sending log query', () => {
      chatPage.selectHosts(['host-001'])
      chatPage.sendMessage('分析系统日志')

      chatPage.getRightPanel().should('be.visible')
      cy.get('[data-testid="log-analysis-panel"]').should('exist')
    })

    it('should show log entries in analysis panel', () => {
      chatPage.selectHosts(['host-001'])
      chatPage.sendMessage('分析系统日志')

      cy.get('[data-testid="log-entry"]').should('have.length.greaterThan', 0)
    })

    it('should display log severity indicators', () => {
      chatPage.selectHosts(['host-001'])
      chatPage.sendMessage('分析系统日志')

      cy.get('[data-testid="log-severity"]').should('exist')
    })
  })

  describe('TC03: Process monitoring message syncs to right panel', () => {
    it('should display process list when sending process query', () => {
      chatPage.selectHosts(['host-001'])
      chatPage.sendMessage('查看进程列表')

      chatPage.getRightPanel().should('be.visible')
      cy.get('[data-testid="process-list"]').should('exist')
    })

    it('should show process details in table', () => {
      chatPage.selectHosts(['host-001'])
      chatPage.sendMessage('查看进程列表')

      cy.get('[data-testid="process-row"]').should('have.length.greaterThan', 0)
      cy.get('[data-testid="process-name"]').should('exist')
      cy.get('[data-testid="process-pid"]').should('exist')
    })

    it('should display CPU and memory usage per process', () => {
      chatPage.selectHosts(['host-001'])
      chatPage.sendMessage('查看进程列表')

      cy.get('[data-testid="process-cpu"]').should('exist')
      cy.get('[data-testid="process-memory"]').should('exist')
    })
  })

  describe('TC04: Network status message syncs to right panel', () => {
    it('should display network status when sending network query', () => {
      chatPage.selectHosts(['host-001'])
      chatPage.sendMessage('查看网络状态')

      chatPage.getRightPanel().should('be.visible')
      cy.get('[data-testid="network-status"]').should('exist')
    })

    it('should show network interfaces', () => {
      chatPage.selectHosts(['host-001'])
      chatPage.sendMessage('查看网络状态')

      cy.get('[data-testid="network-interface"]').should('have.length.greaterThan', 0)
    })

    it('should display network statistics', () => {
      chatPage.selectHosts(['host-001'])
      chatPage.sendMessage('查看网络状态')

      cy.get('[data-testid="network-rx"]').should('exist')
      cy.get('[data-testid="network-tx"]').should('exist')
    })
  })

  describe('Panel switching with different message types', () => {
    it('should switch panels when sending different message types', () => {
      chatPage.selectHosts(['host-001'])

      chatPage.sendMessage('查看系统负载')
      cy.get('[data-testid="cpu-chart"]').should('exist')

      chatPage.sendMessage('查看进程列表')
      cy.get('[data-testid="process-list"]').should('exist')

      chatPage.sendMessage('查看网络状态')
      cy.get('[data-testid="network-status"]').should('exist')
    })
  })

  describe('Multiple hosts panel synchronization', () => {
    it('should display all selected hosts in monitoring panel', () => {
      chatPage.selectHosts(['host-001', 'host-002', 'host-003'])
      chatPage.sendMessage('查看系统负载')

      cy.get('[data-testid="monitor-host-section"]').should('have.length', 3)
    })

    it('should update panels when host selection changes', () => {
      chatPage.selectHosts(['host-001'])
      chatPage.sendMessage('查看系统负载')

      cy.get('[data-testid="monitor-host-section"]').should('have.length', 1)

      chatPage.selectHosts(['host-001', 'host-002'])
      cy.get('[data-testid="monitor-host-section"]').should('have.length', 2)
    })
  })

  describe('Panel collapse and expand', () => {
    it('should allow collapsing right panel', () => {
      chatPage.selectHosts(['host-001'])
      chatPage.sendMessage('查看系统负载')

      chatPage.getRightPanel().should('be.visible')
      cy.get('[data-testid="panel-toggle"]').click()
      chatPage.getRightPanel().should('not.be.visible')
    })

    it('should restore panel when expanding', () => {
      chatPage.selectHosts(['host-001'])
      chatPage.sendMessage('查看系统负载')

      cy.get('[data-testid="panel-toggle"]').click()
      chatPage.getRightPanel().should('not.be.visible')

      cy.get('[data-testid="panel-toggle"]').click()
      chatPage.getRightPanel().should('be.visible')
    })
  })
})
