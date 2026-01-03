describe('P1-2: Panel Synchronization', () => {
  beforeEach(() => {
    cy.visit('/')
  })

  describe('TC01: System load message syncs to right panel', () => {
    it('should display CPU/Memory charts when sending system load query', () => {
      cy.selectHosts(['host-001'])
      cy.sendMessage('查看系统负载')

      cy.get('[data-testid="right-panel"]').should('be.visible')
      cy.get('[data-testid="cpu-chart"]').should('exist')
      cy.get('[data-testid="memory-chart"]').should('exist')
    })
  })

  describe('TC02: Log analysis message syncs to right panel', () => {
    it('should display log analysis panel when sending log query', () => {
      cy.selectHosts(['host-001'])
      cy.sendMessage('分析系统日志')

      cy.get('[data-testid="right-panel"]').should('be.visible')
      cy.get('[data-testid="log-analysis-panel"]').should('exist')
    })
  })

  describe('TC03: Process monitoring message syncs to right panel', () => {
    it('should display process list when sending process query', () => {
      cy.selectHosts(['host-001'])
      cy.sendMessage('查看进程列表')

      cy.get('[data-testid="right-panel"]').should('be.visible')
      cy.get('[data-testid="process-list"]').should('exist')
    })
  })

  describe('TC04: Network status message syncs to right panel', () => {
    it('should display network status when sending network query', () => {
      cy.selectHosts(['host-001'])
      cy.sendMessage('查看网络状态')

      cy.get('[data-testid="right-panel"]').should('be.visible')
      cy.get('[data-testid="network-status"]').should('exist')
    })
  })
})
