describe('P1-3: Long Content Collapsing', () => {
  beforeEach(() => {
    cy.visit('/')
  })

  describe('TC01: Output over 50 lines shows collapse', () => {
    it('should display collapse button for output with >50 lines', () => {
      cy.selectHosts(['host-001'])
      cy.sendMessage('查看完整日志')

      cy.get('[data-testid="message-list"]').within(() => {
        cy.get('[data-testid="collapsible-output"]').should('exist')
        cy.get('[data-testid="collapse-toggle"]').should('contain', '展开全部')
      })
    })
  })

  describe('TC02: Expand button shows full content', () => {
    it('should display full content when clicking expand', () => {
      cy.selectHosts(['host-001'])
      cy.sendMessage('查看完整日志')

      cy.get('[data-testid="message-list"]').within(() => {
        cy.get('[data-testid="collapse-toggle"]').click()
        cy.get('[data-testid="output-content"]').should('contain', '... 省略')
        cy.get('[data-testid="collapse-toggle"]').should('contain', '收起')
      })
    })
  })

  describe('TC03: Collapse button restores collapsed state', () => {
    it('should restore collapsed state when clicking collapse', () => {
      cy.selectHosts(['host-001'])
      cy.sendMessage('查看完整日志')

      cy.get('[data-testid="message-list"]').within(() => {
        cy.get('[data-testid="collapse-toggle"]').click()
        cy.get('[data-testid="output-content"]').should('not.contain', '... 省略')

        cy.get('[data-testid="collapse-toggle"]').click()
        cy.get('[data-testid="output-content"]').should('contain', '... 省略')
        cy.get('[data-testid="collapse-toggle"]').should('contain', '展开全部')
      })
    })
  })

  describe('TC04: Short content does not show collapse button', () => {
    it('should not display collapse button for output with <50 lines', () => {
      cy.selectHosts(['host-001'])
      cy.sendMessage('查看系统负载')

      cy.get('[data-testid="message-list"]').within(() => {
        cy.get('[data-testid="collapsible-output"]').should('exist')
        cy.get('[data-testid="collapse-toggle"]').should('not.exist')
      })
    })
  })
})
